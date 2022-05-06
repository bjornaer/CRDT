package crdt

import (
	g "github.com/bjornaer/crdt/internal/graph"
	s "github.com/bjornaer/crdt/internal/set"
)

type CRDT struct {
	LWWGraph g.LastWriterWinsGraph
	LWWSet   s.LastWriterWinsSet
}

// NewLWWSet returns an implementation of a LastWriterWinsSet
func NewCRDT() *CRDT {
	return &CRDT{
		LWWSet:   s.NewLWWSet(),
		LWWGraph: g.NewLWWGraph(),
	}
}
