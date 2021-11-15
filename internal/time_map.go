package internal


import (
"sync"
"time"
)

type timeSet interface {
	add(interface{}, time.Time) error
	addedAt(interface{}) (time.Time, bool)
	each(func(interface{}, time.Time) error) error
}

// timeMap is an implementation of a timeSet that uses a map data structure. We map items to timestamps.
type timeMap struct {
	elements map[interface{}]time.Time
	mutex    sync.RWMutex // Maps in Go are not thread safe by default and that's why we use a mutex
}

// Add an element in the set if one of the following condition is met:
// - Given element does not exist yet
// - Given element already exists but with a lesser timestamp than the given one
func (s *timeMap) add(value interface{}, t time.Time) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	addedAt, ok := s.elements[value]
	if !ok || (ok && t.After(addedAt)) {
		s.elements[value] = t
	}
	return nil
}

// AddedAt returns the timestamp of a given element if it exists
//
// The second return value (bool) indicates whether the element exists or not
// If the given element does not exist, the second return (bool) is false
func (s *timeMap) addedAt(value interface{}) (time.Time, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	t, ok := s.elements[value]
	return t, ok
}

// Each traverses the items in the Set, calling the provided function
// for each element/timestamp association
func (s *timeMap) each(f func(element interface{}, addedAt time.Time) error) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	for element, addedAt := range s.elements {
		err := f(element, addedAt)
		if err != nil {
			return err
		}
	}
	return nil
}

// newTimeSet returns an empty map-backed implementation of the time set interface
func newTimeSet() timeSet {
	return &timeMap{
		elements: make(map[interface{}]time.Time),
	}
}


