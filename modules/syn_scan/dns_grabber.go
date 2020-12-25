package syn_scan

import (
	"fmt"
	"regexp"

	"github.com/miekg/dns"
)

var chaosParser = regexp.MustCompile(`.*"([^"]+)".*`)

func grabChaos(addr string, q string) string {
	c := new(dns.Client)
	m := new(dns.Msg)
	m.Question = make([]dns.Question, 1)
	m.Question[0] = dns.Question{Name: q, Qtype: dns.TypeTXT, Qclass: dns.ClassCHAOS}
	if in, _, _ := c.Exchange(m, addr); in != nil && len(in.Answer) > 0 {
		s := in.Answer[0].String()
		if match := chaosParser.FindStringSubmatch(s); len(match) > 0 {
			return match[1]
		}
	}
	return ""
}

func dnsGrabber(mod *SynScanner, ip string, port int) string {
	addr := fmt.Sprintf("%s:%d", ip, port)
	if v := grabChaos(addr, "version.bind."); v != "" {
		return v
	} else if v := grabChaos(addr, "hostname.bind."); v != "" {
		return v
	}
	return ""
}
