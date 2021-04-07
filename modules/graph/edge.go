package graph

import (
	"fmt"
	"github.com/bettercap/bettercap/session"
	"time"
)

type EdgeType string

const (
	Is         EdgeType = "is"
	ProbesFor  EdgeType = "probes_for"
	ConnectsTo EdgeType = "connects_to"
	Manages    EdgeType = "manages"
)

type EdgeEvent struct {
	Left *Node
	Edge *Edge
	Right *Node
}

type Edge struct {
	Type      EdgeType    `json:"type"`
	CreatedAt time.Time   `json:"created_at"`
	Position  session.GPS `json:"position"`
}

func (e Edge) Dot(left, right *Node) string {
	edgeLen := 2.0
	if e.Type == Is {
		edgeLen = 1.0
	}
	return fmt.Sprintf("\"%s\" -> \"%s\" [label=\"%s\", len=%.2f];",
		left.String(),
		right.String(),
		e.Type,
		edgeLen)
}
