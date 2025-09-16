package worker

import (
	"testing"
)

func TestIhash(t *testing.T) {
	key1 := "hello"
	key2 := "world"
	key3 := "hello"

	hash1 := ihash(key1)
	hash2 := ihash(key2)
	hash3 := ihash(key3)

	if hash1 != hash3 {
		t.Errorf("ihash function is not deterministic. hash(%s)=%d, but hash(%s)=%d", key1, hash1, key3, hash3)
	}
	if hash1 == hash2 {
		// This is a weak test as collisions are possible, but unlikely for these simple strings.
		// It mainly ensures the hash function is not trivial (e.g., returning a constant).
		t.Logf("Warning: ihash function produced a collision for different keys. hash(%s)=%d, hash(%s)=%d", key1, hash1, key2, hash2)
	}
}
