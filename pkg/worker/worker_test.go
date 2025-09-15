package worker

import (
	"reflect"
	"sort"
	"testing"

	"github.com/rodrigocitadin/mapreduce/pkg/types"
)

func TestMapF(t *testing.T) {
	contents := "hello world goodbye world"
	expected := []types.KeyValue{
		{Key: "goodbye", Value: "1"},
		{Key: "hello", Value: "1"},
		{Key: "world", Value: "1"},
		{Key: "world", Value: "1"},
	}

	result := mapF(contents)

	// Sort both slices for consistent comparison
	sort.Slice(result, func(i, j int) bool {
		return result[i].Key < result[j].Key
	})

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("mapF() = %v, want %v", result, expected)
	}
}

func TestReduceF(t *testing.T) {
	values := []string{"1", "1", "1"}
	expected := "3"

	result := reduceF(values)

	if result != expected {
		t.Errorf("reduceF() = %v, want %v", result, expected)
	}
}

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
