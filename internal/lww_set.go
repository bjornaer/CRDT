package internal

import (
	"time"
)

type LastWriterWinsSet interface {
	Add(interface{}, time.Time) error
	Remove(interface{}, time.Time) error
	Exists(interface{}) bool
	Get() ([]interface{}, error)
	Merge(LastWriterWinsSet) error
	getAdditions() timeSet
	getRemovals() timeSet
}

// LWWSet is a Last-Writer-Wins Set implementation
type LWWSet struct {
	additions timeSet
	removals  timeSet
}

// Add marks an element to be added at a given timestamp
func (s *LWWSet) Add(value interface{}, t time.Time) error {
	return s.additions.add(value, t)
}

func (s *LWWSet) getAdditions() timeSet {
	return s.additions
}

// Remove marks an element to be removed at a given timestamp
func (s *LWWSet) Remove(value interface{}, t time.Time) error {
	return s.removals.add(value, t)
}

func (s *LWWSet) getRemovals() timeSet {
	return s.removals
}

// Exists checks if an element is marked as present in the set
func (s LWWSet) Exists(value interface{}) bool {
	addedAt, added := s.additions.addedAt(value)

	removed := s.isRemoved(value, addedAt)

	return added && !removed
}

// isRemoved checks if an element is marked for removal
func (s LWWSet) isRemoved(value interface{}, since time.Time) bool {
	removedAt, removed := s.removals.addedAt(value)

	if !removed {
		return false
	}
	if since.Before(removedAt) {
		return true
	}
	return false
}

// Get returns set content
func (s LWWSet) Get() ([]interface{}, error) {
	var result []interface{}

	err := s.additions.each(func(element interface{}, addedAt time.Time) error {
		removed := s.isRemoved(element, addedAt)

		if !removed {
			result = append(result, element)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// Merge additions and removals from other LWWSet into current set
func (s LWWSet) Merge(other LastWriterWinsSet) error {
	err := other.getAdditions().each(func(element interface{}, addedAt time.Time) error {
		err := s.Add(element, addedAt)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	err = other.getRemovals().each(func(element interface{}, addedAt time.Time) error {
		err := s.Remove(element, addedAt)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// NewLWWSet returns an implementation of a LastWriterWinsSet
func NewLWWSet() LastWriterWinsSet {
	return &LWWSet{
		additions: newTimeSet(),
		removals:  newTimeSet(),
	}
}
