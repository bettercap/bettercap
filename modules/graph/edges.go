package graph

import (
	"encoding/json"
	"github.com/evilsocket/islazy/fs"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"sync"
	"time"
)

const edgesIndexName = "edges.json"

type EdgesTo map[string][]Edge

type EdgesCallback func(string, []Edge, string) error

type Edges struct {
	sync.RWMutex
	timestamp time.Time
	fileName  string
	size      int
	from      map[string]EdgesTo
}

type edgesJSON struct {
	Timestamp time.Time          `json:"timestamp"`
	Size      int                `json:"size"`
	Edges     map[string]EdgesTo `json:"edges"`
}

func LoadEdges(basePath string) (*Edges, error) {
	edges := Edges{
		fileName: path.Join(basePath, edgesIndexName),
		from:     make(map[string]EdgesTo),
	}

	if fs.Exists(edges.fileName) {
		var js edgesJSON

		if raw, err := ioutil.ReadFile(edges.fileName); err != nil {
			return nil, err
		} else if err = json.Unmarshal(raw, &js); err != nil {
			return nil, err
		}

		edges.timestamp = js.Timestamp
		edges.from = js.Edges
		edges.size = js.Size
	}

	return &edges, nil
}

func (e *Edges) flush() error {
	e.timestamp = time.Now()
	js := edgesJSON{
		Timestamp: e.timestamp,
		Size:      e.size,
		Edges:     e.from,
	}

	if raw, err := json.Marshal(js); err != nil {
		return err
	} else if err = ioutil.WriteFile(e.fileName, raw, os.ModePerm); err != nil {
		return err
	}

	return nil
}

func (e *Edges) Flush() error {
	e.RLock()
	defer e.RUnlock()
	return e.flush()
}

func (e *Edges) ForEachEdge(cb EdgesCallback) error {
	e.RLock()
	defer e.RUnlock()

	for from, edgesTo := range e.from {
		for to, edges := range edgesTo {
			if err := cb(from, edges, to); err != nil {
				return err
			}
		}
	}

	return nil
}

func (e *Edges) ForEachEdgeFrom(nodeID string, cb EdgesCallback) error {
	e.RLock()
	defer e.RUnlock()

	if edgesTo, found := e.from[nodeID]; found {
		for to, edges := range edgesTo {
			if err := cb(nodeID, edges, to); err != nil {
				return err
			}
		}
	}

	return nil
}

func (e *Edges) IsConnected(nodeID string) bool {
	e.RLock()
	defer e.RUnlock()

	if edgesTo, found := e.from[nodeID]; found {
		return len(edgesTo) > 0
	}

	return false
}

func (e *Edges) FindEdges(fromID, toID string, doSort bool) []Edge {
	e.RLock()
	defer e.RUnlock()

	if edgesTo, foundFrom := e.from[fromID]; foundFrom {
		if edges, foundTo := edgesTo[toID]; foundTo {
			if doSort {
				// sort edges from oldest to newer
				sort.Slice(edges, func(i, j int) bool {
					return edges[i].CreatedAt.Before(edges[j].CreatedAt)
				})
			}
			return edges
		}
	}

	return nil
}

func (e *Edges) Connect(fromID, toID string, edge Edge) error {
	e.Lock()
	defer e.Unlock()

	if edgesTo, foundFrom := e.from[fromID]; foundFrom {
		edges := edgesTo[toID]
		edges = append(edges, edge)
		e.from[fromID][toID] = edges
	} else {
		// create the entire path
		e.from[fromID] = EdgesTo{
			toID: {edge},
		}
	}

	e.size++

	return e.flush()
}
