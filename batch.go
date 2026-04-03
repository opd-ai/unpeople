// Package unpeople – parallel batch generation
//
// BatchGenerator provides a worker-pool API for generating multiple humanoid
// meshes concurrently. This is useful for populating game worlds with large
// numbers of NPCs where generation can be parallelized across CPU cores.
package unpeople

import (
	"context"
	"runtime"
	"sync"
)

// BatchResult contains the result of generating a single mesh in a batch.
type BatchResult struct {
	// Params is the input parameters used for generation.
	Params Params
	// Index is the position in the original input slice.
	Index int
	// Mesh is the generated mesh (nil if Err is set).
	Mesh *Mesh
	// Material is the generated material (nil if Err is set or not requested).
	Material *Material
	// Err is the error if generation failed.
	Err error
}

// BatchOptions configures batch generation behavior.
type BatchOptions struct {
	// Workers is the number of concurrent workers. Default (0) uses runtime.NumCPU().
	Workers int
	// IncludeMaterial requests material generation alongside each mesh.
	IncludeMaterial bool
}

// BatchGenerator generates multiple meshes concurrently using a worker pool.
// It wraps a Generator (or CachedGenerator) and distributes work across workers.
type BatchGenerator struct {
	gen       *Generator
	cachedGen *CachedGenerator
}

// NewBatchGenerator creates a BatchGenerator using a fresh Generator.
func NewBatchGenerator() *BatchGenerator {
	return &BatchGenerator{
		gen: NewGenerator(),
	}
}

// NewBatchGeneratorWithCache creates a BatchGenerator backed by a CachedGenerator.
// This is recommended when the same character variants may appear multiple times
// across batches, as cache hits avoid redundant mesh rebuilds.
func NewBatchGeneratorWithCache(cacheSize int) *BatchGenerator {
	return &BatchGenerator{
		cachedGen: NewCachedGenerator(cacheSize),
	}
}

// resolveWorkerCount returns the number of workers to use for batch generation.
func resolveWorkerCount(requested, numJobs int) int {
	workers := requested
	if workers <= 0 {
		workers = runtime.NumCPU()
	}
	if workers > numJobs {
		workers = numJobs
	}
	return workers
}

// spawnBatchWorker starts a single worker goroutine that processes jobs from the channel.
func (bg *BatchGenerator) spawnBatchWorker(
	ctx context.Context,
	wg *sync.WaitGroup,
	jobs <-chan int,
	params []Params,
	opts BatchOptions,
	results []BatchResult,
) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case idx, ok := <-jobs:
			if !ok {
				return
			}
			bg.processJob(idx, params[idx], opts.IncludeMaterial, &results[idx])
		}
	}
}

// enqueueJobs sends all job indices to the jobs channel, stopping if context is cancelled.
func enqueueJobs(ctx context.Context, jobs chan<- int, numJobs int) {
	for i := 0; i < numJobs; i++ {
		select {
		case <-ctx.Done():
			return
		case jobs <- i:
		}
	}
}

// markUnprocessedJobs sets error status for jobs that weren't processed due to cancellation.
func markUnprocessedJobs(results []BatchResult, params []Params, ctxErr error) {
	for i := range results {
		if results[i].Mesh == nil && results[i].Err == nil {
			results[i].Index = i
			results[i].Params = params[i]
			results[i].Err = ctxErr
		}
	}
}

// GenerateBatch generates meshes for all input parameters concurrently.
// Results are returned in the same order as inputs.
// The context can be used to cancel generation early.
func (bg *BatchGenerator) GenerateBatch(ctx context.Context, params []Params, opts BatchOptions) []BatchResult {
	if len(params) == 0 {
		return nil
	}

	workers := resolveWorkerCount(opts.Workers, len(params))
	results := make([]BatchResult, len(params))
	jobs := make(chan int, len(params))
	var wg sync.WaitGroup

	// Start workers
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go bg.spawnBatchWorker(ctx, &wg, jobs, params, opts, results)
	}

	enqueueJobs(ctx, jobs, len(params))
	close(jobs)

	wg.Wait()
	markUnprocessedJobs(results, params, ctx.Err())

	return results
}

// processJob generates a single mesh and stores the result.
func (bg *BatchGenerator) processJob(idx int, p Params, includeMaterial bool, result *BatchResult) {
	result.Index = idx
	result.Params = p

	var mesh *Mesh
	var err error

	if bg.cachedGen != nil {
		mesh, err = bg.cachedGen.Generate(p)
	} else {
		mesh, err = bg.gen.Generate(p)
	}

	if err != nil {
		result.Err = err
		return
	}

	result.Mesh = mesh

	if includeMaterial {
		skinColor := ComputeSkinColor(p.SkinTone, p.SkinUndertone)
		mat := DefaultSkinMaterial(skinColor)
		result.Material = &mat
	}
}

// GenerateBatchSimple is a convenience method that generates meshes without
// material data and uses default options.
func (bg *BatchGenerator) GenerateBatchSimple(ctx context.Context, params []Params) []*Mesh {
	results := bg.GenerateBatch(ctx, params, BatchOptions{})
	meshes := make([]*Mesh, len(results))
	for i, r := range results {
		meshes[i] = r.Mesh
	}
	return meshes
}

// GenerateBatchWithMaterial generates meshes with accompanying materials
// using default worker count.
func (bg *BatchGenerator) GenerateBatchWithMaterial(ctx context.Context, params []Params) []BatchResult {
	return bg.GenerateBatch(ctx, params, BatchOptions{
		IncludeMaterial: true,
	})
}

// CacheStats returns cache statistics if using a cached generator.
// Returns (0, 0) if not using cache.
func (bg *BatchGenerator) CacheStats() (size, maxSize int) {
	if bg.cachedGen != nil {
		return bg.cachedGen.CacheStats()
	}
	return 0, 0
}
