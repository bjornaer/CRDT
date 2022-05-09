package set

import (
	"time"

	backends "github.com/bjornaer/crdt/internal/backends"
)

type LastWriterWinsSet[T comparable] interface {
	Add(T, time.Time) error
	Remove(T, time.Time) error
	Exists(T) bool
	Get() ([]T, error)
	Merge(LastWriterWinsSet[T]) error
	getAdditions() backends.TimeSet[T]
	getRemovals() backends.TimeSet[T]
}

// LWWSet is a Last-Writer-Wins Set implementation
type LWWSet[T comparable] struct {
	additions backends.TimeSet[T]
	removals  backends.TimeSet[T]
}

// Add marks an element to be added at a given timestamp
func (s *LWWSet[T]) Add(value T, t time.Time) error {
	return s.additions.Add(value, t)
}

func (s *LWWSet[T]) getAdditions() backends.TimeSet[T] {
	return s.additions
}

// Remove marks an element to be removed at a given timestamp
func (s *LWWSet[T]) Remove(value T, t time.Time) error {
	return s.removals.Add(value, t)
}

func (s *LWWSet[T]) getRemovals() backends.TimeSet[T] {
	return s.removals
}

// Exists checks if an element is marked as present in the set
func (s LWWSet[T]) Exists(value T) bool {
	addedAt, added := s.additions.AddedAt(value)

	removed := s.isRemoved(value, addedAt)

	return added && !removed
}

// isRemoved checks if an element is marked for removal
func (s LWWSet[T]) isRemoved(value T, since time.Time) bool {
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
func (s LWWSet[T]) Get() ([]T, error) {
	var result []T

	err := s.additions.Each(func(element T, addedAt time.Time) error {
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
func (s LWWSet[T]) Merge(other LastWriterWinsSet[T]) error {
	err := other.getAdditions().Each(func(element T, addedAt time.Time) error {
		err := s.Add(element, addedAt)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	err = other.getRemovals().Each(func(element T, addedAt time.Time) error {
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
func NewLWWSet[T comparable]() LastWriterWinsSet[T] {
	return &LWWSet[T]{
		additions: backends.NewTimeSet[T](),
		removals:  backends.NewTimeSet[T](),
	}
}
