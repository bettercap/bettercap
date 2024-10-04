package dns_proxy

import (
	"strings"

	"github.com/miekg/dns"
)

func shortenResourceRecords(records []string) []string {
	shorterRecords := []string{}
	for _, record := range records {
		shorterRecord := strings.ReplaceAll(record, "\t", " ")
		shorterRecords = append(shorterRecords, shorterRecord)
	}
	return shorterRecords
}

func (p *DNSProxy) logRequestAction(j *JSQuery) {
	p.Sess.Events.Add(p.Name+".spoofed-request", struct {
		Client    string
		Questions []string
	}{
		j.Client["IP"],
		shortenResourceRecords(j.Questions),
	})
}

func (p *DNSProxy) logResponseAction(j *JSQuery) {
	p.Sess.Events.Add(p.Name+".spoofed-response", struct {
		client      string
		Extras      []string
		Answers     []string
		Nameservers []string
		Questions   []string
	}{
		j.Client["IP"],
		shortenResourceRecords(j.Extras),
		shortenResourceRecords(j.Answers),
		shortenResourceRecords(j.Nameservers),
		shortenResourceRecords(j.Questions),
	})
}

func (p *DNSProxy) onRequestFilter(query *dns.Msg, clientIP string) (req, res *dns.Msg) {
	p.Debug("< %s %s", clientIP, query.Question)

	// do we have a proxy script?
	if p.Script == nil {
		return query, nil
	}

	// run the module OnRequest callback if defined
	jsreq, jsres := p.Script.OnRequest(query, clientIP)
	if jsreq != nil {
		// the request has been changed by the script
		p.logRequestAction(jsreq)
		return jsreq.ToQuery(), nil
	} else if jsres != nil {
		// a fake response has been returned by the script
		p.logResponseAction(jsres)
		return query, jsres.ToQuery()
	}

	return query, nil
}

func (p *DNSProxy) onResponseFilter(req, res *dns.Msg, clientIP string) *dns.Msg {
	// sometimes it happens ¯\_(ツ)_/¯
	if res == nil {
		return nil
	}

	p.Debug("> %s %s", clientIP, res.Answer)

	// do we have a proxy script?
	if p.Script != nil {
		_, jsres := p.Script.OnResponse(req, res, clientIP)
		if jsres != nil {
			// the response has been changed by the script
			p.logResponseAction(jsres)
			return jsres.ToQuery()
		}
	}

	return res
}
