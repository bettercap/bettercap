package dns_proxy

import (
	"fmt"
	"net"

	"github.com/bettercap/bettercap/v2/log"
	"github.com/miekg/dns"
)

func NewJSEDNS0(e dns.EDNS0) (jsEDNS0 map[string]interface{}, err error) {
	option := e.Option()

	jsEDNS0 = map[string]interface{}{
		"Option": option,
	}

	var jsVal map[string]interface{}

	switch opt := e.(type) {
	case *dns.EDNS0_LLQ:
		jsVal = map[string]interface{}{
			"Code":      opt.Code,
			"Error":     opt.Error,
			"Id":        opt.Id,
			"LeaseLife": opt.LeaseLife,
			"Opcode":    opt.Opcode,
			"Version":   opt.Version,
		}
	case *dns.EDNS0_UL:
		jsVal = map[string]interface{}{
			"Code":     opt.Code,
			"Lease":    opt.Lease,
			"KeyLease": opt.KeyLease,
		}
	case *dns.EDNS0_NSID:
		jsVal = map[string]interface{}{
			"Code": opt.Code,
			"Nsid": opt.Nsid,
		}
	case *dns.EDNS0_ESU:
		jsVal = map[string]interface{}{
			"Code": opt.Code,
			"Uri":  opt.Uri,
		}
	case *dns.EDNS0_DAU:
		jsVal = map[string]interface{}{
			"AlgCode": opt.AlgCode,
			"Code":    opt.Code,
		}
	case *dns.EDNS0_DHU:
		jsVal = map[string]interface{}{
			"AlgCode": opt.AlgCode,
			"Code":    opt.Code,
		}
	case *dns.EDNS0_N3U:
		jsVal = map[string]interface{}{
			"AlgCode": opt.AlgCode,
			"Code":    opt.Code,
		}
	case *dns.EDNS0_SUBNET:
		jsVal = map[string]interface{}{
			"Address":       opt.Address.String(),
			"Code":          opt.Code,
			"Family":        opt.Family,
			"SourceNetmask": opt.SourceNetmask,
			"SourceScope":   opt.SourceScope,
		}
	case *dns.EDNS0_EXPIRE:
		jsVal = map[string]interface{}{
			"Code":   opt.Code,
			"Empty":  opt.Empty,
			"Expire": opt.Expire,
		}
	case *dns.EDNS0_COOKIE:
		jsVal = map[string]interface{}{
			"Code":   opt.Code,
			"Cookie": opt.Cookie,
		}
	case *dns.EDNS0_TCP_KEEPALIVE:
		jsVal = map[string]interface{}{
			"Code":    opt.Code,
			"Length":  opt.Length,
			"Timeout": opt.Timeout,
		}
	case *dns.EDNS0_PADDING:
		jsVal = map[string]interface{}{
			"Padding": string(opt.Padding),
		}
	case *dns.EDNS0_EDE:
		jsVal = map[string]interface{}{
			"ExtraText": opt.ExtraText,
			"InfoCode":  opt.InfoCode,
		}
	case *dns.EDNS0_LOCAL:
		jsVal = map[string]interface{}{
			"Code": opt.Code,
			"Data": string(opt.Data),
		}
	default:
		return nil, fmt.Errorf("unsupported EDNS0 option: %d", option)
	}

	jsEDNS0["Value"] = jsVal

	return jsEDNS0, nil
}

