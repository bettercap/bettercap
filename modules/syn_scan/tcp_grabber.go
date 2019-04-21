package syn_scan

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

func tcpGrabber(mod *SynScanner, ip string, port int) string {
	if conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port)); err == nil {
		defer conn.Close()
		msg, _ := bufio.NewReader(conn).ReadString('\n')
		return strings.Trim(msg, "\r\n\t ")
	}

	return ""
}
