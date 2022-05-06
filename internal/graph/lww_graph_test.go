package graph_test

// We add the test file in a separate package to keep testing and actual implementation details separate
// By doing so, in the tests we will only have access to the public part of our code

import (
	"testing"

	graph "github.com/bjornaer/crdt/internal/graph"
)

func setupTestGraph() graph.LastWriterWinsGraph {
	v1 := "vertex1"
	v2 := "vertex2"
	v3 := "vertex3"

	g := graph.NewLWWGraph()
	g.AddVertex(v1)
	g.AddVertex(v2)
	g.AddVertex(v3)
	g.AddEdge(v1, v2)
	g.AddEdge(v2, v3)
	return g
}

// compares generic but ordered interface slice to string slice - used for the findPath test cases purposes
func pathsAreEqual(s1 []interface{}, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i, v := range s1 {
		if v != s2[i] {
			return false
		}
	}
	return true
}

// checks element is contained within set
func contains(s []interface{}, e interface{}) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// checks for set equality -- independent of order
func setsAreEqual(s1, s2 []interface{}) bool {
	if len(s1) != len(s2) {
		return false
	}
	for _, v := range s1 {
		if !contains(s2, v) {
			return false
		}
	}
	return true
}

func TestLWWGraph_AddVertex(t *testing.T) {
	g := setupTestGraph()
	v := "new_vertex"
	err := g.AddVertex(v)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	vertices, err := g.GetAllVertices()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !g.VertexExists(v) {
		t.Errorf("Missing vertex, got: %v, expected: %v.", vertices, append(vertices, v))
	}
}

func TestLWWGraph_GetAllVertices(t *testing.T) {
	g := setupTestGraph()
	vertices, err := g.GetAllVertices()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	expected := []interface{}{"vertex1", "vertex2", "vertex3"}
	if !setsAreEqual(vertices, expected) {
		t.Errorf("Vertices mismatch, got: %v, expected: %v.", vertices, expected)
	}
}

func TestLWWGraph_RemoveVertex(t *testing.T) {
	g := setupTestGraph()
	v := "vertex3"
	err := g.RemoveVertex(v)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	vertices, err := g.GetAllVertices()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if g.VertexExists(v) {
		t.Errorf("Extra vertex found, got: %v, expected: %v.", vertices, []string{"vertex1", "vertex2"})
	}
}

func TestLWWGraph_VertexExists(t *testing.T) {
	g := setupTestGraph()
	tests := []struct {
		vertex   string
		expected bool
	}{
		{"vertex1", true},
		{"inexistent_vertex", false},
	}
	for _, tt := range tests {
		got := g.VertexExists(tt.vertex)
		if got != tt.expected {
			t.Errorf("Existence check failed, got: %v, expected: %v.", got, tt.expected)
		}
	}
}

func TestLWWGraph_AddEdge(t *testing.T) {
	g := setupTestGraph()
	v1 := "vertex1"
	v3 := "vertex3"
	err := g.AddEdge(v1, v3)
	edges1, _ := g.GetVertexEdges(v1)
	edges3, _ := g.GetVertexEdges(v3)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !g.EdgeExists(v1, v3) {
		t.Errorf("Missing edge, got:[ %v, %v ], expected: [ %v, %v ].", edges1, edges3, []string{"vertex2", "vertex3"}, []string{"vertex1", "vertex2"})
	}
}

func TestLWWGraph_RemoveEdge(t *testing.T) {
	g := setupTestGraph()
	v1 := "vertex1"
	v2 := "vertex2"
	err := g.RemoveEdge(v1, v2)
	edges1, _ := g.GetVertexEdges(v1)
	edges2, _ := g.GetVertexEdges(v2)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if g.EdgeExists(v1, v2) {
		t.Errorf("Extra edge found, got:[ %v, %v ], expected: [ %v, %v ].", edges1, edges2, []string{}, []string{"vertex3"})
	}
}

func TestLWWGraph_EdgeExists(t *testing.T) {
	g := setupTestGraph()
	tests := []struct {
		vertices []string
		expected bool
	}{
		{[]string{"vertex1", "vertex2"}, true},
		{[]string{"vertex1", "inexistent_vertex"}, false},
	}
	for _, tt := range tests {
		got := g.EdgeExists(tt.vertices[0], tt.vertices[1])
		if got != tt.expected {
			t.Errorf("Existence check failed, got: %v, expected: %v.", got, tt.expected)
		}
	}
}

func TestLWWGraph_GetVertexEdges(t *testing.T) {
	g := setupTestGraph()
	v := "vertex1"
	edges, err := g.GetVertexEdges(v)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if edges[0] != "vertex2" {
		t.Errorf("Extra edge found, got: %v, expected: %v.", edges, []string{"vertex2"})
	}
}

