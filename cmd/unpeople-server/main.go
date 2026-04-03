// Package main provides an HTTP server for the unpeople mesh generator.
//
// The server exposes a REST API for generating humanoid character meshes,
// enabling integration with non-Go game engines like Unity, Unreal, and Godot.
//
// Usage:
//
//	unpeople-server                    # Start on default port 8080
//	unpeople-server -addr :9000        # Start on port 9000
//	unpeople-server -addr 0.0.0.0:8080 # Bind to all interfaces
//
// Endpoints:
//
//	GET  /health   - Health check
//	POST /generate - Generate mesh from JSON parameters
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/opd-ai/unpeople"
)

var addrFlag = flag.String("addr", ":8080", "Server address (host:port)")

func main() {
	flag.Parse()

	srv := NewServer()
	log.Printf("Starting unpeople server on %s", *addrFlag)
	if err := http.ListenAndServe(*addrFlag, srv); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

// Server handles HTTP requests for mesh generation.
type Server struct {
	gen         *unpeople.Generator
	mux         *http.ServeMux
	rateLimiter *RateLimiter
}

// NewServer creates a new mesh generation server.
func NewServer() *Server {
	s := &Server{
		gen:         unpeople.NewGenerator(),
		mux:         http.NewServeMux(),
		rateLimiter: NewRateLimiter(100, time.Second), // 100 req/sec
	}
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/generate", s.handleGenerate)
	return s
}

// ServeHTTP implements the http.Handler interface and routes requests
// through the server's multiplexer with CORS headers for browser clients.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers for browser-based clients
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	s.mux.ServeHTTP(w, r)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

// resolveParams applies defaults if needed and validates parameters.
func resolveParams(params *unpeople.Params) error {
	if params.Seed == 0 {
		*params = unpeople.DefaultParams()
	}
	return params.Validate()
}

// writeOBJResponse writes mesh in OBJ format with appropriate headers.
func writeOBJResponse(w http.ResponseWriter, mesh *unpeople.Mesh) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=character.obj")
	unpeople.ExportOBJ(w, mesh, "character")
}

// writeGLTFResponse writes mesh in glTF JSON format.
func writeGLTFResponse(w http.ResponseWriter, mesh *unpeople.Mesh) {
	w.Header().Set("Content-Type", "model/gltf+json")
	unpeople.ExportGLTFDefault(w, mesh)
}

// writeGLBResponse writes mesh in glTF Binary format.
func writeGLBResponse(w http.ResponseWriter, mesh *unpeople.Mesh) {
	w.Header().Set("Content-Type", "model/gltf-binary")
	w.Header().Set("Content-Disposition", "attachment; filename=character.glb")
	unpeople.ExportGLB(w, mesh, unpeople.DefaultGLTFOptions())
}

// writeMeshResponse writes the mesh in the requested format.
func writeMeshResponse(w http.ResponseWriter, mesh *unpeople.Mesh, format string) {
	switch format {
	case "obj":
		writeOBJResponse(w, mesh)
	case "glb":
		writeGLBResponse(w, mesh)
	default:
		writeGLTFResponse(w, mesh)
	}
}

func (s *Server) handleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !s.rateLimiter.Allow() {
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	var params unpeople.Params
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	if err := resolveParams(&params); err != nil {
		http.Error(w, fmt.Sprintf("Invalid parameters: %v", err), http.StatusBadRequest)
		return
	}

	mesh, err := s.gen.Generate(params)
	if err != nil {
		http.Error(w, fmt.Sprintf("Generation failed: %v", err), http.StatusInternalServerError)
		return
	}

	writeMeshResponse(w, mesh, s.negotiateFormat(r))
}

func (s *Server) negotiateFormat(r *http.Request) string {
	// Check query parameter first
	if format := r.URL.Query().Get("format"); format != "" {
		return format
	}

	// Check Accept header
	accept := r.Header.Get("Accept")
	if accept == "" {
		return "gltf"
	}

	accept = strings.ToLower(accept)
	switch {
	case strings.Contains(accept, "model/gltf-binary"):
		return "glb"
	case strings.Contains(accept, "model/gltf+json"):
		return "gltf"
	case strings.Contains(accept, "text/plain"):
		return "obj"
	default:
		return "gltf"
	}
}

// RateLimiter implements a simple token bucket rate limiter.
type RateLimiter struct {
	tokens     int
	maxTokens  int
	refillRate time.Duration
	lastRefill time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a rate limiter with the given capacity and refill interval.
func NewRateLimiter(maxTokens int, refillRate time.Duration) *RateLimiter {
	return &RateLimiter{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Allow checks if a request should be allowed.
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Refill tokens based on time elapsed
	now := time.Now()
	elapsed := now.Sub(rl.lastRefill)
	if elapsed >= rl.refillRate {
		refills := int(elapsed / rl.refillRate)
		rl.tokens += refills
		if rl.tokens > rl.maxTokens {
			rl.tokens = rl.maxTokens
		}
		rl.lastRefill = now
	}

	if rl.tokens > 0 {
		rl.tokens--
		return true
	}
	return false
}
