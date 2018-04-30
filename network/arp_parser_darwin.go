package network

import "regexp"

var ArpTableParser = regexp.MustCompile(`^[^\d\.]+([\d\.]+).+\s+([a-f0-9:]{11,17})\s+on\s+([^\s]+)\s+.+$`)
var ArpTableTokens = 4
var ArpTableTokenIndex = []int{1, 2, 3}
var ArpCmd = "arp"
var ArpCmdOpts = []string{"-a", "-n"}
