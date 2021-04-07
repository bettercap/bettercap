package graph

import (
	"fmt"
	"time"
	"unicode"
)

type NodeType string

const (
	SSID        NodeType = "ssid"
	BLEServer   NodeType = "ble_server"
	Station     NodeType = "station"
	AccessPoint NodeType = "access_point"
	Endpoint    NodeType = "endpoint"
	Gateway     NodeType = "gateway"
)

var NodeTypes = []NodeType{
	SSID,
	Station,
	AccessPoint,
	Endpoint,
	Gateway,
	BLEServer,
}

var nodeDotStyles = map[NodeType]string{
	SSID:        "shape=diamond",
	BLEServer:   "shape=box, style=filled, color=dodgerblue3",
	Endpoint:    "shape=box, style=filled, color=azure2",
	Gateway:     "shape=diamond, style=filled, color=azure4",
	Station:     "shape=box, style=filled, color=gold",
	AccessPoint: "shape=diamond, style=filled, color=goldenrod3",
}

type Node struct {
	Type        NodeType    `json:"type"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	ID          string      `json:"id"`
	Annotations string      `json:"annotations"`
	Entity      interface{} `json:"entity"`
}

func (n Node) String() string {
	return fmt.Sprintf("%s_%s", n.Type, n.ID)
}

func (n Node) Label() string {
	switch n.Type {
	case SSID:
		s := n.Entity.(string)
		allPrint := true

		for _, rn := range s {
			if !unicode.IsPrint(rune(rn)) {
				allPrint = false
				break
			}
		}

		if !allPrint {
			s = fmt.Sprintf("0x%x", s)
		}
		return s
	case BLEServer:
		return fmt.Sprintf("%s\\n(%s)",
			n.Entity.(map[string]interface{})["mac"].(string),
			n.Entity.(map[string]interface{})["vendor"].(string))
	case Station:
		return fmt.Sprintf("%s\\n(%s)",
			n.Entity.(map[string]interface{})["mac"].(string),
			n.Entity.(map[string]interface{})["vendor"].(string))
	case AccessPoint:
		return fmt.Sprintf("%s\\n(%s)",
			n.Entity.(map[string]interface{})["hostname"].(string),
			n.Entity.(map[string]interface{})["mac"].(string))
	case Endpoint:
		return fmt.Sprintf("%s\\n(%s %s)",
			n.Entity.(map[string]interface{})["ipv4"].(string),
			n.Entity.(map[string]interface{})["mac"].(string),
			n.Entity.(map[string]interface{})["vendor"].(string))
	case Gateway:
		return fmt.Sprintf("%s\\n(%s %s)",
			n.Entity.(map[string]interface{})["ipv4"].(string),
			n.Entity.(map[string]interface{})["mac"].(string),
			n.Entity.(map[string]interface{})["vendor"].(string))
	}
	return "?"
}

func (n Node) Dot(isTarget bool) string {
	style := nodeDotStyles[n.Type]
	if isTarget {
		style += ", color=red"
	}
	return fmt.Sprintf("node [%s]; {node [label=\"%s\"] \"%s\";};",
		style,
		n.Label(),
		n.String())
}