func TestLWWGraph_FindPath(t *testing.T) {
	simpleG := setupTestGraph()
	cycleG := setupTestGraph()
	cycleG.AddVertex("vertex4")
	cycleG.AddEdge("vertex3", "vertex4")
	disconnectedG := setupTestGraph()
	disconnectedG.RemoveEdge("vertex1", "vertex2")
	tests := []struct {
		name     string
		graph    graph.LastWriterWinsGraph
		from     string
		target   string
		expected []string
	}{
		{"simple graph", simpleG, "vertex1", "vertex3", []string{"vertex1", "vertex2", "vertex3"}},
		{"cycle graph", cycleG, "vertex1", "vertex4", []string{"vertex1", "vertex2", "vertex3", "vertex4"}},
		{"disconnected graph", disconnectedG, "vertex1", "vertex3", []string{"vertex1"}},
	}
	for _, tt := range tests {
		got, err := tt.graph.FindPath(tt.from, tt.target)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if !pathsAreEqual(got, tt.expected) {
			t.Errorf("Finding path for %s failed, got: %v, expected: %v.", tt.name, got, tt.expected)
		}
	}
}

func TestLWWGraph_Merge(t *testing.T) {
	g1 := setupTestGraph()
	g2 := setupTestGraph()
	v1 := "vertex1"
	v2 := "vertex2"
	v3 := "vertex3"
	nv := "new_vertex"
	g2.RemoveVertex(v2)
	g1.AddVertex(v2)
	g2.AddEdge(v1, v3)
	g2.AddVertex(nv)
	g1.RemoveVertex(nv)
	err := g1.Merge(g2)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !g1.VertexExists(v2) || g1.VertexExists(nv) {
		t.Errorf("Vertices merge failed")
	}
	if !g1.EdgeExists(v1, v3) {
		t.Errorf("Edges merge failed")
	}
}

// testing associativity behavior when merging graphs -- we'll check vertices to keep it simple
// ie: g1 v (g2 v g3) = (g1 v g2) v g3
func TestLWWGraph_Associativity(t *testing.T) {
	// "g1 v (g2 v g3)"
	g1 := setupTestGraph()
	g2 := setupTestGraph()
	g3 := setupTestGraph()
	g1.AddVertex("vertex4")
	g3.RemoveVertex("vertex4")
	g2.Merge(g3)
	g1.Merge(g2)
	// "(g1 v g2) v g3"
	// where "g4 ~ g1", "g5 ~ g2", "g6 ~ g3" (because timestamps are not equal)
	g4 := setupTestGraph()
	g5 := setupTestGraph()
	g6 := setupTestGraph()
	g4.AddVertex("vertex4")
	g6.RemoveVertex("vertex4")
	g4.Merge(g5)
	g4.Merge(g6)

	// g1 and g4 hold the results of the union operations, they should contain equal vertices elements
	// although timestamps mismatch the order of operations is equal, given the underlying monotonicity
	// we should end up with equal vertices on both graphs
	v1, _ := g1.GetAllVertices()
	v4, _ := g4.GetAllVertices()
	if !setsAreEqual(v1, v4) {
		t.Errorf("Merge not associative, g1 v (g2 v g3): %v, (g1 v g2) v g3: %v.", v1, v4)
	}
}

// testing commutativity behavior when merging graphs -- we'll check vertices to keep it simple
// // ie: g1 v g2 = g2 v g1
func TestLWWGraph_Commutativity(t *testing.T) {
	// "g1 v g2"
	g1 := setupTestGraph()
	g2 := setupTestGraph()
	g1.AddVertex("vertex4")
	g2.RemoveVertex("vertex4")
	g1.Merge(g2)
	// "g2 v g1"
	// where "g4 ~ g1", "g5 ~ g2" (because timestamps are not equal)
	g4 := setupTestGraph()
	g5 := setupTestGraph()
	g4.AddVertex("vertex4")
	g5.RemoveVertex("vertex4")
	g5.Merge(g4)

	// g1 and g5 hold the results of the union operations, they should contain equal vertices elements
	// although timestamps mismatch the order of operations is equal, given the underlying monotonicity
	// we should end up with equal vertices on both graphs
	v1, _ := g1.GetAllVertices()
	v5, _ := g5.GetAllVertices()
	if !setsAreEqual(v1, v5) {
		t.Errorf("Merge not associative, g1 v g2: %v, g2 v g1: %v.", v1, v5)
	}
}

// testing idempotence behavior when merging graphs -- we'll check vertices to keep it simple
// // ie: g1 v g1 = g1
func TestLWWGraph_Idempotence(t *testing.T) {
	// g and gCopy act as the same structure (g1) modified in the same manner in two instances separately
	g := setupTestGraph()
	gCopy := setupTestGraph()
	g.AddVertex("vertex4")
	g.RemoveVertex("vertex3")
	gCopy.AddVertex("vertex4")
	gCopy.RemoveVertex("vertex3")

	beforeVertices, _ := gCopy.GetAllVertices()
	g.Merge(gCopy)
	afterVertices, _ := g.GetAllVertices()
	if !setsAreEqual(beforeVertices, afterVertices) {
		t.Errorf("Merge not idempotent, g1: %v, g1 v g1: %v.", beforeVertices, afterVertices)
	}
}
