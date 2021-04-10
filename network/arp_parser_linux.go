package network

import "regexp"

var ArpTableParser = regexp.MustCompile(`^([a-f\d\.:]+)\s+dev\s+(\w+)\s+\w+\s+([a-f0-9:]{17})\s+.+$`)
var ArpTableTokens = 4
var ArpTableTokenIndex = []int{1, 3, 2}
var ArpCmd = "ip"
var ArpCmdOpts = []string{"neigh"}
