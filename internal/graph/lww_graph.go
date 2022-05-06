package graph

import (
	"errors"
	"fmt"
	"sync"
	"time"

	set "github.com/bjornaer/crdt/internal/set"
)

type LastWriterWinsGraph interface {
	AddVertex(interface{}) error
	GetAllVertices() ([]interface{}, error)
	RemoveVertex(interface{}) error
	VertexExists(interface{}) bool
	AddEdge(v1, v2 interface{}) error
	RemoveEdge(v1, v2 interface{}) error
	EdgeExists(v1, v2 interface{}) bool
	GetVertexEdges(v interface{}) ([]interface{}, error)
	FindPath(v1, v2 interface{}) ([]interface{}, error)
	Merge(LastWriterWinsGraph) error
	getV() set.LastWriterWinsSet
	getE() map[interface{}]set.LastWriterWinsSet
}

// LWWGraph is a structure for a graph with vertices and edges based on LWW sets
type LWWGraph struct {
	vertices set.LastWriterWinsSet
	edges    map[interface{}]set.LastWriterWinsSet
	mutex    sync.RWMutex // Maps in Go are not thread safe by default and that's why we use a mutex
}

// NewLWWGraph returns an empty LWW based LWWGraph
func NewLWWGraph() LastWriterWinsGraph {
	return &LWWGraph{
		vertices: set.NewLWWSet(),
	}
}

// access private vertices
func (g *LWWGraph) getV() set.LastWriterWinsSet {
	return g.vertices
}

// access private edges
func (g *LWWGraph) getE() map[interface{}]set.LastWriterWinsSet {
	return g.edges
}

// AddVertex adds a vertex to the graph
func (g *LWWGraph) AddVertex(v interface{}) error {
	return g.vertices.Add(v, time.Now())
}

// GetAllVertices get all vertices from the LWWGraph
func (g *LWWGraph) GetAllVertices() ([]interface{}, error) {
	return g.vertices.Get()
}

// RemoveVertex removes a vertex from the LWWGraph
func (g *LWWGraph) RemoveVertex(v interface{}) error {
	return g.vertices.Remove(v, time.Now())
}

// VertexExists checks if a vertex is in the LWWGraph
func (g *LWWGraph) VertexExists(v interface{}) bool {
	return g.vertices.Exists(v)
}

// AddEdge adds an edge to the LWWGraph
func (g *LWWGraph) AddEdge(v1, v2 interface{}) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if !g.vertices.Exists(v1) {
		return fmt.Errorf("cannot add edge, missing node in graph: %v", v1)
	} else if !g.vertices.Exists(v2) {
		return fmt.Errorf("cannot add edge, missing node in graph: %v", v2)
	}

	if g.edges == nil {
		g.edges = make(map[interface{}]set.LastWriterWinsSet)
	}
	if _, ok := g.edges[v1]; !ok {
		g.edges[v1] = set.NewLWWSet()
	}
	err := g.edges[v1].Add(v2, time.Now())
	if err != nil {
		return err
	}
	if _, ok := g.edges[v2]; !ok {
		g.edges[v2] = set.NewLWWSet()
	}
	err = g.edges[v2].Add(v1, time.Now())
	if err != nil {
		return err
	}
	return nil
}

// RemoveEdge removes an edge from the LWWGraph
func (g *LWWGraph) RemoveEdge(v1, v2 interface{}) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if g.edges == nil {
		g.edges = make(map[interface{}]set.LastWriterWinsSet)
	}
	if _, ok := g.edges[v1]; !ok {
		g.edges[v1] = set.NewLWWSet()
	}
	err := g.edges[v1].Remove(v2, time.Now())
	if err != nil {
		return err
	}
	if _, ok := g.edges[v2]; !ok {
		g.edges[v2] = set.NewLWWSet()
	}
	err = g.edges[v2].Remove(v1, time.Now())
	if err != nil {
		return err
	}
	return nil
}

// EdgeExists checks if two vertices share an edge
func (g *LWWGraph) EdgeExists(v1, v2 interface{}) bool {
	return g.edges[v1].Exists(v2) && g.edges[v2].Exists(v1)
}

// GetVertexEdges allows querying for all vertices connected to a single vertex
func (g *LWWGraph) GetVertexEdges(v interface{}) ([]interface{}, error) {
	if !g.VertexExists(v) {
		return nil, errors.New("cannot query for edges, vertex does not exist")
	}
	return g.edges[v].Get()
}

// FindPath finds a connecting path between two given vertices
func (g *LWWGraph) FindPath(v1, v2 interface{}) ([]interface{}, error) {
	if !g.vertices.Exists(v1) {
		return nil, fmt.Errorf("cannot find path, missing node in graph: %v", v1)
	} else if !g.vertices.Exists(v2) {
		return nil, fmt.Errorf("cannot find path, missing node in graph: %v", v2)
	}

	seen := set.NewLWWSet()
	var emptyPath []interface{}
	_, path, err := g.findPathRecursive(v1, v2, seen, emptyPath)
	if err != nil {
		return nil, err
	}

	return path, nil
}

func (g *LWWGraph) findPathRecursive(
	v1,
	v2 interface{},
	seen set.LastWriterWinsSet,
	path []interface{}) (set.LastWriterWinsSet, []interface{}, error) {
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
func (g *LWWGraph) Merge(other LastWriterWinsGraph) error {
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
