package crdt

import (
	"github.com/bjornaer/crdt/internal/graph"
	"github.com/bjornaer/crdt/internal/set"
)

type LastWriterWinsSet[T comparable] interface {
	set.LastWriterWinsSet[T]
}

type LastWriterWinsGraph[T comparable] interface {
	graph.LastWriterWinsGraph[T]
}

func NewLWWSet[T comparable]() LastWriterWinsSet[T] {
	return set.NewLWWSet[T]()
}

func NewLWWGraph[T comparable]() LastWriterWinsGraph[T] {
	return graph.NewLWWGraph[T]()
}
