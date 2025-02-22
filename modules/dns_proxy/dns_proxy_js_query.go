package dns_proxy

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"reflect"

	"github.com/bettercap/bettercap/v2/log"
	"github.com/bettercap/bettercap/v2/session"

	"github.com/miekg/dns"
)

type JSQuery struct {
	Answers     []map[string]interface{}
	Client      map[string]string
	Compress    bool
	Extras      []map[string]interface{}
	Header      JSQueryHeader
	Nameservers []map[string]interface{}
	Questions   []map[string]interface{}

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

func jsPropToMap(obj map[string]interface{}, key string) map[string]interface{} {
	if v, ok := obj[key].(map[string]interface{}); ok {
		return v
	}
	log.Error("error converting JS property to map[string]interface{} where key is: %s", key)
	return map[string]interface{}{}
}

func jsPropToMapArray(obj map[string]interface{}, key string) []map[string]interface{} {
	if v, ok := obj[key].([]map[string]interface{}); ok {
		return v
	}
	log.Error("error converting JS property to []map[string]interface{} where key is: %s", key)
	return []map[string]interface{}{}
}

func jsPropToString(obj map[string]interface{}, key string) string {
	if v, ok := obj[key].(string); ok {
		return v
	}
	log.Error("error converting JS property to string where key is: %s", key)
	return ""
}

func jsPropToStringArray(obj map[string]interface{}, key string) []string {
	if v, ok := obj[key].([]string); ok {
		return v
	}
	log.Error("error converting JS property to []string where key is: %s", key)
	return []string{}
}

func jsPropToUint8(obj map[string]interface{}, key string) uint8 {
	if v, ok := obj[key].(int64); ok {
		if v >= 0 && v <= math.MaxUint8 {
			return uint8(v)
		}
	}
	log.Error("error converting JS property to uint8 where key is: %s", key)
	return uint8(0)
}

func jsPropToUint8Array(obj map[string]interface{}, key string) []uint8 {
	if arr, ok := obj[key].([]interface{}); ok {
		vArr := make([]uint8, 0, len(arr))
		for _, item := range arr {
			if v, ok := item.(int64); ok {
				if v >= 0 && v <= math.MaxUint8 {
					vArr = append(vArr, uint8(v))
				} else {
					log.Error("error converting JS property to []uint8 where key is: %s", key)
					return []uint8{}
				}
			}
		}
		return vArr
	}
	log.Error("error converting JS property to []uint8 where key is: %s", key)
	return []uint8{}
}

func jsPropToUint16(obj map[string]interface{}, key string) uint16 {
	if v, ok := obj[key].(int64); ok {
		if v >= 0 && v <= math.MaxUint16 {
			return uint16(v)
		}
	}
	log.Error("error converting JS property to uint16 where key is: %s", key)
	return uint16(0)
}

func jsPropToUint16Array(obj map[string]interface{}, key string) []uint16 {
	if arr, ok := obj[key].([]interface{}); ok {
		vArr := make([]uint16, 0, len(arr))
		for _, item := range arr {
			if v, ok := item.(int64); ok {
				if v >= 0 && v <= math.MaxUint16 {
					vArr = append(vArr, uint16(v))
				} else {
					log.Error("error converting JS property to []uint16 where key is: %s", key)
					return []uint16{}
				}
			}
		}
		return vArr
	}
	log.Error("error converting JS property to []uint16 where key is: %s", key)
	return []uint16{}
}

func jsPropToUint32(obj map[string]interface{}, key string) uint32 {
	if v, ok := obj[key].(int64); ok {
		if v >= 0 && v <= math.MaxUint32 {
			return uint32(v)
		}
	}
	log.Error("error converting JS property to uint32 where key is: %s", key)
	return uint32(0)
}

func jsPropToUint64(obj map[string]interface{}, key string) uint64 {
	prop, found := obj[key]
	if found {
		switch reflect.TypeOf(prop).String() {
		case "float64":
			if f, ok := prop.(float64); ok {
				bigInt := new(big.Float).SetFloat64(f)
				v, _ := bigInt.Uint64()
				if v >= 0 {
					return v
				}
			}
			break
		case "int64":
			if v, ok := prop.(int64); ok {
				if v >= 0 {
					return uint64(v)
				}
			}
			break
		case "uint64":
			if v, ok := prop.(uint64); ok {
				return v
			}
			break
		}
	}
	log.Error("error converting JS property to uint64 where key is: %s", key)
	return uint64(0)
}

func uint16ArrayToInt64Array(arr []uint16) []int64 {
	vArr := make([]int64, 0, len(arr))
	for _, item := range arr {
		vArr = append(vArr, int64(item))
	}
	return vArr
}

func (j *JSQuery) NewHash() string {
	answers, _ := json.Marshal(j.Answers)
	extras, _ := json.Marshal(j.Extras)
	nameservers, _ := json.Marshal(j.Nameservers)
	questions, _ := json.Marshal(j.Questions)

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
		answers,
		j.Client["IP"],
		j.Compress,
		extras,
		headerHash,
		nameservers,
		questions)

	return hash
}

