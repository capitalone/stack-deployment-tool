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
// SPDX-Copyright: Copyright (c) Capital One Services, LLC
// SPDX-License-Identifier: Apache-2.0
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

	//fmt.Printf("%#v\n", dag.VertexListFromRoot())
	for i, v := range dag.VertexListFromRoot() {
		//fmt.Printf("[%d] = %#v\n", i, v)
		switch i {
		case 0:
			assert.Equal(t, v.Name, "blah")
		case 1:
			assert.Equal(t, v.Name, "99999")
		case 2:
			assert.Equal(t, v.Name, "C4")
		case 3:
			assert.Equal(t, v.Name, "alskdjf")
		case 4:
			assert.Equal(t, v.Name, "12341234")
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

	assert.Equal(t, dag.FindVertexByName("99999"), c3)
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

	fmt.Printf("%#v\n", dag.VertexListFromRoot())
	for i, v := range dag.VertexListFromRoot() {
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

func TestDagDFS(t *testing.T) {
	dag := NewDAG()

	r := "ROOT"
	s1 := "nagios-internal-dns"
	s2 := "nagios-r53"
	s3 := "nagios-elb"
	s4 := "nagios-server"
	s5 := "nagios-db"
	// r-> s1 -> s2  r->s3->s4->s5
	dag.AddRoot(&Vertex{Name: r})
	assert.NotNil(t, dag.AddEdgeBetweenVertices(r, s1))
	assert.NotNil(t, dag.AddEdgeBetweenVertices(s1, s2))

	assert.NotNil(t, dag.AddEdgeBetweenVertices(r, s3))
	assert.NotNil(t, dag.AddEdgeBetweenVertices(s3, s4))
	assert.NotNil(t, dag.AddEdgeBetweenVertices(s4, s5))

	index := 0
	dag.VisitDepthFirst(dag.Root, func(vert *Vertex) bool {
		switch index {
		case 0:
			assert.Equal(t, r, vert.Name)
		case 1:
			assert.Equal(t, s3, vert.Name)
		case 2:
			assert.Equal(t, s4, vert.Name)
		case 3:
			assert.Equal(t, s5, vert.Name)
		case 4:
			assert.Equal(t, s1, vert.Name)
		case 5:
			assert.Equal(t, s2, vert.Name)
		default:
			assert.True(t, false) // shouldnt get here
		}
		index++
		return true
	})
}

func TestSimpleDag(t *testing.T) {
	dag := NewDAG()
	r := "ROOT"
	s1 := "nagios-internal-dns"
	s2 := "nagios-r53"
	s3 := "nagios-elb"
	s4 := "nagios-server"
	/*
		vr := &Vertex{Name: r}
		v1 := &Vertex{Name: s1}
		v2 := &Vertex{Name: s2}
		v3 := &Vertex{Name: s3}
		v4 := &Vertex{Name: s4}
		dag.AddRoot(vr)
		dag.AddEdge(&Edge{Parent: vr, Child: v1})
		dag.AddEdge(&Edge{Parent: v1, Child: v2})
		dag.AddEdge(&Edge{Parent: vr, Child: v3})

		assert.Equal(t, dag.FindVertexByName(s3), v3)
		dag.AddEdge(&Edge{Parent: v3, Child: v4})
	*/
	// r-> s1 -> s2  r->s3->s4
	dag.AddRoot(&Vertex{Name: r})
	assert.NotNil(t, dag.AddEdgeBetweenVertices(r, s1))
	assert.NotNil(t, dag.AddEdgeBetweenVertices(s1, s2))

	assert.NotNil(t, dag.AddEdgeBetweenVertices(r, s3))
	assert.NotNil(t, dag.AddEdgeBetweenVertices(s3, s4))

	for i, v := range dag.VertexListFromRoot() {
		//fmt.Printf("[%d] = %#v\n", i, v)
		switch i {
		case 0:
			assert.Equal(t, v.Name, "ROOT")
		case 1:
			assert.Equal(t, v.Name, "nagios-elb")
		case 2:
			assert.Equal(t, v.Name, "nagios-server")
		case 3:
			assert.Equal(t, v.Name, "nagios-internal-dns")
		case 4:
			assert.Equal(t, v.Name, "nagios-r53")
		default:
			assert.True(t, false) // shouldnt get here
		}
	}
	// for k, v := range dag.Vertices {
	// 	fmt.Printf("k[%s] = %#v\n", k, v)
	// 	for _, e := range v {
	// 		fmt.Printf(" k[%s] = %#v -> %#v\n", k, e.Parent.Name, e.Child.Name)
	// 	}
	// }
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
	fmt.Printf("%#v\n", dag.VertexListFromRoot())

	for i, v := range dag.VertexListFromRoot() {
		fmt.Printf("[%d] = %#v\n", i, v)
		switch i {
		case 0:
			assert.Equal(t, v.Name, "A")
		case 1:
			assert.Equal(t, v.Name, "C")
		case 2:
			assert.Equal(t, v.Name, "D")
		case 3:
			assert.Equal(t, v.Name, "E")
		case 4:
			assert.Equal(t, v.Name, "B")
		default:
			assert.True(t, false) // shouldnt get here
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
			assert.True(t, false) // shouldnt get here
		}
		vertIdx++
	})

	fmt.Printf("------------\n")
}
