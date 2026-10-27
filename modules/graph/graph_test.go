package graph

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestSafeFileName(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		windows bool
		want    string
	}{
		{"mac on windows", "endpoint_8c:1d:89:fe:12:eb", true, "endpoint_8c_1d_89_fe_12_eb"},
		{"mac elsewhere keeps existing name", "endpoint_8c:1d:89:fe:12:eb", false, "endpoint_8c:1d:89:fe:12:eb"},
		{"other reserved characters", `a<b>c"d|e?f*g\h`, true, "a_b_c_d_e_f_g_h"},
		{"already safe", "ssid_4142", true, "ssid_4142"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := safeFileName(tc.in, tc.windows); got != tc.want {
				t.Errorf("safeFileName(%q, %v) = %q, want %q", tc.in, tc.windows, got, tc.want)
			}
		})
	}
}

// Nodes are keyed by "<type>_<mac>"; the file for that key must be writable
// and readable back on the platform the test runs on.
func TestNodeFileRoundTrip(t *testing.T) {
	g := &Graph{path: t.TempDir()}
	node := &Node{Type: Endpoint, ID: "8c:1d:89:fe:12:eb"}

	file := g.nodeFile(node.String())
	if filepath.Dir(file) != filepath.Clean(g.path) {
		t.Fatalf("node file %q is not directly inside %q", file, g.path)
	}
	if err := CreateNode(file, node); err != nil {
		t.Fatalf("CreateNode(%q): %v", file, err)
	}

	found, err := g.FindNode(Endpoint, node.ID)
	if err != nil {
		t.Fatalf("FindNode: %v", err)
	}
	if found == nil || found.ID != node.ID {
		t.Fatalf("FindNode = %+v, want node with ID %q (GOOS=%s)", found, node.ID, runtime.GOOS)
	}
}

// graph on walks every node file, which is where the Windows failure surfaced.
func TestEachNodeReadsBackNodes(t *testing.T) {
	g := &Graph{path: t.TempDir()}
	ids := map[string]bool{"8c:1d:89:fe:12:eb": false, "00:11:22:33:44:55": false}

	for id := range ids {
		node := &Node{Type: Endpoint, ID: id}
		if err := CreateNode(g.nodeFile(node.String()), node); err != nil {
			t.Fatalf("CreateNode(%s): %v", id, err)
		}
	}

	if err := g.EachNode(func(n *Node) { ids[n.ID] = true }); err != nil {
		t.Fatalf("EachNode: %v", err)
	}
	for id, seen := range ids {
		if !seen {
			t.Errorf("EachNode did not return node %s", id)
		}
	}
}
