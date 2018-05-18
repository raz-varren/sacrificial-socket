package ss

import (
	"math/rand"
	"sync"
	"time"
)

//RNG is a random number generator that is safe for concurrent use by multiple go routines
type RNG struct {
	r  *rand.Rand
	mu *sync.Mutex
}

//Read reads len(b) random bytes into b and never returns a nil error
func (r *RNG) Read(b []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.r.Read(b)
}

//NewRNG creates a new random number generator
func NewRNG() *RNG {
	return &RNG{
		r:  rand.New(rand.NewSource(time.Now().UnixNano())),
		mu: &sync.Mutex{},
	}
}
