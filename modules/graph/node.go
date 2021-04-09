package graph

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
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

var nodeTypeDescs = map[NodeType]string{
	SSID:        "WiFI SSID probe",
	BLEServer:   "BLE Device",
	Station:     "WiFi Client",
	AccessPoint: "WiFi AP",
	Endpoint:    "IP Client",
	Gateway:     "IP Gateway",
}

var nodeDotStyles = map[NodeType]string{
	SSID:        "shape=circle style=filled color=lightgray fillcolor=lightgray fixedsize=true penwidth=0.5",
	BLEServer:   "shape=box style=filled color=dodgerblue3",
	Endpoint:    "shape=box style=filled color=azure2",
	Gateway:     "shape=diamond style=filled color=azure4",
	Station:     "shape=box style=filled color=gold",
	AccessPoint: "shape=diamond style=filled color=goldenrod3",
}

type Node struct {
	Type        NodeType    `json:"type"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	ID          string      `json:"id"`
	Annotations string      `json:"annotations"`
	Entity      interface{} `json:"entity"`
	Dummy       bool        `json:"-"`
}

func ReadNode(fileName string) (*Node, error) {
	var node Node
	if raw, err := ioutil.ReadFile(fileName); err != nil {
		return nil, fmt.Errorf("error while reading %s: %v", fileName, err)
	} else if err = json.Unmarshal(raw, &node); err != nil {
		return nil, fmt.Errorf("error while decoding %s: %v", fileName, err)
	}
	return &node, nil
}

func WriteNode(fileName string, node *Node, update bool) error {
	if update {
		node.UpdatedAt = time.Now()
	} else {
		node.CreatedAt = time.Now()
	}

	if raw, err := json.Marshal(node); err != nil {
		return fmt.Errorf("error creating data for %s: %v", fileName, err)
	} else if err = ioutil.WriteFile(fileName, raw, os.ModePerm); err != nil {
		return fmt.Errorf("error creating %s: %v", fileName, err)
	}
	return nil
}

func CreateNode(fileName string, node *Node) error {
	return WriteNode(fileName, node, false)
}

func UpdateNode(fileName string, node *Node) error {
	return WriteNode(fileName, node, true)
}

func (n Node) String() string {
	if n.Dummy == false {
		return fmt.Sprintf("%s_%s", n.Type, n.ID)
	}
	return string(n.Type)
}

func (n Node) Label() string {
	if n.Dummy {
		return n.Annotations
	}

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
		return fmt.Sprintf("%s\\n%s\\n(%s)",
			n.Entity.(map[string]interface{})["hostname"].(string),
			n.Entity.(map[string]interface{})["mac"].(string),
			n.Entity.(map[string]interface{})["vendor"].(string))
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
	return fmt.Sprintf("\"%s\" [%s, label=\"%s\"];",
		n.String(),
		style,
		strings.ReplaceAll(n.Label(), "\"", "\\\""))
}

func (n Node) ToMap() (map[string]interface{}, error) {
	var m map[string]interface{}

	if raw, err := json.Marshal(n); err != nil {
		return nil, err
	} else if err = json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}

	return m, nil
}
