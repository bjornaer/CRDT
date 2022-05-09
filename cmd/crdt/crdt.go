package crdt

import (
	g "github.com/bjornaer/crdt/internal/graph"
	s "github.com/bjornaer/crdt/internal/set"
)

type CRDT[T comparable] struct {
	LWWGraph g.LastWriterWinsGraph[T]
	LWWSet   s.LastWriterWinsSet[T]
}

// NewLWWSet returns an implementation of a LastWriterWinsSet
func NewCRDT[T comparable]() *CRDT[T] {
	return &CRDT[T]{
		LWWSet:   s.NewLWWSet[T](),
		LWWGraph: g.NewLWWGraph[T](),
	}
}
