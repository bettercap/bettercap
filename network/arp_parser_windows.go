package network

import "regexp"

var ArpTableParser = regexp.MustCompile(`^[^\d\.]+([\d\.]+).+\s+([a-f0-9\-]{11,17})\s+.+$`)
var ArpTableTokens = 3
var ArpTableTokenIndex = []int{1, 2, -1}
var ArpCmd = "arp"
var ArpCmdOpts = []string{"-a"}
