package set

import (
	"time"

	backends "github.com/bjornaer/crdt/internal/backends"
)

type LastWriterWinsSet interface {
	Add(interface{}, time.Time) error
	Remove(interface{}, time.Time) error
	Exists(interface{}) bool
	Get() ([]interface{}, error)
	Merge(LastWriterWinsSet) error
	getAdditions() backends.TimeSet
	getRemovals() backends.TimeSet
}

// LWWSet is a Last-Writer-Wins Set implementation
type LWWSet struct {
	additions backends.TimeSet
	removals  backends.TimeSet
}

// Add marks an element to be added at a given timestamp
func (s *LWWSet) Add(value interface{}, t time.Time) error {
	return s.additions.Add(value, t)
}

func (s *LWWSet) getAdditions() backends.TimeSet {
	return s.additions
}

// Remove marks an element to be removed at a given timestamp
func (s *LWWSet) Remove(value interface{}, t time.Time) error {
	return s.removals.Add(value, t)
}

func (s *LWWSet) getRemovals() backends.TimeSet {
	return s.removals
}

// Exists checks if an element is marked as present in the set
func (s LWWSet) Exists(value interface{}) bool {
	addedAt, added := s.additions.AddedAt(value)

	removed := s.isRemoved(value, addedAt)

	return added && !removed
}

// isRemoved checks if an element is marked for removal
func (s LWWSet) isRemoved(value interface{}, since time.Time) bool {
	removedAt, removed := s.removals.AddedAt(value)

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

	err := s.additions.Each(func(element interface{}, addedAt time.Time) error {
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
	err := other.getAdditions().Each(func(element interface{}, addedAt time.Time) error {
		err := s.Add(element, addedAt)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	err = other.getRemovals().Each(func(element interface{}, addedAt time.Time) error {
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
		additions: backends.NewTimeSet(),
		removals:  backends.NewTimeSet(),
	}
}
