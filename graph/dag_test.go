//
// Copyright 2016 Capital One Services, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and limitations under the License.
//
package graph

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDagCreation(t *testing.T) {
	dag := NewDAG()
	r := &Vertex{Name: "blah"}
	c := &Vertex{Name: "alskdjf"}
	c2 := &Vertex{Name: "12341234"}
	c3 := &Vertex{Name: "99999"}
	c4 := &Vertex{Name: "C4"}
	dag.AddRoot(r)
	dag.AddEdge(&Edge{Parent: r, Child: c})
	dag.AddEdge(&Edge{Parent: c, Child: c2})
	dag.AddEdge(&Edge{Parent: c2, Child: c3})
	dag.AddEdge(&Edge{Parent: r, Child: c3})
	dag.AddEdge(&Edge{Parent: c3, Child: c4})

	//fmt.Printf("%#v\n", dag.VertexList())
	for i, v := range dag.VertexList() {
		//fmt.Printf("[%d] = %#v\n", i, v)
		switch i {
		case 0:
			assert.Equal(t, v.Name, "blah")
		case 1:
			assert.Equal(t, v.Name, "alskdjf")
		case 2:
			assert.Equal(t, v.Name, "99999")
		case 3:
			assert.Equal(t, v.Name, "12341234")
		case 4:
			assert.Equal(t, v.Name, "C4")
		default:
			assert.True(t, false)
		}
	}

	// fmt.Printf("------------\n")
	// fmt.Printf("%#v -> %#v\n", r, c)
	// fmt.Printf("%#v -> %#v\n", c, c2)
	// fmt.Printf("%#v -> %#v\n", c2, c3)
	// fmt.Printf("%#v -> %#v\n", r, c3)
	// fmt.Printf("%#v -> %#v\n", c3, c4)
	// fmt.Printf("------------\n")
	dag.Print(os.Stdout)
}

func TestDagTransitiveReduction(t *testing.T) {
	dag := NewDAG()
	r := &Vertex{Name: "blah"}
	c := &Vertex{Name: "alskdjf"}
	c2 := &Vertex{Name: "12341234"}
	c3 := &Vertex{Name: "99999"}
	c4 := &Vertex{Name: "C4"}
	dag.AddRoot(r)
	dag.AddEdge(&Edge{Parent: r, Child: c})
	dag.AddEdge(&Edge{Parent: c, Child: c2})
	dag.AddEdge(&Edge{Parent: c2, Child: c3})
	dag.AddEdge(&Edge{Parent: r, Child: c3}) // extra edge
	dag.AddEdge(&Edge{Parent: c3, Child: c4})

	assert.Equal(t, dag.Edges.Len(), 5)
	dag.TransitiveReduction()
	assert.Equal(t, dag.Edges.Len(), 4)
	dag.VisitEdges(func(edge *Edge) {
		assert.False(t, edge.Parent == r && edge.Child == c3)
	})
	dag.Print(os.Stdout)

	fmt.Printf("%#v\n", dag.VertexList())
	for i, v := range dag.VertexList() {
		fmt.Printf("[%d] = %#v\n", i, v)
	}

	// fmt.Printf("------------\n")
	// fmt.Printf("%#v -> %#v\n", r, c)
	// fmt.Printf("%#v -> %#v\n", c, c2)
	// fmt.Printf("%#v -> %#v\n", c2, c3)
	// fmt.Printf("%#v -> %#v\n", r, c3)
	// fmt.Printf("%#v -> %#v\n", c3, c4)
	// fmt.Printf("------------\n")
}

func TestDagCycleCheck1(t *testing.T) {
	dag := NewDAG()
	a := &Vertex{Name: "A"}
	b := &Vertex{Name: "B"}
	c := &Vertex{Name: "C"}
	d := &Vertex{Name: "D"}
	e := &Vertex{Name: "E"}

	dag.AddRoot(a)
	// A->B->C->D  A->D->E  C->A
	dag.AddEdge(&Edge{Parent: a, Child: b})
	dag.AddEdge(&Edge{Parent: a, Child: d})
	dag.AddEdge(&Edge{Parent: b, Child: c})
	dag.AddEdge(&Edge{Parent: c, Child: d})
	dag.AddEdge(&Edge{Parent: c, Child: a})
	dag.AddEdge(&Edge{Parent: d, Child: e})

	assert.True(t, dag.HasCycles())
}

