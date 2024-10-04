package dns_proxy

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bettercap/bettercap/v2/log"
	"github.com/bettercap/bettercap/v2/session"

	"github.com/miekg/dns"
)

var whiteSpaceRegexp = regexp.MustCompile(`\s+`)
var stripWhiteSpaceRegexp = regexp.MustCompile(`^\s*(.*?)\s*$`)

type JSQuery struct {
	Answers     []string
	Client      map[string]string
	Compress    bool `json:"-"`
	Extras      []string
	Header      *JSQueryHeader
	Nameservers []string
	Questions   []string

	refHash string
}

type JSQueryHeader struct {
	AuthenticatedData  bool
	Authoritative      bool
	CheckingDisabled   bool
	Id                 uint16
	Opcode             int
	Rcode              int
	RecursionAvailable bool
	RecursionDesired   bool
	Response           bool
	Truncated          bool
	Zero               bool
}

func (j *JSQuery) NewHash() string {
	headerHash := fmt.Sprintf("%t.%t.%t.%d.%d.%d.%t.%t.%t.%t.%t",
		j.Header.AuthenticatedData,
		j.Header.Authoritative,
		j.Header.CheckingDisabled,
		j.Header.Id,
		j.Header.Opcode,
		j.Header.Rcode,
		j.Header.RecursionAvailable,
		j.Header.RecursionDesired,
		j.Header.Response,
		j.Header.Truncated,
		j.Header.Zero)
	hash := fmt.Sprintf("%s.%s.%t.%s.%s.%s.%s",
		strings.Join(j.Answers, ""),
		j.Client["IP"],
		j.Compress,
		strings.Join(j.Extras, ""),
		headerHash,
		strings.Join(j.Nameservers, ""),
		strings.Join(j.Questions, ""))
	return hash
}

func NewJSQuery(query *dns.Msg, clientIP string) *JSQuery {
	answers := []string{}
	extras := []string{}
	nameservers := []string{}
	questions := []string{}

	header := &JSQueryHeader{
		AuthenticatedData:  query.MsgHdr.AuthenticatedData,
		Authoritative:      query.MsgHdr.Authoritative,
		CheckingDisabled:   query.MsgHdr.CheckingDisabled,
		Id:                 query.MsgHdr.Id,
		Opcode:             query.MsgHdr.Opcode,
		Rcode:              query.MsgHdr.Rcode,
		RecursionAvailable: query.MsgHdr.RecursionAvailable,
		RecursionDesired:   query.MsgHdr.RecursionDesired,
		Response:           query.MsgHdr.Response,
		Truncated:          query.MsgHdr.Truncated,
		Zero:               query.MsgHdr.Zero,
	}

	for _, rr := range query.Answer {
		answers = append(answers, rr.String())
	}
	for _, rr := range query.Extra {
		extras = append(extras, rr.String())
	}
	for _, rr := range query.Ns {
		nameservers = append(nameservers, rr.String())
	}
	for _, q := range query.Question {
		qType := dns.Type(q.Qtype).String()
		qClass := dns.Class(q.Qclass).String()
		questions = append(questions, fmt.Sprintf("%s\t%s\t%s", q.Name, qClass, qType))
	}

	clientMAC := ""
	clientAlias := ""
	if endpoint := session.I.Lan.GetByIp(clientIP); endpoint != nil {
		clientMAC = endpoint.HwAddress
		clientAlias = endpoint.Alias
	}
	client := map[string]string{"IP": clientIP, "MAC": clientMAC, "Alias": clientAlias}

	jsquery := &JSQuery{
		Answers:     answers,
		Client:      client,
		Compress:    query.Compress,
		Extras:      extras,
		Header:      header,
		Nameservers: nameservers,
		Questions:   questions,
	}
	jsquery.UpdateHash()

	return jsquery
}

func stringToClass(s string) (uint16, error) {
	for i, dnsClass := range dns.ClassToString {
		if s == dnsClass {
			return i, nil
		}
	}
	return 0, fmt.Errorf("unkown DNS class (got %s)", s)
}

func stringToType(s string) (uint16, error) {
	for i, dnsType := range dns.TypeToString {
		if s == dnsType {
			return i, nil
		}
	}
	return 0, fmt.Errorf("unkown DNS type (got %s)", s)
}

func (j *JSQuery) ToQuery() *dns.Msg {
	var answers []dns.RR
	var extras []dns.RR
	var nameservers []dns.RR
	var questions []dns.Question

	for _, s := range j.Answers {
		rr, err := dns.NewRR(s)
		if err != nil {
			log.Error("error parsing DNS answer resource record: %s", err.Error())
			return nil
		} else {
			answers = append(answers, rr)
		}
	}
	for _, s := range j.Extras {
		rr, err := dns.NewRR(s)
		if err != nil {
			log.Error("error parsing DNS extra resource record: %s", err.Error())
			return nil
		} else {
			extras = append(extras, rr)
		}
	}
	for _, s := range j.Nameservers {
		rr, err := dns.NewRR(s)
		if err != nil {
			log.Error("error parsing DNS nameserver resource record: %s", err.Error())
			return nil
		} else {
			nameservers = append(nameservers, rr)
		}
	}

	for _, s := range j.Questions {
		qStripped := stripWhiteSpaceRegexp.FindStringSubmatch(s)
		qParts := whiteSpaceRegexp.Split(qStripped[1], -1)

		if len(qParts) != 3 {
			log.Error("invalid DNS question format: (got %s)", s)
			return nil
		}

		qName := dns.Fqdn(qParts[0])
		qClass, err := stringToClass(qParts[1])
		if err != nil {
			log.Error("error parsing DNS question class: %s", err.Error())
			return nil
		}
		qType, err := stringToType(qParts[2])
		if err != nil {
			log.Error("error parsing DNS question type: %s", err.Error())
			return nil
		}

		questions = append(questions, dns.Question{qName, qType, qClass})
	}

	query := &dns.Msg{
		MsgHdr: dns.MsgHdr{
			Id:                 j.Header.Id,
			Response:           j.Header.Response,
			Opcode:             j.Header.Opcode,
			Authoritative:      j.Header.Authoritative,
			Truncated:          j.Header.Truncated,
			RecursionDesired:   j.Header.RecursionDesired,
			RecursionAvailable: j.Header.RecursionAvailable,
			Zero:               j.Header.Zero,
			AuthenticatedData:  j.Header.AuthenticatedData,
			CheckingDisabled:   j.Header.CheckingDisabled,
			Rcode:              j.Header.Rcode,
		},
		Compress: j.Compress,
		Question: questions,
		Answer:   answers,
		Ns:       nameservers,
		Extra:    extras,
	}

	return query
}

func (j *JSQuery) UpdateHash() {
	j.refHash = j.NewHash()
}

func (j *JSQuery) WasModified() bool {
	// check if any of the fields has been changed
	return j.NewHash() != j.refHash
}
