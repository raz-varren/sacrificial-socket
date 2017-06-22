/*
Package tools is really just used during socket creation to generate random numbers for socket IDs.
*/
package tools

import (
	crand "crypto/rand"
	"fmt"
	"io"
	"math/rand"
	"time"
)

func RandomInt(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min+1) + min
}

func RandomInt64(min, max int64) int64 {
	rand.Seed(time.Now().UnixNano())
	return rand.Int63n(max-min+1) + min
}

func UID() string {
	uid := make([]byte, 16)
	io.ReadFull(crand.Reader, uid)
	return fmt.Sprintf("%x", uid)
}