func TestDagCycleCheck2(t *testing.T) {
	dag := NewDAG()
	a := &Vertex{Name: "0"}
	b := &Vertex{Name: "1"}
	c := &Vertex{Name: "2"}
	d := &Vertex{Name: "3"}

	dag.AddRoot(a)

	// 0->1->2  0->2  2->0  2->3   3->3
	dag.AddEdge(&Edge{Parent: a, Child: b})
	dag.AddEdge(&Edge{Parent: a, Child: c})

	dag.AddEdge(&Edge{Parent: b, Child: c})

	dag.AddEdge(&Edge{Parent: c, Child: a})
	dag.AddEdge(&Edge{Parent: c, Child: d})

	dag.AddEdge(&Edge{Parent: d, Child: d})

	assert.True(t, dag.HasCycles())
}

func TestDagExampleTransitiveReduction(t *testing.T) {
	dag := NewDAG()
	a := &Vertex{Name: "A"}
	b := &Vertex{Name: "B"}
	c := &Vertex{Name: "C"}
	d := &Vertex{Name: "D"}
	e := &Vertex{Name: "E"}

	dag.AddRoot(a)
	dag.AddEdge(&Edge{Parent: a, Child: b})
	dag.AddEdge(&Edge{Parent: a, Child: d})
	dag.AddEdge(&Edge{Parent: a, Child: c})
	dag.AddEdge(&Edge{Parent: a, Child: e})

	dag.AddEdge(&Edge{Parent: b, Child: d})

	dag.AddEdge(&Edge{Parent: d, Child: e})

	dag.AddEdge(&Edge{Parent: c, Child: d})
	dag.AddEdge(&Edge{Parent: c, Child: e})

	dag.Print(os.Stdout)
	dag.TransitiveReduction()
	fmt.Printf("------------\n")
	dag.Print(os.Stdout)
	fmt.Printf("------------\n")
	fmt.Printf("%#v\n", dag.VertexList())

	for i, v := range dag.VertexList() {
		fmt.Printf("[%d] = %#v\n", i, v)
		switch i {
		case 0:
			assert.Equal(t, v.Name, "A")
		case 1:
			assert.Equal(t, v.Name, "B")
		case 2:
			assert.Equal(t, v.Name, "C")
		case 3:
			assert.Equal(t, v.Name, "D")
		case 4:
			assert.Equal(t, v.Name, "E")
		default:
			assert.True(t, false)
		}
	}
	/*
	   el[0]: &graph.Vertex{Name:"A"} - &graph.Vertex{Name:"B"}
	   el[1]: &graph.Vertex{Name:"A"} - &graph.Vertex{Name:"C"}
	   el[2]: &graph.Vertex{Name:"B"} - &graph.Vertex{Name:"D"}
	   el[3]: &graph.Vertex{Name:"D"} - &graph.Vertex{Name:"E"}
	   el[4]: &graph.Vertex{Name:"C"} - &graph.Vertex{Name:"D"}
	*/
	vertIdx := 0
	dag.VisitEdges(func(edge *Edge) {
		fmt.Printf("el[%d]: %#v - %#v\n", vertIdx, edge.Parent, edge.Child)
		switch vertIdx {
		case 0:
			assert.Equal(t, edge.Parent.Name, "A")
			assert.Equal(t, edge.Child.Name, "B")
		case 1:
			assert.Equal(t, edge.Parent.Name, "A")
			assert.Equal(t, edge.Child.Name, "C")
		case 2:
			assert.Equal(t, edge.Parent.Name, "B")
			assert.Equal(t, edge.Child.Name, "D")
		case 3:
			assert.Equal(t, edge.Parent.Name, "D")
			assert.Equal(t, edge.Child.Name, "E")
		case 4:
			assert.Equal(t, edge.Parent.Name, "C")
			assert.Equal(t, edge.Child.Name, "D")
		default:
			assert.True(t, false)
		}
		vertIdx++
	})

	fmt.Printf("------------\n")
}
