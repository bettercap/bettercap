package dns_proxy

import (
	"strings"

	"github.com/miekg/dns"
)

func questionsToStrings(qs []dns.Question) []string {
	questions := []string{}
	for _, q := range qs {
		questions = append(questions, tabsToSpaces(q.String()))
	}
	return questions
}

func recordsToStrings(rrs []dns.RR) []string {
	records := []string{}
	for _, rr := range rrs {
		if rr != nil {
			records = append(records, tabsToSpaces(rr.String()))
		}
	}
	return records
}

func tabsToSpaces(s string) string {
	return strings.ReplaceAll(s, "\t", " ")
}

func (p *DNSProxy) logRequestAction(m *dns.Msg, clientIP string) {
	p.Sess.Events.Add(p.Name+".spoofed-request", struct {
		Client    string
		Questions []string
	}{
		clientIP,
		questionsToStrings(m.Question),
	})
}

func (p *DNSProxy) logResponseAction(m *dns.Msg, clientIP string) {
	p.Sess.Events.Add(p.Name+".spoofed-response", struct {
		client      string
		Answers     []string
		Extras      []string
		Nameservers []string
		Questions   []string
	}{
		clientIP,
		recordsToStrings(m.Answer),
		recordsToStrings(m.Extra),
		recordsToStrings(m.Ns),
		questionsToStrings(m.Question),
	})
}

func (p *DNSProxy) onRequestFilter(query *dns.Msg, clientIP string) (req, res *dns.Msg) {
	if p.shouldProxy(clientIP) {
		p.Debug("< %s q[%s]",
			clientIP,
			strings.Join(questionsToStrings(query.Question), ","))

		// do we have a proxy script?
		if p.Script == nil {
			return query, nil
		}

		// run the module OnRequest callback if defined
		jsreq, jsres := p.Script.OnRequest(query, clientIP)
		if jsreq != nil {
			// the request has been changed by the script
			req := jsreq.ToQuery()
			p.logRequestAction(req, clientIP)
			return req, nil
		} else if jsres != nil {
			// a fake response has been returned by the script
			res := jsres.ToQuery()
			p.logResponseAction(res, clientIP)
			return query, res
		}
	}

	return query, nil
}

func (p *DNSProxy) onResponseFilter(req, res *dns.Msg, clientIP string) *dns.Msg {
	if p.shouldProxy(clientIP) {
		// sometimes it happens ¯\_(ツ)_/¯
		if res == nil {
			return nil
		}

		p.Debug("> %s q[%s] a[%s] e[%s] n[%s]",
			clientIP,
			strings.Join(questionsToStrings(res.Question), ","),
			strings.Join(recordsToStrings(res.Answer), ","),
			strings.Join(recordsToStrings(res.Extra), ","),
			strings.Join(recordsToStrings(res.Ns), ","))

		// do we have a proxy script?
		if p.Script != nil {
			_, jsres := p.Script.OnResponse(req, res, clientIP)
			if jsres != nil {
				// the response has been changed by the script
				res := jsres.ToQuery()
				p.logResponseAction(res, clientIP)
				return res
			}
		}
	}

	return res
}
