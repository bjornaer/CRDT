package internal

import (
	"sync"
	"time"
)

type TimeSet[T comparable] interface {
	Add(T, time.Time) error
	AddedAt(T) (time.Time, bool)
	Each(func(T, time.Time) error) error
	Size() int
}

// TimeMap is an implementation of a timeSet that uses a map data structure. We map items to timestamps.
type TimeMap[T comparable] struct {
	Elements map[T]time.Time `json:"elements"`
	mutex    sync.RWMutex // Maps in Go are not thread safe by default and that's why we use a mutex
}

// Add an element in the set if one of the following condition is met:
// - Given element does not exist yet
// - Given element already exists but with a lesser timestamp than the given one
func (s *TimeMap[T]) Add(value T, t time.Time) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	addedAt, ok := s.Elements[value]
	if !ok || (ok && t.After(addedAt)) {
		s.Elements[value] = t
	}
	return nil
}

// AddedAt returns the timestamp of a given element if it exists
//
// The second return value (bool) indicates whether the element exists or not
// If the given element does not exist, the second return (bool) is false
func (s *TimeMap[T]) AddedAt(value T) (time.Time, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	t, ok := s.Elements[value]
	return t, ok
}

// Each traverses the items in the Set, calling the provided function
// for each element/timestamp association
func (s *TimeMap[T]) Each(f func(element T, addedAt time.Time) error) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	for element, addedAt := range s.Elements {
		err := f(element, addedAt)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *TimeMap[T]) Size() int {
	size := 0
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	for range s.Elements {
		size++
	}
	return size
}

// newTimeSet returns an empty map-backed implementation of the time set interface
func NewTimeSet[T comparable]() TimeSet[T] {
	return &TimeMap[T]{
		Elements: make(map[T]time.Time),
	}
}