func ToEDNS0(jsEDNS0 map[string]interface{}) (e dns.EDNS0, err error) {
	option := jsPropToUint16(jsEDNS0, "Option")

	jsVal := jsPropToMap(jsEDNS0, "Value")

	switch option {
	case dns.EDNS0LLQ:
		e = &dns.EDNS0_LLQ{
			Code:      jsPropToUint16(jsVal, "Code"),
			Error:     jsPropToUint16(jsVal, "Error"),
			Id:        jsPropToUint64(jsVal, "Id"),
			LeaseLife: jsPropToUint32(jsVal, "LeaseLife"),
			Opcode:    jsPropToUint16(jsVal, "Opcode"),
			Version:   jsPropToUint16(jsVal, "Version"),
		}
	case dns.EDNS0UL:
		e = &dns.EDNS0_UL{
			Code:     jsPropToUint16(jsVal, "Code"),
			Lease:    jsPropToUint32(jsVal, "Lease"),
			KeyLease: jsPropToUint32(jsVal, "KeyLease"),
		}
	case dns.EDNS0NSID:
		e = &dns.EDNS0_NSID{
			Code: jsPropToUint16(jsVal, "Code"),
			Nsid: jsPropToString(jsVal, "Nsid"),
		}
	case dns.EDNS0ESU:
		e = &dns.EDNS0_ESU{
			Code: jsPropToUint16(jsVal, "Code"),
			Uri:  jsPropToString(jsVal, "Uri"),
		}
	case dns.EDNS0DAU:
		e = &dns.EDNS0_DAU{
			AlgCode: jsPropToUint8Array(jsVal, "AlgCode"),
			Code:    jsPropToUint16(jsVal, "Code"),
		}
	case dns.EDNS0DHU:
		e = &dns.EDNS0_DHU{
			AlgCode: jsPropToUint8Array(jsVal, "AlgCode"),
			Code:    jsPropToUint16(jsVal, "Code"),
		}
	case dns.EDNS0N3U:
		e = &dns.EDNS0_N3U{
			AlgCode: jsPropToUint8Array(jsVal, "AlgCode"),
			Code:    jsPropToUint16(jsVal, "Code"),
		}
	case dns.EDNS0SUBNET:
		e = &dns.EDNS0_SUBNET{
			Address:       net.ParseIP(jsPropToString(jsVal, "Address")),
			Code:          jsPropToUint16(jsVal, "Code"),
			Family:        jsPropToUint16(jsVal, "Family"),
			SourceNetmask: jsPropToUint8(jsVal, "SourceNetmask"),
			SourceScope:   jsPropToUint8(jsVal, "SourceScope"),
		}
	case dns.EDNS0EXPIRE:
		if empty, ok := jsVal["Empty"].(bool); !ok {
			log.Error("invalid or missing EDNS0_EXPIRE.Empty bool value, skipping field.")
			e = &dns.EDNS0_EXPIRE{
				Code:   jsPropToUint16(jsVal, "Code"),
				Expire: jsPropToUint32(jsVal, "Expire"),
			}
		} else {
			e = &dns.EDNS0_EXPIRE{
				Code:   jsPropToUint16(jsVal, "Code"),
				Expire: jsPropToUint32(jsVal, "Expire"),
				Empty:  empty,
			}
		}
	case dns.EDNS0COOKIE:
		e = &dns.EDNS0_COOKIE{
			Code:   jsPropToUint16(jsVal, "Code"),
			Cookie: jsPropToString(jsVal, "Cookie"),
		}
	case dns.EDNS0TCPKEEPALIVE:
		e = &dns.EDNS0_TCP_KEEPALIVE{
			Code:    jsPropToUint16(jsVal, "Code"),
			Length:  jsPropToUint16(jsVal, "Length"),
			Timeout: jsPropToUint16(jsVal, "Timeout"),
		}
	case dns.EDNS0PADDING:
		e = &dns.EDNS0_PADDING{
			Padding: []byte(jsPropToString(jsVal, "Padding")),
		}
	case dns.EDNS0EDE:
		e = &dns.EDNS0_EDE{
			ExtraText: jsPropToString(jsVal, "ExtraText"),
			InfoCode:  jsPropToUint16(jsVal, "InfoCode"),
		}
	case dns.EDNS0LOCALSTART, dns.EDNS0LOCALEND, 0x8000:
		// _DO = 0x8000
		e = &dns.EDNS0_LOCAL{
			Code: jsPropToUint16(jsVal, "Code"),
			Data: []byte(jsPropToString(jsVal, "Data")),
		}
	default:
		return nil, fmt.Errorf("unsupported EDNS0 option: %d", option)
	}

	return e, nil
}
