package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"sync/atomic"
	"time"
)

var counter uint32

// GenerateID generates a unique 24-character hexadecimal ID.
//
// The ID contains:
//   - 4 bytes of Unix timestamp
//   - 5 bytes of cryptographic randomness
//   - 3 bytes of process-local counter
func GenerateID() string {
	id := make([]byte, 12)

	// Timestamp
	binary.BigEndian.PutUint32(id[0:4], uint32(time.Now().Unix()))

	// Randomness
	if _, err := rand.Read(id[4:9]); err != nil {
		panic(err)
	}

	// Counter
	value := atomic.AddUint32(&counter, 1)

	id[9] = byte(value >> 16)
	id[10] = byte(value >> 8)
	id[11] = byte(value)

	return hex.EncodeToString(id)
}

// Scope represents the scope of an idempotency key.
type Scope string

// Generator generates deterministic idempotency keys.
type Generator struct{}

// NewGenerator creates a new idempotency key generator.
func NewGenerator() *Generator {
	return &Generator{}
}

// GenerateKey generates an idempotency key from a scope and parameters.
func (g *Generator) GenerateKey(scope Scope, params map[string]interface{}) string {
	// Sort params for deterministic output.
	keys := make([]string, 0, len(params))

	for key := range params {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	// Build deterministic hash input.
	var b strings.Builder

	b.WriteString(string(scope))

	for _, key := range keys {
		fmt.Fprintf(
			&b,
			":%s=%v",
			key,
			params[key],
		)
	}

	// SHA-256
	hash := sha256.Sum256([]byte(b.String()))

	// First 8 bytes = 16 hex characters.
	return fmt.Sprintf(
		"%s-%s",
		scope,
		hex.EncodeToString(hash[:8]),
	)
}

// ValidateKey validates whether an idempotency key
// matches the expected scope and parameters.
func (g *Generator) ValidateKey(
	scope Scope,
	params map[string]interface{},
	key string,
) bool {
	return g.GenerateKey(scope, params) == key
}
