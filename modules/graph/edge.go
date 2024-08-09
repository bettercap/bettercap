package graph

import (
	"fmt"
	"time"

	"github.com/bettercap/bettercap/v2/session"
)

type EdgeType string

const (
	Is         EdgeType = "is"
	ProbesFor  EdgeType = "probes_for"
	ProbedBy   EdgeType = "probed_by"
	ConnectsTo EdgeType = "connects_to"
	Manages    EdgeType = "manages"
)

type EdgeEvent struct {
	Left  *Node
	Edge  *Edge
	Right *Node
}

type Edge struct {
	Type      EdgeType     `json:"type"`
	CreatedAt time.Time    `json:"created_at"`
	Position  *session.GPS `json:"position,omitempty"`
}

func (e Edge) Dot(left, right *Node, width float64) string {
	edgeLen := 1.0
	if e.Type == Is {
		edgeLen = 0.3
	}
	return fmt.Sprintf("\"%s\" -> \"%s\" [label=\"%s\", len=%.2f, penwidth=%.2f];",
		left.String(),
		right.String(),
		e.Type,
		edgeLen,
		width)
}
