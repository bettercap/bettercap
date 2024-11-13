package dns_proxy

import (
	"fmt"
	"net"

	"github.com/bettercap/bettercap/v2/log"
	"github.com/miekg/dns"
)

func NewJSSVCBKeyValue(kv dns.SVCBKeyValue) (map[string]interface{}, error) {
	key := kv.Key()

	jsKv := map[string]interface{}{
		"Key": uint16(key),
	}

	switch v := kv.(type) {
	case *dns.SVCBAlpn:
		jsKv["Alpn"] = v.Alpn
	case *dns.SVCBNoDefaultAlpn:
		break
	case *dns.SVCBECHConfig:
		jsKv["ECH"] = string(v.ECH)
	case *dns.SVCBPort:
		jsKv["Port"] = v.Port
	case *dns.SVCBIPv4Hint:
		ips := v.Hint
		jsIps := make([]string, len(ips))
		for i, ip := range ips {
			jsIps[i] = ip.String()
		}
		jsKv["Hint"] = jsIps
	case *dns.SVCBIPv6Hint:
		ips := v.Hint
		jsIps := make([]string, len(ips))
		for i, ip := range ips {
			jsIps[i] = ip.String()
		}
		jsKv["Hint"] = jsIps
	case *dns.SVCBDoHPath:
		jsKv["Template"] = v.Template
	case *dns.SVCBOhttp:
		break
	case *dns.SVCBMandatory:
		keys := v.Code
		jsKeys := make([]uint16, len(keys))
		for i, _key := range keys {
			jsKeys[i] = uint16(_key)
		}
		jsKv["Code"] = jsKeys
	default:
		return nil, fmt.Errorf("error creating JSSVCBKeyValue: unknown key: %d", key)
	}

	return jsKv, nil
}

func ToSVCBKeyValue(jsKv map[string]interface{}) (dns.SVCBKeyValue, error) {
	var kv dns.SVCBKeyValue

	key := dns.SVCBKey(jsPropToUint16(jsKv, "Key"))

	switch key {
	case dns.SVCB_ALPN:
		kv = &dns.SVCBAlpn{
			Alpn: jsPropToStringArray(jsKv, "Value"),
		}
	case dns.SVCB_NO_DEFAULT_ALPN:
		kv = &dns.SVCBNoDefaultAlpn{}
	case dns.SVCB_ECHCONFIG:
		kv = &dns.SVCBECHConfig{
			ECH: []byte(jsPropToString(jsKv, "Value")),
		}
	case dns.SVCB_PORT:
		kv = &dns.SVCBPort{
			Port: jsPropToUint16(jsKv, "Value"),
		}
	case dns.SVCB_IPV4HINT:
		jsIps := jsPropToStringArray(jsKv, "Value")
		var ips []net.IP
		for _, jsIp := range jsIps {
			ip := net.ParseIP(jsIp)
			if ip == nil {
				log.Error("error converting to SVCBKeyValue: invalid IPv4Hint IP: %s", jsIp)
				continue
			}
			ips = append(ips, ip)
		}
		kv = &dns.SVCBIPv4Hint{
			Hint: ips,
		}
	case dns.SVCB_IPV6HINT:
		jsIps := jsPropToStringArray(jsKv, "Value")
		var ips []net.IP
		for _, jsIp := range jsIps {
			ip := net.ParseIP(jsIp)
			if ip == nil {
				log.Error("error converting to SVCBKeyValue: invalid IPv6Hint IP: %s", jsIp)
				continue
			}
			ips = append(ips, ip)
		}
		kv = &dns.SVCBIPv6Hint{
			Hint: ips,
		}
	case dns.SVCB_DOHPATH:
		kv = &dns.SVCBDoHPath{
			Template: jsPropToString(jsKv, "Value"),
		}
	case dns.SVCB_OHTTP:
		kv = &dns.SVCBOhttp{}
	case dns.SVCB_MANDATORY:
		v := jsPropToUint16Array(jsKv, "Value")
		keys := make([]dns.SVCBKey, len(v))
		for i, jsKey := range v {
			keys[i] = dns.SVCBKey(jsKey)
		}
		kv = &dns.SVCBMandatory{
			Code: keys,
		}
	default:
		return nil, fmt.Errorf("error converting to dns.SVCBKeyValue: unknown key: %d", key)
	}

	return kv, nil
}
