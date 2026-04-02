package unpeople

// splitmix64 is a small, self-contained PRNG whose algorithm is fixed in this
// package and will therefore produce identical output regardless of the Go
// toolchain version.  Using stdlib math/rand would couple determinism to Go's
// internal PRNG implementation, which is not guaranteed stable across releases.
//
// Algorithm: SplitMix64 by Sebastiano Vigna (public domain).
// Reference: https://xoshiro.di.unimi.it/splitmix64.c
type splitmix64 struct {
	state uint64
}

func newSplitmix64(seed int64) *splitmix64 {
	return &splitmix64{state: uint64(seed)}
}

func (r *splitmix64) next() uint64 {
	r.state += 0x9e3779b97f4a7c15
	z := r.state
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	return z ^ (z >> 31)
}

// Float32 returns a pseudo-random float32 in [0, 1).
func (r *splitmix64) Float32() float32 {
	// Use the top 24 bits for a float32 mantissa.
	return float32(r.next()>>40) / float32(1<<24)
}
