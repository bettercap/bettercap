package graph

import (
	"github.com/bettercap/bettercap/v2/log"
)

type graphPackage struct{}

func (g graphPackage) IsConnected(nodeType, nodeID string) bool {
	if Loaded == nil {
		log.Error("graph.IsConnected: graph not loaded")
		return false
	}
	return Loaded.IsConnected(nodeType, nodeID)
}
