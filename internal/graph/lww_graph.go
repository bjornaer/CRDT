package graph

import (
	"errors"
	"fmt"
	"sync"
	"time"

	set "github.com/bjornaer/crdt/internal/set"
)

type LastWriterWinsGraph[T comparable] interface {
	AddVertex(T) error
	GetAllVertices() ([]T, error)
	RemoveVertex(T) error
	VertexExists(T) bool
	AddEdge(v1, v2 T) error
	RemoveEdge(v1, v2 T) error
	EdgeExists(v1, v2 T) bool
	GetVertexEdges(v T) ([]T, error)
	FindPath(v1, v2 T) ([]T, error)
	Merge(LastWriterWinsGraph[T]) error
	getV() set.LastWriterWinsSet[T]
	getE() map[T]set.LastWriterWinsSet[T]
}

// LWWGraph is a structure for a graph with vertices and edges based on LWW sets
type LWWGraph[T comparable] struct {
	vertices set.LastWriterWinsSet[T]
	edges    map[T]set.LastWriterWinsSet[T]
	mutex    sync.RWMutex // Maps in Go are not thread safe by default and that's why we use a mutex
}

// NewLWWGraph returns an empty LWW based LWWGraph
func NewLWWGraph[T comparable]() LastWriterWinsGraph[T] {
	return &LWWGraph[T]{
		vertices: set.NewLWWSet[T](),
	}
}

// access private vertices
func (g *LWWGraph[T]) getV() set.LastWriterWinsSet[T] {
	return g.vertices
}

// access private edges
func (g *LWWGraph[T]) getE() map[T]set.LastWriterWinsSet[T] {
	return g.edges
}

// AddVertex adds a vertex to the graph
func (g *LWWGraph[T]) AddVertex(v T) error {
	return g.vertices.Add(v, time.Now())
}

// GetAllVertices get all vertices from the LWWGraph
func (g *LWWGraph[T]) GetAllVertices() ([]T, error) {
	return g.vertices.Get()
}

// RemoveVertex removes a vertex from the LWWGraph
func (g *LWWGraph[T]) RemoveVertex(v T) error {
	return g.vertices.Remove(v, time.Now())
}

// VertexExists checks if a vertex is in the LWWGraph
func (g *LWWGraph[T]) VertexExists(v T) bool {
	return g.vertices.Exists(v)
}

// AddEdge adds an edge to the LWWGraph
func (g *LWWGraph[T]) AddEdge(v1, v2 T) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if !g.vertices.Exists(v1) {
		return fmt.Errorf("cannot add edge, missing node in graph: %v", v1)
	} else if !g.vertices.Exists(v2) {
		return fmt.Errorf("cannot add edge, missing node in graph: %v", v2)
	}

	if g.edges == nil {
		g.edges = make(map[T]set.LastWriterWinsSet[T])
	}
	if _, ok := g.edges[v1]; !ok {
		g.edges[v1] = set.NewLWWSet[T]()
	}
	err := g.edges[v1].Add(v2, time.Now())
	if err != nil {
		return err
	}
	if _, ok := g.edges[v2]; !ok {
		g.edges[v2] = set.NewLWWSet[T]()
	}
	err = g.edges[v2].Add(v1, time.Now())
	if err != nil {
		return err
	}
	return nil
}

// RemoveEdge removes an edge from the LWWGraph
func (g *LWWGraph[T]) RemoveEdge(v1, v2 T) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if g.edges == nil {
		g.edges = make(map[T]set.LastWriterWinsSet[T])
	}
	if _, ok := g.edges[v1]; !ok {
		g.edges[v1] = set.NewLWWSet[T]()
	}
	err := g.edges[v1].Remove(v2, time.Now())
	if err != nil {
		return err
	}
	if _, ok := g.edges[v2]; !ok {
		g.edges[v2] = set.NewLWWSet[T]()
	}
	err = g.edges[v2].Remove(v1, time.Now())
	if err != nil {
		return err
	}
	return nil
}

// EdgeExists checks if two vertices share an edge
func (g *LWWGraph[T]) EdgeExists(v1, v2 T) bool {
	return g.edges[v1].Exists(v2) && g.edges[v2].Exists(v1)
}

// GetVertexEdges allows querying for all vertices connected to a single vertex
func (g *LWWGraph[T]) GetVertexEdges(v T) ([]T, error) {
	if !g.VertexExists(v) {
		return nil, errors.New("cannot query for edges, vertex does not exist")
	}
	return g.edges[v].Get()
}

// FindPath finds a connecting path between two given vertices
func (g *LWWGraph[T]) FindPath(v1, v2 T) ([]T, error) {
	if !g.vertices.Exists(v1) {
		return nil, fmt.Errorf("cannot find path, missing node in graph: %v", v1)
	} else if !g.vertices.Exists(v2) {
		return nil, fmt.Errorf("cannot find path, missing node in graph: %v", v2)
	}

	seen := set.NewLWWSet[T]()
	var emptyPath []T
	_, path, err := g.findPathRecursive(v1, v2, seen, emptyPath)
	if err != nil {
		return nil, err
	}

	return path, nil
}

func (g *LWWGraph[T]) findPathRecursive(
	v1,
	v2 T,
	seen set.LastWriterWinsSet[T],
	path []T) (set.LastWriterWinsSet[T], []T, error) {
	err := seen.Add(v1, time.Now())
	path = append(path, v1)
	if err != nil {
		return nil, nil, err
	}

	if v1 == v2 {
		return seen, path, nil
	}

	edges, err := g.edges[v1].Get()
	if err != nil {
		return nil, nil, err
	}

	for _, vertex := range edges {
		if !seen.Exists(vertex) {
			newSeen, newPath, err := g.findPathRecursive(vertex, v2, seen, path)
			if err != nil {
				return nil, nil, err
			}
			if newSeen.Exists(v2) {
				path = newPath
				seen = newSeen
				break
			}
		}
	}
	return seen, path, nil
}

// Merge another LWWGraph into its instance by merging vertices and edges
func (g *LWWGraph[T]) Merge(other LastWriterWinsGraph[T]) error {
	if other == nil {
		return errors.New("cannot merge, other graph is nil")
	}

	err := g.vertices.Merge(other.getV())
	if err != nil {
		return err
	}

	for otherVertex, otherEdges := range other.getE() {
		if currentEdges, ok := g.edges[otherVertex]; ok {
			err = currentEdges.Merge(otherEdges)
			if err != nil {
				return err
			}
		} else {
			g.edges[otherVertex] = otherEdges
		}
	}
	return nil
}
