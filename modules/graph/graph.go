package graph

import (
	"encoding/json"
	"fmt"
	"path"
	"sync"
	"time"

	"github.com/bettercap/bettercap/v2/session"
	"github.com/evilsocket/islazy/fs"
)

var Loaded = (*Graph)(nil)

type NodeCallback func(*Node)
type EdgeCallback func(*Node, []Edge, *Node)

type Graph struct {
	sync.Mutex

	path  string
	edges *Edges
}

func NewGraph(path string) (*Graph, error) {
	if edges, err := LoadEdges(path); err != nil {
		return nil, err
	} else {
		Loaded = &Graph{
			path:  path,
			edges: edges,
		}
		return Loaded, nil
	}
}

func (g *Graph) EachNode(cb NodeCallback) error {
	g.Lock()
	defer g.Unlock()

	for _, nodeType := range NodeTypes {
		err := fs.Glob(g.path, fmt.Sprintf("%s_*.json", nodeType), func(fileName string) error {
			if node, err := ReadNode(fileName); err != nil {
				return err
			} else {
				cb(node)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *Graph) EachEdge(cb EdgeCallback) error {
	g.Lock()
	defer g.Unlock()

	return g.edges.ForEachEdge(func(fromID string, edges []Edge, toID string) error {
		var left, right *Node
		var err error

		leftFileName := path.Join(g.path, fromID+".json")
		rightFileName := path.Join(g.path, toID+".json")

		if left, err = ReadNode(leftFileName); err != nil {
			return err
		} else if right, err = ReadNode(rightFileName); err != nil {
			return err
		}

		cb(left, edges, right)

		return nil
	})
}

func (g *Graph) Traverse(root string, onNode NodeCallback, onEdge EdgeCallback) error {
	if root == "" {
		// traverse the entire graph
		if err := g.EachNode(onNode); err != nil {
			return err
		} else if err = g.EachEdge(onEdge); err != nil {
			return err
		}
	} else {
		// start by a specific node
		roots, err := g.FindOtherTypes("", root)
		if err != nil {
			return err
		}

		stack := NewStack()
		for _, root := range roots {
			stack.Push(root)
		}

		type edgeBucket struct {
			left  *Node
			edges []Edge
			right *Node
		}

		allEdges := make([]edgeBucket, 0)
		visited := make(map[string]bool)

		for {
			if last := stack.Pop(); last == nil {
				break
			} else {
				node := last.(*Node)
				nodeID := node.String()
				if _, found := visited[nodeID]; found {
					continue
				} else {
					visited[nodeID] = true
				}

				onNode(node)

				// collect all edges starting from this node
				err = g.edges.ForEachEdgeFrom(nodeID, func(_ string, edges []Edge, toID string) error {
					rightFileName := path.Join(g.path, toID+".json")
					if right, err := ReadNode(rightFileName); err != nil {
						return err
					} else {
						// collect new node
						if _, found := visited[toID]; !found {
							stack.Push(right)
						}
						// collect all edges, we'll emit this later
						allEdges = append(allEdges, edgeBucket{
							left:  node,
							edges: edges,
							right: right,
						})
					}
					return nil
				})
			}
		}

		for _, edge := range allEdges {
			onEdge(edge.left, edge.edges, edge.right)
		}
	}

	return nil
}

func (g *Graph) IsConnected(nodeType string, nodeID string) bool {
	return g.edges.IsConnected(fmt.Sprintf("%s_%s", nodeType, nodeID))
}

func (g *Graph) Dot(filter, layout, name string, disconnected bool) (string, int, int, error) {
	size := 0
	discarded := 0

	data := fmt.Sprintf("digraph %s {\n", name)
	data += fmt.Sprintf("  layout=%s\n", layout)

	typeMap := make(map[NodeType]bool)

	type typeCount struct {
		edge  Edge
		count int
	}

	if err := g.Traverse(filter, func(node *Node) {
		include := false
		if disconnected {
			include = true
		} else {
			include = g.edges.IsConnected(node.String())
		}

		if include {
			size++
			typeMap[node.Type] = true
			data += fmt.Sprintf("  %s\n", node.Dot(filter == node.ID))
		} else {
			discarded++
		}
	}, func(left *Node, edges []Edge, right *Node) {
		// collect counters by edge type in order to calculate proportional widths
		byType := make(map[string]typeCount)
		tot := len(edges)

		for _, edge := range edges {
			if c, found := byType[string(edge.Type)]; found {
				c.count++
			} else {
				byType[string(edge.Type)] = typeCount{
					edge:  edge,
					count: 1,
				}
			}
		}

		max := 2.0
		for _, c := range byType {
			w := max * float64(c.count/tot)
			if w < 0.5 {
				w = 0.5
			}
			data += fmt.Sprintf("  %s\n", c.edge.Dot(left, right, w))
		}
	}); err != nil {
		return "", 0, 0, err
	}

	/*
		data += "\n"
		data += "node [style=filled height=0.55 fontname=\"Verdana\" fontsize=10];\n"
		data += "subgraph legend {\n" +
				"graph[style=dotted];\n" +
				"label = \"Legend\";\n"

		var types []NodeType
		for nodeType, _ := range typeMap {
			types = append(types, nodeType)
			node := Node{
				Type:        nodeType,
				Annotations: nodeTypeDescs[nodeType],
				Dummy:       true,
			}
			data += fmt.Sprintf("  %s\n", node.Dot(false))
		}

		ntypes := len(types)
		for i := 0; i < ntypes - 1; i++ {
			data += fmt.Sprintf("  \"%s\" -> \"%s\" [style=invis];\n", types[i], types[i + 1])
		}
		data += "}\n"
	*/

	data += "\n"
	data += "  overlap=false\n"
	data += "}"

	return data, size, discarded, nil
}

func (g *Graph) JSON(filter string, disconnected bool) (string, int, int, error) {
	size := 0
	discarded := 0

	type link struct {
		Source string      `json:"source"`
		Target string      `json:"target"`
		Edge   interface{} `json:"edge"`
	}

	type data struct {
		Nodes []map[string]interface{} `json:"nodes"`
		Links []link                   `json:"links"`
	}

	jsData := data{
		Nodes: make([]map[string]interface{}, 0),
		Links: make([]link, 0),
	}

	if err := g.Traverse(filter, func(node *Node) {
		include := false
		if disconnected {
			include = true
		} else {
			include = g.edges.IsConnected(node.String())
		}

		if include {
			size++

			if nm, err := node.ToMap(); err != nil {
				panic(err)
			} else {
				// patch id
				nm["id"] = node.String()
				jsData.Nodes = append(jsData.Nodes, nm)
			}
		} else {
			discarded++
		}
	}, func(left *Node, edges []Edge, right *Node) {
		for _, edge := range edges {
			jsData.Links = append(jsData.Links, link{
				Source: left.String(),
				Target: right.String(),
				Edge:   edge,
			})
		}
	}); err != nil {
		return "", 0, 0, err
	}

	if raw, err := json.Marshal(jsData); err != nil {
		return "", 0, 0, err
	} else {
		return string(raw), size, discarded, nil
	}
}

func (g *Graph) FindNode(t NodeType, id string) (*Node, error) {
	g.Lock()
	defer g.Unlock()

	nodeFileName := path.Join(g.path, fmt.Sprintf("%s_%s.json", t, id))
	if fs.Exists(nodeFileName) {
		return ReadNode(nodeFileName)
	}

	return nil, nil
}

func (g *Graph) FindOtherTypes(t NodeType, id string) ([]*Node, error) {
	g.Lock()
	defer g.Unlock()

	var otherNodes []*Node

	for _, otherType := range NodeTypes {
		if otherType != t {
			if nodeFileName := path.Join(g.path, fmt.Sprintf("%s_%s.json", otherType, id)); fs.Exists(nodeFileName) {
				if node, err := ReadNode(nodeFileName); err != nil {
					return nil, err
				} else {
					otherNodes = append(otherNodes, node)
				}
			}
		}
	}

	return otherNodes, nil
}

func (g *Graph) CreateNode(t NodeType, id string, entity interface{}, annotations string) (*Node, error) {
	g.Lock()
	defer g.Unlock()

	node := &Node{
		Type:        t,
		ID:          id,
		Entity:      entity,
		Annotations: annotations,
	}

	nodeFileName := path.Join(g.path, fmt.Sprintf("%s.json", node.String()))
	if err := CreateNode(nodeFileName, node); err != nil {
		return nil, err
	}

	session.I.Events.Add("graph.node.new", node)

	return node, nil
}

func (g *Graph) UpdateNode(node *Node) error {
	g.Lock()
	defer g.Unlock()

	nodeFileName := path.Join(g.path, fmt.Sprintf("%s.json", node.String()))
	if err := UpdateNode(nodeFileName, node); err != nil {
		return err
	}

	return nil
}

func (g *Graph) FindLastEdgeOfType(from, to *Node, edgeType EdgeType) (*Edge, error) {
	edges := g.edges.FindEdges(from.String(), to.String(), true)
	num := len(edges)
	for i := range edges {
		// loop backwards
		idx := num - 1 - i
		edge := edges[idx]
		if edge.Type == edgeType {
			return &edge, nil
		}
	}
	return nil, nil
}

func (g *Graph) FindLastRecentEdgeOfType(from, to *Node, edgeType EdgeType, staleTime time.Duration) (*Edge, error) {
	edges := g.edges.FindEdges(from.String(), to.String(), true)
	num := len(edges)
	for i := range edges {
		// loop backwards
		idx := num - 1 - i
		edge := edges[idx]
		if edge.Type == edgeType {
			if time.Since(edge.CreatedAt) >= staleTime {
				return nil, nil
			}
			return &edge, nil
		}
	}

	return nil, nil
}

func (g *Graph) CreateEdge(from, to *Node, edgeType EdgeType) (*Edge, error) {
	edge := Edge{
		Type:      edgeType,
		CreatedAt: time.Now(),
	}

	if session.I.GPS.Updated.IsZero() == false {
		edge.Position = &session.I.GPS
	}

	if err := g.edges.Connect(from.String(), to.String(), edge); err != nil {
		return nil, err
	}

	session.I.Events.Add("graph.edge.new", EdgeEvent{
		Left:  from,
		Edge:  &edge,
		Right: to,
	})

	return &edge, nil
}
