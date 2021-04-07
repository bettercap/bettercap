package graph

import (
	"encoding/json"
	"fmt"
	"strings"
	"github.com/bettercap/bettercap/session"
	"github.com/evilsocket/islazy/fs"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"sort"
	"sync"
	"time"
)

var edgesParser = regexp.MustCompile(`^edges_(.+_[a-fA-F0-9:]{17})_(.+_.+)\.json$`)

type NodeCallback func(*Node)
type EdgeCallback func(*Node, *Edge, *Node)

type Graph struct {
	sync.Mutex

	path string
}

func NewGraph(path string) (*Graph, error) {
	g := &Graph{
		path: path,
	}
	return g, nil
}

func (g *Graph) EachNode(cb NodeCallback) error {
	g.Lock()
	defer g.Unlock()

	for _, nodeType := range NodeTypes {
		err := fs.Glob(g.path, fmt.Sprintf("%s_*.json", nodeType), func(fileName string) error {
			var node Node
			if raw, err := ioutil.ReadFile(fileName); err != nil {
				return fmt.Errorf("error while reading %s: %v", fileName, err)
			} else if err = json.Unmarshal(raw, &node); err != nil {
				return fmt.Errorf("error while decoding %s: %v", fileName, err)
			} else {
				cb(&node)
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

	return fs.Glob(g.path, "edges_*.json", func(fileName string) error {
		matches := edgesParser.FindAllStringSubmatch(path.Base(fileName), -1)
		if len(matches) > 0 && len(matches[0]) == 3 {
			var left, right Node
			leftFileName := path.Join(g.path, matches[0][1]+".json")
			rightFileName := path.Join(g.path, matches[0][2]+".json")

			if raw, err := ioutil.ReadFile(leftFileName); err != nil {
				return fmt.Errorf("error while reading %s: %v", leftFileName, err)
			} else if err = json.Unmarshal(raw, &left); err != nil {
				return fmt.Errorf("error while decoding %s: %v", leftFileName, err)
			} else if raw, err = ioutil.ReadFile(rightFileName); err != nil {
				return fmt.Errorf("error while reading %s: %v", rightFileName, err)
			} else if err = json.Unmarshal(raw, &right); err != nil {
				return fmt.Errorf("error while decoding %s: %v", rightFileName, err)
			}

			var edges []*Edge
			if raw, err := ioutil.ReadFile(fileName); err != nil {
				return fmt.Errorf("error while reading %s: %v", fileName, err)
			} else if err = json.Unmarshal(raw, &edges); err != nil {
				return fmt.Errorf("error while decoding %s: %v", fileName, err)
			}

			for _, edge := range edges {
				cb(&left, edge, &right)
			}
		} else {
			return fmt.Errorf("filename %s doesn't match edges parser", fileName)
		}
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
			left *Node
			edge *Edge
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

				// find all edges starting from this node
				edgesFilter := fmt.Sprintf("edges_%s_*.json", nodeID)
				err = fs.Glob(g.path, edgesFilter, func(edgeFileName string) error {
					right := new(Node)

					base := path.Base(edgeFileName)
					base = strings.ReplaceAll(base, "edges_", "")
					base = strings.ReplaceAll(base, nodeID + "_", "")

					// read right node
					rightFileName := path.Join(g.path, base)
					if raw, err := ioutil.ReadFile(rightFileName); err != nil {
						return fmt.Errorf("error while reading %s: %v", rightFileName, err)
					} else if err = json.Unmarshal(raw, right); err != nil {
						return fmt.Errorf("error while decoding %s: %v", rightFileName, err)
					}

					stack.Push(right)

					// read edges
					var edges []*Edge
					if raw, err := ioutil.ReadFile(edgeFileName); err != nil {
						return fmt.Errorf("error while reading %s: %v", edgeFileName, err)
					} else if err = json.Unmarshal(raw, &edges); err != nil {
						return fmt.Errorf("error while decoding %s: %v", edgeFileName, err)
					}

					for _, edge := range edges {
						allEdges = append(allEdges, edgeBucket {
							left: node,
							edge: edge,
							right: right,
						})
					}

					return nil
				})
				if err != nil {
					return err
				}
			}
		}

		for _, edge := range allEdges {
			onEdge(edge.left, edge.edge, edge.right)
		}
	}

	return nil
}

func (g *Graph) Dot(filter, layout, name string) (string, error) {
	data := fmt.Sprintf("digraph %s {\n", name)
	data += fmt.Sprintf("  layout=%s\n", layout)

	if err := g.Traverse(filter, func(node *Node) {
		data += fmt.Sprintf("  %s\n", node.Dot(filter == node.ID))
	}, func(left *Node, edge *Edge, right *Node) {
		data += fmt.Sprintf("  %s\n", edge.Dot(left, right))
	}); err != nil {
		return "", err
	}

	data += "\n"
	data += "  overlap=false\n"
	data += "}"

	return data, nil
}

func (g *Graph) FindNode(t NodeType, id string) (*Node, error) {
	g.Lock()
	defer g.Unlock()

	nodeFileName := path.Join(g.path, fmt.Sprintf("%s_%s.json", t, id))
	if fs.Exists(nodeFileName) {
		var node Node
		if raw, err := ioutil.ReadFile(nodeFileName); err != nil {
			return nil, fmt.Errorf("error while reading %s: %v", nodeFileName, err)
		} else if err = json.Unmarshal(raw, &node); err != nil {
			return nil, fmt.Errorf("error while decoding %s: %v", nodeFileName, err)
		}
		return &node, nil
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
				var node Node
				if raw, err := ioutil.ReadFile(nodeFileName); err != nil {
					return nil, fmt.Errorf("error while reading %s: %v", nodeFileName, err)
				} else if err = json.Unmarshal(raw, &node); err != nil {
					return nil, fmt.Errorf("error while decoding %s: %v", nodeFileName, err)
				} else {
					otherNodes = append(otherNodes, &node)
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
		CreatedAt:   time.Now(),
		Entity:      entity,
		Annotations: annotations,
	}

	nodeFileName := path.Join(g.path, fmt.Sprintf("%s.json", node.String()))
	if raw, err := json.Marshal(node); err != nil {
		return nil, fmt.Errorf("error creating data for %s: %v", nodeFileName, err)
	} else if err = ioutil.WriteFile(nodeFileName, raw, os.ModePerm); err != nil {
		return nil, fmt.Errorf("error creating %s: %v", nodeFileName, err)
	}

	session.I.Events.Add("graph.node.new", node)

	return node, nil
}

func (g *Graph) UpdateNode(node *Node) error {
	g.Lock()
	defer g.Unlock()

	node.UpdatedAt = time.Now()
	nodeFileName := path.Join(g.path, fmt.Sprintf("%s.json", node.String()))
	if raw, err := json.Marshal(node); err != nil {
		return fmt.Errorf("error creating new data for %s: %v", nodeFileName, err)
	} else if err = ioutil.WriteFile(nodeFileName, raw, os.ModePerm); err != nil {
		return fmt.Errorf("error updating %s: %v", nodeFileName, err)
	}

	return nil
}

func (g *Graph) findEdgesUnlocked(from, to *Node) (string, []*Edge, error) {
	edgesFileName := path.Join(g.path, fmt.Sprintf("edges_%s_%s.json", from.String(), to.String()))
	if fs.Exists(edgesFileName) {
		var edges []*Edge
		if raw, err := ioutil.ReadFile(edgesFileName); err != nil {
			return edgesFileName, nil, fmt.Errorf("error while reading %s: %v", edgesFileName, err)
		} else if err = json.Unmarshal(raw, &edges); err != nil {
			return edgesFileName, nil, fmt.Errorf("error while decoding %s: %v", edgesFileName, err)
		}

		// sort edges from oldest to newer
		sort.Slice(edges, func(i, j int) bool {
			return edges[i].CreatedAt.Before(edges[j].CreatedAt)
		})

		return edgesFileName, edges, nil
	}

	return edgesFileName, nil, nil
}

func (g *Graph) FindEdges(from, to *Node) ([]*Edge, error) {
	g.Lock()
	defer g.Unlock()

	_, edges, err := g.findEdgesUnlocked(from, to)
	return edges, err
}

func (g *Graph) FindLastEdgeOfType(from, to *Node, edgeType EdgeType) (*Edge, error) {
	g.Lock()
	defer g.Unlock()

	if _, edges, err := g.findEdgesUnlocked(from, to); err != nil {
		return nil, err
	} else {
		num := len(edges)
		for i := range edges {
			// loop backwards
			idx := num - 1 - i
			edge := edges[idx]
			if edge.Type == edgeType {
				return edge, nil
			}
		}
	}

	return nil, nil
}

func (g *Graph) FindLastRecentEdgeOfType(from, to *Node, edgeType EdgeType, staleTime time.Duration) (*Edge, error) {
	g.Lock()
	defer g.Unlock()

	if _, edges, err := g.findEdgesUnlocked(from, to); err != nil {
		return nil, err
	} else {
		num := len(edges)
		for i := range edges {
			// loop backwards
			idx := num - 1 - i
			edge := edges[idx]
			if edge.Type == edgeType {
				if time.Since(edge.CreatedAt) >= staleTime {
					return nil, nil
				}
				// edge is still fresh
				return edge, nil
			}
		}
	}

	return nil, nil
}

func (g *Graph) CreateEdge(from, to *Node, edgeType EdgeType) (*Edge, error) {
	g.Lock()
	defer g.Unlock()

	var edgesFileName string
	var edges []*Edge

	edge := &Edge{
		Type:      edgeType,
		CreatedAt: time.Now(),
		Position:  session.I.GPS,
	}

	if edgesFileName, edges, _ = g.findEdgesUnlocked(from, to); edges != nil {
		edges = append(edges, edge)
	} else {
		edges = []*Edge{edge}
	}

	if raw, err := json.Marshal(edges); err != nil {
		return nil, fmt.Errorf("error creating data for %s: %v", edgesFileName, err)
	} else if err = ioutil.WriteFile(edgesFileName, raw, os.ModePerm); err != nil {
		return nil, fmt.Errorf("error writing %s: %v", edgesFileName, err)
	}

	session.I.Events.Add("graph.edge.new", EdgeEvent{
		Left: from,
		Edge: edge,
		Right: to,
	})

	return edge, nil
}