func NewJSQuery(query *dns.Msg, clientIP string) (jsQuery *JSQuery) {
	answers := make([]map[string]interface{}, len(query.Answer))
	extras := make([]map[string]interface{}, len(query.Extra))
	nameservers := make([]map[string]interface{}, len(query.Ns))
	questions := make([]map[string]interface{}, len(query.Question))

	for i, rr := range query.Answer {
		jsRecord, err := NewJSResourceRecord(rr)
		if err != nil {
			log.Error(err.Error())
			continue
		}
		answers[i] = jsRecord
	}

	for i, rr := range query.Extra {
		jsRecord, err := NewJSResourceRecord(rr)
		if err != nil {
			log.Error(err.Error())
			continue
		}
		extras[i] = jsRecord
	}

	for i, rr := range query.Ns {
		jsRecord, err := NewJSResourceRecord(rr)
		if err != nil {
			log.Error(err.Error())
			continue
		}
		nameservers[i] = jsRecord
	}

	for i, question := range query.Question {
		questions[i] = map[string]interface{}{
			"Name":   question.Name,
			"Qtype":  int64(question.Qtype),
			"Qclass": int64(question.Qclass),
		}
	}

	clientMAC := ""
	clientAlias := ""
	if endpoint := session.I.Lan.GetByIp(clientIP); endpoint != nil {
		clientMAC = endpoint.HwAddress
		clientAlias = endpoint.Alias
	}
	client := map[string]string{"IP": clientIP, "MAC": clientMAC, "Alias": clientAlias}

	jsquery := &JSQuery{
		Answers:  answers,
		Client:   client,
		Compress: query.Compress,
		Extras:   extras,
		Header: JSQueryHeader{
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
		},
		Nameservers: nameservers,
		Questions:   questions,
	}
	jsquery.UpdateHash()

	return jsquery
}

func (j *JSQuery) ToQuery() *dns.Msg {
	var answers []dns.RR
	var extras []dns.RR
	var nameservers []dns.RR
	var questions []dns.Question

	for _, jsRR := range j.Answers {
		rr, err := ToRR(jsRR)
		if err != nil {
			log.Error(err.Error())
			continue
		}
		answers = append(answers, rr)
	}
	for _, jsRR := range j.Extras {
		rr, err := ToRR(jsRR)
		if err != nil {
			log.Error(err.Error())
			continue
		}
		extras = append(extras, rr)
	}
	for _, jsRR := range j.Nameservers {
		rr, err := ToRR(jsRR)
		if err != nil {
			log.Error(err.Error())
			continue
		}
		nameservers = append(nameservers, rr)
	}

	for _, jsQ := range j.Questions {
		questions = append(questions, dns.Question{
			Name:   jsPropToString(jsQ, "Name"),
			Qtype:  jsPropToUint16(jsQ, "Qtype"),
			Qclass: jsPropToUint16(jsQ, "Qclass"),
		})
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

func (j *JSQuery) CheckIfModifiedAndUpdateHash() bool {
	// check if query was changed and update its hash
	newHash := j.NewHash()
	wasModified := j.refHash != newHash
	j.refHash = newHash
	return wasModified
}
