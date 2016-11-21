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
	"container/list"
	"fmt"
	"io"
)

type Vertex struct {
	Name string
}

type Edge struct {
	Parent *Vertex
	Child  *Vertex
}

type DAG struct {
	Root     *Vertex
	Edges    *list.List
	Vertices map[*Vertex][]*Edge
}

func NewDAG() *DAG {
	return &DAG{Vertices: make(map[*Vertex][]*Edge), Edges: list.New()}
}

func (d *DAG) AddRoot(v *Vertex) {
	d.Root = v
	d.AddVertex(v)
	// check for duplicates
}

func (d *DAG) AddVertex(v *Vertex) {
	if _, ok := d.Vertices[v]; !ok {
		d.Vertices[v] = []*Edge{}
	}
}

func (d *DAG) AddEdge(e *Edge) {
	if e.Parent == nil && d.Root != nil {
		e.Parent = d.Root
	}
	d.Edges.PushBack(e)
	if !d.VertexExists(e.Parent) {
		d.AddVertex(e.Parent)
	}
	if !d.VertexExists(e.Child) {
		d.AddVertex(e.Child)
	}
	if edges, ok := d.Vertices[e.Parent]; ok {
		d.Vertices[e.Parent] = append(edges, e)
	}
}

func (d *DAG) FindVertexByName(name string) *Vertex {
	for k, _ := range d.Vertices {
		if k.Name == name {
			return k
		}
	}
	return nil
}

func (d *DAG) AddEdgeBetweenVertices(parentName, childName string) *Edge {
	p := d.FindVertexByName(parentName)
	if p == nil {
		p = &Vertex{Name: parentName}
	}

	c := d.FindVertexByName(childName)
	if c == nil {
		c = &Vertex{Name: childName}
	}

	e := &Edge{Parent: p, Child: c}
	d.AddEdge(e)
	return e
}

func (d *DAG) VertexExists(v *Vertex) bool {
	if _, ok := d.Vertices[v]; ok {
		return true
	}
	return false
}

func (d *DAG) VertexListFromRoot() []*Vertex {
	return d.VertexList(d.Root)
}

func (d *DAG) VertexList(start *Vertex) []*Vertex {
	result := []*Vertex{}
	if start != nil {
		d.VisitDepthFirst(start, func(vert *Vertex) bool {
			result = append(result, vert)
			return true
		})
	}

	return result
}

// VertexVisitorFunc visits each vertex, only once
// and returns false to stop visiting, or else true to continue
type VertexVisitorFunc func(vertex *Vertex) bool

func (d *DAG) VisitDepthFirstFromRoot(visitor VertexVisitorFunc) {
	d.VisitDepthFirst(d.Root, visitor)
}

func (d *DAG) VisitDepthFirst(startVertex *Vertex, visitor VertexVisitorFunc) {
	visited := map[*Vertex]bool{}
	result := []*Vertex{}
	start := startVertex

	if start != nil {
		downVerts := list.New()
		downVerts.PushBack(start)
		for el := downVerts.Front(); el != nil; el = el.Next() {
			vert := castVertex(el)
			if _, found := visited[vert]; !found {
				result = append(result, vert)
				visited[vert] = true

				cont := visitor(vert)
				if !cont {
					return
				}
				edges := d.Vertices[vert]
				for _, edge := range edges {
					downVerts.PushBack(edge.Child)
				}
			}
		}
	}
}

func (d *DAG) Print(w io.Writer) {
	start := d.Root
	downVerts := list.New()
	downVerts.PushBack(start)
	for el := downVerts.Front(); el != nil; el = el.Next() {
		vert := castVertex(el)
		edges := d.Vertices[vert]
		for _, edge := range edges {
			downVerts.PushBack(edge.Child)
			fmt.Fprintf(w, "%#v -> %#v\n", edge.Parent, edge.Child)
		}
	}
}

func (d *DAG) RemoveEdge(e *Edge) {
	if edges, ok := d.Vertices[e.Parent]; ok {
		rmIndex := -1
		for index, edge := range edges {
			if edge == e {
				rmIndex = index
			}
		}
		if rmIndex >= 0 {
			edges = append(edges[:rmIndex], edges[rmIndex+1:]...)
		}
		d.Vertices[e.Parent] = edges
	}
	rmEdges := []*list.Element{}
	for el := d.Edges.Front(); el != nil; el = el.Next() {
		edge := castEdge(el)
		if edge == e {
			rmEdges = append(rmEdges, el)
		}
	}
	for _, el := range rmEdges {
		d.Edges.Remove(el)
	}
}

func arrayContainsVertex(verts []*Vertex, vert *Vertex) bool {
	for _, v := range verts {
		if v == vert {
			return true
		}
	}
	return false
}

func (d *DAG) TransitiveReduction() {
	downVerts := list.New()
	downVerts.PushBack(d.Root)
	for el := downVerts.Front(); el != nil; el = el.Next() {
		vert := castVertex(el)

		// get edges for this vertex
		edges := d.Vertices[vert]
		lowerEdges := []*Edge{}
		for _, edge := range edges {
			// get the list of edges below each child of this vertex
			lowerEdges = append(lowerEdges, d.Vertices[edge.Child]...)
		}
		// A -> B -> C & A -> C, edges: A->B, lower edge: B->C

		// check if any of the other vertices in adjacenet edges exist below the vertex being examined
		for _, lowerEdge := range lowerEdges {
			verts := d.VertexList(lowerEdge.Child)
			for _, upperEdge := range edges {
				if arrayContainsVertex(verts, upperEdge.Child) {
					d.RemoveEdge(upperEdge)
				}
			}
		}

		edges = d.Vertices[vert]
		for _, edge := range edges {
			downVerts.PushBack(edge.Child)
		}
	}
}

func (d *DAG) HasCycles() bool {

	edges := map[string]bool{}
	vertsVisited := map[*Vertex]bool{}

	cycleFound := false
	d.VisitDepthFirstFromRoot(func(vertex *Vertex) bool {
		vertEdges := d.Vertices[vertex]
		vertsVisited[vertex] = true
		for _, edge := range vertEdges {
			id := fmt.Sprintf("%v%v", edge.Parent, edge.Child)

			if _, found := vertsVisited[edge.Child]; found {
				cycleFound = true
				return false
			}

			if _, found := edges[id]; !found {
				edges[id] = true
			} else {
				cycleFound = true
				return false
			}
		}
		return true
	})
	return cycleFound
}

func castEdge(el *list.Element) *Edge {
	return el.Value.(*Edge)
}

func castVertex(el *list.Element) *Vertex {
	return el.Value.(*Vertex)
}

type EdgeVistorFunc func(edge *Edge)

func (d *DAG) VisitEdges(visitor EdgeVistorFunc) {
	for el := d.Edges.Front(); el != nil; el = el.Next() {
		edge := castEdge(el)
		visitor(edge)
	}
}

func (e *Edge) String() string {
	return fmt.Sprintf("%s -> %s", e.Parent.Name, e.Child.Name)
}
func (v *Vertex) String() string {
	return fmt.Sprintf("%s", v.Name)
}
