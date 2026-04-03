// Package main provides tests for the unpeople-server.
package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/opd-ai/unpeople"
)

func TestHealthEndpoint(t *testing.T) {
	srv := NewServer()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp["status"] != "ok" {
		t.Errorf("Expected status 'ok', got %q", resp["status"])
	}

	if _, ok := resp["time"]; !ok {
		t.Error("Expected 'time' field in response")
	}
}

func TestHealthEndpointWrongMethod(t *testing.T) {
	srv := NewServer()
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", rec.Code)
	}
}

func TestGenerateEndpointGLTF(t *testing.T) {
	srv := NewServer()

	params := unpeople.DefaultParams()
	params.Seed = 42
	body, _ := json.Marshal(params)

	req := httptest.NewRequest(http.MethodPost, "/generate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "model/gltf+json")
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "model/gltf+json" {
		t.Errorf("Expected Content-Type model/gltf+json, got %s", contentType)
	}

	// Verify it's valid glTF JSON
	var gltf map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&gltf); err != nil {
		t.Errorf("Response is not valid JSON: %v", err)
	}

	if _, ok := gltf["asset"]; !ok {
		t.Error("glTF response missing 'asset' field")
	}
}

func TestGenerateEndpointGLB(t *testing.T) {
	srv := NewServer()

	params := unpeople.DefaultParams()
	params.Seed = 42
	body, _ := json.Marshal(params)

	req := httptest.NewRequest(http.MethodPost, "/generate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "model/gltf-binary")
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	// Check GLB magic
	data := rec.Body.Bytes()
	if len(data) < 4 || string(data[:4]) != "glTF" {
		t.Error("Response is not valid GLB (missing magic)")
	}
}

func TestGenerateEndpointOBJ(t *testing.T) {
	srv := NewServer()

	params := unpeople.DefaultParams()
	params.Seed = 42
	body, _ := json.Marshal(params)

	req := httptest.NewRequest(http.MethodPost, "/generate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/plain")
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	output := rec.Body.String()
	if !strings.HasPrefix(output, "# Wavefront OBJ") {
		t.Error("Response is not valid OBJ")
	}
}

func TestGenerateEndpointFormatQuery(t *testing.T) {
	srv := NewServer()

	params := unpeople.DefaultParams()
	params.Seed = 42
	body, _ := json.Marshal(params)

	req := httptest.NewRequest(http.MethodPost, "/generate?format=obj", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	output := rec.Body.String()
	if !strings.HasPrefix(output, "# Wavefront OBJ") {
		t.Error("Query param format=obj should produce OBJ output")
	}
}

func TestGenerateEndpointInvalidJSON(t *testing.T) {
	srv := NewServer()

	req := httptest.NewRequest(http.MethodPost, "/generate", strings.NewReader("{invalid"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestGenerateEndpointWrongMethod(t *testing.T) {
	srv := NewServer()

	req := httptest.NewRequest(http.MethodGet, "/generate", nil)
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", rec.Code)
	}
}

func TestGenerateEndpointDefaultParams(t *testing.T) {
	srv := NewServer()

	// Empty body should use defaults
	req := httptest.NewRequest(http.MethodPost, "/generate", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCORSHeaders(t *testing.T) {
	srv := NewServer()

	req := httptest.NewRequest(http.MethodOptions, "/generate", nil)
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("Missing CORS Allow-Origin header")
	}
	if !strings.Contains(rec.Header().Get("Access-Control-Allow-Methods"), "POST") {
		t.Error("Missing POST in CORS Allow-Methods header")
	}
}

func TestRateLimiter(t *testing.T) {
	rl := NewRateLimiter(5, time.Second)

	// Should allow first 5 requests
	for i := 0; i < 5; i++ {
		if !rl.Allow() {
			t.Errorf("Request %d should be allowed", i)
		}
	}

	// 6th request should be denied
	if rl.Allow() {
		t.Error("6th request should be denied")
	}

	// After waiting, should refill
	time.Sleep(time.Second + 10*time.Millisecond)
	if !rl.Allow() {
		t.Error("Request after refill should be allowed")
	}
}

func TestDeterministicOutput(t *testing.T) {
	srv := NewServer()

	params := unpeople.DefaultParams()
	params.Seed = 42
	body, _ := json.Marshal(params)

	// Generate twice
	req1 := httptest.NewRequest(http.MethodPost, "/generate?format=obj", bytes.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")
	rec1 := httptest.NewRecorder()
	srv.ServeHTTP(rec1, req1)

	body2, _ := json.Marshal(params)
	req2 := httptest.NewRequest(http.MethodPost, "/generate?format=obj", bytes.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	srv.ServeHTTP(rec2, req2)

	if rec1.Body.String() != rec2.Body.String() {
		t.Error("Same seed should produce identical output")
	}
}
