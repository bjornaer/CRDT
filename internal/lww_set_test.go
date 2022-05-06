package internal_test

// We add the test file in a separate package to keep testing and actual implementation details separate
// By doing so, in the tests we will only have access to the public part of our code

import (
	"github.com/bjornaer/crdt/internal"
	"testing"
	"time"
)

func setupTestSet() internal.LastWriterWinsSet {
	i1 := "item1"
	i2 := "item2"
	i3 := "item3"

	s := internal.NewLWWSet()
	s.Add(i1, time.Now())
	s.Add(i2, time.Now())
	s.Add(i3, time.Now())
	return s
}

// checks element is contained within set
func contains(s []interface{}, e interface{}) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// checks for set equality -- independent of order
func setsAreEqual(s1, s2 []interface{}) bool {
	if len(s1) != len(s2) {
		return false
	}
	for _, v := range s1 {
		if !contains(s2, v) {
			return false
		}
	}
	return true
}

func TestLWWSet_Add(t *testing.T) {
	s := setupTestSet()
	i := "item4"
	err := s.Add(i, time.Now())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	items, err := s.Get()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	expected := []interface{}{"item1", "item2", "item3", "item4"}
	if !s.Exists(i) || !setsAreEqual(items, expected) {
		t.Errorf("Missing item, got: %v, expected: %v.", items, expected)
	}
}

func TestLWWSet_Get(t *testing.T) {
	s := setupTestSet()
	items, err := s.Get()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	expected := []interface{}{"item1", "item2", "item3"}
	if !setsAreEqual(items, expected) {
		t.Errorf("Items mismatch, got: %v, expected: %v.", items, expected)
	}
}

func TestLWWSet_Remove(t *testing.T) {
	s := setupTestSet()
	i := "item3"
	err := s.Remove(i, time.Now())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	items, err := s.Get()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	expected := []interface{}{"item1", "item2"}
	if s.Exists(i) || !setsAreEqual(items, expected) {
		t.Errorf("Extra item found, got: %v, expected: %v.", items, expected)
	}
}

func TestLWWSet_Exists(t *testing.T) {
	s := setupTestSet()
	tests := []struct {
		vertex   string
		expected bool
	}{
		{"item1", true},
		{"inexistent_item", false},
	}
	for _, tt := range tests {
		got := s.Exists(tt.vertex)
		if got != tt.expected {
			t.Errorf("Existence check failed, got: %v, expected: %v.", got, tt.expected)
		}
	}
}

func TestLWWSet_Merge(t *testing.T) {
	s1 := setupTestSet()
	s2 := setupTestSet()
	i2 := "item2"
	ni := "new_item"
	s2.Remove(i2, time.Now())
	s1.Add(i2, time.Now())
	s1.Add(ni, time.Now())
	s2.Remove(ni, time.Now())
	err := s1.Merge(s2)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !s1.Exists(i2) || s1.Exists(ni) {
		t.Errorf("merge failed, items mismatch")
	}
}

// testing associativity behavior when merging sets
// ie: s1 v (s2 v s3) = (s1 v s2) v s3
func TestLWWSet_Associativity(t *testing.T) {
	// "s1 v (s2 v s3)"
	s1 := setupTestSet()
	s2 := setupTestSet()
	s3 := setupTestSet()
	s1.Add("item4", time.Now())
	s3.Remove("item4", time.Now())
	s2.Merge(s3)
	s1.Merge(s2)
	// "(s1 v s2) v s3"
	// where "s4 ~ s1", "s5 ~ s2", "s6 ~ s3" (because timestamps are not equal)
	s4 := setupTestSet()
	s5 := setupTestSet()
	s6 := setupTestSet()
	s4.Add("item4", time.Now())
	s6.Remove("item4", time.Now())
	s4.Merge(s5)
	s4.Merge(s6)

	// s1 and s4 hold the results of the union operations, they should contain equal elements
	// although timestamps mismatch the order of operations is equal, given the underlying monotonicity
	// we should end up with equal items on both graphs
	items1, _ := s1.Get()
	items4, _ := s4.Get()
	if !setsAreEqual(items1, items4) {
		t.Errorf("Merge not associative, s1 v (s2 v s3): %v, (s1 v s2) v s3: %v.", items1, items4)
	}
}

// testing commutativity behavior when merging sets
// // ie: s1 v s2 = s2 v s1
func TestLWWSet_Commutativity(t *testing.T) {
	// "s1 v s2"
	s1 := setupTestSet()
	s2 := setupTestSet()
	s1.Add("item4", time.Now())
	s2.Remove("item4", time.Now())
	s1.Merge(s2)
	// "s2 v s1"
	// where "s3 ~ s1", "s4 ~ s2" (because timestamps are not equal)
	s3 := setupTestSet()
	s4 := setupTestSet()
	s3.Add("item4", time.Now())
	s4.Remove("item4", time.Now())
	s4.Merge(s3)

	// s1 and s4 hold the results of the union operations, they should contain equal elements
	// although timestamps mismatch the order of operations is equal, given the underlying monotonicity
	// we should end up with equal vertices on both graphs
	items1, _ := s1.Get()
	items4, _ := s4.Get()
	if !setsAreEqual(items1, items4) {
		t.Errorf("Merge not associative, s1 v s2: %v, s2 v s1: %v.", items1, items4)
	}
}

// testing idempotence behavior when merging sets
// // ie: s1 v s1 = s1
func TestLWWSet_Idempotence(t *testing.T) {
	// s and sCopy act as the same structure, modified in the same manner in two instances separately
	s := setupTestSet()
	sCopy := setupTestSet()
	s.Add("item4", time.Now())
	s.Remove("item3", time.Now())
	sCopy.Add("item4", time.Now())
	sCopy.Remove("item3", time.Now())

	beforeItems, _ := sCopy.Get()
	s.Merge(sCopy)
	afterItems, _ := s.Get()
	if !setsAreEqual(beforeItems, afterItems) {
		t.Errorf("Merge not idempotent, g1: %v, g1 v g1: %v.", beforeItems, afterItems)
	}
}
