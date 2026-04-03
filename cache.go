// Package unpeople – mesh caching
//
// CachedGenerator wraps a Generator with an LRU cache to avoid regenerating
// identical meshes when the same Params are requested multiple times.
package unpeople

import "sync"

// cacheEntry holds a cached mesh along with its generation order for LRU eviction.
type cacheEntry struct {
	mesh     *Mesh
	lastUsed uint64 // monotonic counter for LRU ordering
}

// CachedGenerator wraps a Generator with an in-memory LRU cache.
// Repeated calls with identical Params return the cached mesh immediately
// without rebuilding the geometry.
//
// CachedGenerator is safe for concurrent use from multiple goroutines.
type CachedGenerator struct {
	gen      *Generator
	cache    map[string]*cacheEntry
	maxSize  int
	mu       sync.RWMutex
	useCount uint64 // monotonic counter for LRU tracking
}

// NewCachedGenerator creates a CachedGenerator with the specified maximum
// cache size. When the cache reaches maxSize entries, the least recently
// used entry is evicted before adding a new one.
//
// A maxSize of 0 disables caching entirely (equivalent to using Generator directly).
// Typical values for maxSize range from 100 to 10000 depending on available memory
// and the number of unique character variants expected.
func NewCachedGenerator(maxSize int) *CachedGenerator {
	return &CachedGenerator{
		gen:     NewGenerator(),
		cache:   make(map[string]*cacheEntry),
		maxSize: maxSize,
	}
}

// tryCacheHit attempts to retrieve a mesh from the cache.
// Returns the mesh if found, nil otherwise.
func (cg *CachedGenerator) tryCacheHit(key string) *Mesh {
	cg.mu.RLock()
	entry, ok := cg.cache[key]
	cg.mu.RUnlock()

	if ok {
		cg.mu.Lock()
		cg.useCount++
		entry.lastUsed = cg.useCount
		cg.mu.Unlock()
		return entry.mesh
	}
	return nil
}

// storeMeshInCache adds a mesh to the cache, evicting LRU if necessary.
// Returns the stored mesh (may be a previously cached mesh if concurrent insert).
func (cg *CachedGenerator) storeMeshInCache(key string, mesh *Mesh) *Mesh {
	cg.mu.Lock()
	defer cg.mu.Unlock()

	// Double-check another goroutine didn't add it while we were generating
	if entry, ok := cg.cache[key]; ok {
		cg.useCount++
		entry.lastUsed = cg.useCount
		return entry.mesh
	}

	// Evict if at capacity
	if len(cg.cache) >= cg.maxSize {
		cg.evictLRU()
	}

	// Add to cache
	cg.useCount++
	cg.cache[key] = &cacheEntry{
		mesh:     mesh,
		lastUsed: cg.useCount,
	}

	return mesh
}

// Generate produces a humanoid Mesh from the supplied parameters.
// If a mesh for the exact same Params exists in the cache, it is returned
// immediately. Otherwise, the mesh is generated, cached, and returned.
//
// The returned *Mesh should be treated as read-only; modifying it will
// affect subsequent cache hits for the same Params.
func (cg *CachedGenerator) Generate(p Params) (*Mesh, error) {
	if cg.maxSize <= 0 {
		return cg.gen.Generate(p)
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}

	key := paramsToKey(p)

	if mesh := cg.tryCacheHit(key); mesh != nil {
		return mesh, nil
	}

	mesh, err := cg.gen.Generate(p)
	if err != nil {
		return nil, err
	}

	return cg.storeMeshInCache(key, mesh), nil
}

// GenerateWithMaterial produces a MeshWithMaterial with caching support.
// The mesh component is cached; the material is always freshly computed
// (since it's cheap and may vary with params not in the mesh key).
func (cg *CachedGenerator) GenerateWithMaterial(p Params) (*MeshWithMaterial, error) {
	mesh, err := cg.Generate(p)
	if err != nil {
		return nil, err
	}

	skinColor := ComputeSkinColor(p.SkinTone, p.SkinUndertone)
	material := DefaultSkinMaterial(skinColor)

	return &MeshWithMaterial{
		Mesh:     mesh,
		Material: material,
	}, nil
}

// evictLRU removes the least recently used cache entry.
// Caller must hold cg.mu write lock.
func (cg *CachedGenerator) evictLRU() {
	var oldestKey string
	var oldestTime uint64 = ^uint64(0) // max uint64

	for key, entry := range cg.cache {
		if entry.lastUsed < oldestTime {
			oldestTime = entry.lastUsed
			oldestKey = key
		}
	}

	if oldestKey != "" {
		delete(cg.cache, oldestKey)
	}
}

// CacheSize returns the current number of entries in the cache.
func (cg *CachedGenerator) CacheSize() int {
	cg.mu.RLock()
	defer cg.mu.RUnlock()
	return len(cg.cache)
}

// CacheHitRate returns the approximate cache hit rate.
// This is calculated from the number of cache entries and total use count.
// For a more accurate metric, call ResetStats first and check after use.
func (cg *CachedGenerator) CacheStats() (size, maxSize int) {
	cg.mu.RLock()
	defer cg.mu.RUnlock()
	return len(cg.cache), cg.maxSize
}

// ClearCache removes all entries from the cache.
func (cg *CachedGenerator) ClearCache() {
	cg.mu.Lock()
	defer cg.mu.Unlock()
	cg.cache = make(map[string]*cacheEntry)
	cg.useCount = 0
}

// Invalidate removes a specific parameter set from the cache.
// Returns true if an entry was removed, false if not found.
func (cg *CachedGenerator) Invalidate(p Params) bool {
	if err := p.Validate(); err != nil {
		return false
	}

	key := paramsToKey(p)

	cg.mu.Lock()
	defer cg.mu.Unlock()

	if _, ok := cg.cache[key]; ok {
		delete(cg.cache, key)
		return true
	}
	return false
}

// paramsToKey generates a cache key from Params.
// This must match the key format used in Generator.Generate.
func paramsToKey(p Params) string {
	hairSlot := 0
	if p.HasHairSlot {
		hairSlot = 1
	}

	// This format must stay in sync with generator.go
	return "humanoid_sp" + itoa(int(p.Species)) +
		"_ht" + itoa(int(p.Height)) +
		"_bl" + itoa(int(p.Build)) +
		"_pr" + itoa(int(p.Proportions)) +
		"_ph" + itoa(int(p.Phenotype)) +
		"_ag" + itoa(int(p.Age)) +
		"_po" + itoa(int(p.Posture)) +
		"_fs" + itoa(int(p.FaceShape)) +
		"_jw" + itoa(int(p.Jaw)) +
		"_br" + itoa(int(p.Brow)) +
		"_er" + itoa(int(p.Ears)) +
		"_sw" + itoa(int(p.ShoulderWidth)) +
		"_hw" + itoa(int(p.HipWidth)) +
		"_ll" + itoa(int(p.LimbLength)) +
		"_nl" + itoa(int(p.NeckLength)) +
		"_hs" + itoa(int(p.HandSize)) +
		"_fl" + itoa(int(p.FingerLength)) +
		"_ft" + itoa(int(p.FootSize)) +
		"_hr" + itoa(hairSlot) +
		"_sk" + itoa(int(p.SkinTone)) +
		"_ut" + itoa(int(p.SkinUndertone)) +
		"_se" + itoa64(p.Seed)
}

// itoa converts an int to a decimal string without importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

// itoa64 converts an int64 to a decimal string without importing strconv.
func itoa64(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [21]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
