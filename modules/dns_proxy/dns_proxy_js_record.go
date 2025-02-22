package dns_proxy

import (
	"fmt"
	"net"

	"github.com/bettercap/bettercap/v2/log"
	"github.com/miekg/dns"
)

func NewJSResourceRecord(rr dns.RR) (jsRecord map[string]interface{}, err error) {
	header := rr.Header()

	jsRecord = map[string]interface{}{
		"Header": map[string]interface{}{
			"Class":  int64(header.Class),
			"Name":   header.Name,
			"Rrtype": int64(header.Rrtype),
			"Ttl":    int64(header.Ttl),
		},
	}

	switch rr := rr.(type) {
	case *dns.A:
		jsRecord["A"] = rr.A.String()
	case *dns.AAAA:
		jsRecord["AAAA"] = rr.AAAA.String()
	case *dns.APL:
		jsPrefixes := make([]map[string]interface{}, len(rr.Prefixes))
		for i, v := range rr.Prefixes {
			jsPrefixes[i] = map[string]interface{}{
				"Negation": v.Negation,
				"Network":  v.Network.String(),
			}
		}
		jsRecord["Prefixes"] = jsPrefixes
	case *dns.CNAME:
		jsRecord["Target"] = rr.Target
	case *dns.MB:
		jsRecord["Mb"] = rr.Mb
	case *dns.MD:
		jsRecord["Md"] = rr.Md
	case *dns.MF:
		jsRecord["Mf"] = rr.Mf
	case *dns.MG:
		jsRecord["Mg"] = rr.Mg
	case *dns.MR:
		jsRecord["Mr"] = rr.Mr
	case *dns.MX:
		jsRecord["Mx"] = rr.Mx
		jsRecord["Preference"] = int64(rr.Preference)
	case *dns.NULL:
		jsRecord["Data"] = rr.Data
	case *dns.SOA:
		jsRecord["Expire"] = int64(rr.Expire)
		jsRecord["Minttl"] = int64(rr.Minttl)
		jsRecord["Ns"] = rr.Ns
		jsRecord["Refresh"] = int64(rr.Refresh)
		jsRecord["Retry"] = int64(rr.Retry)
		jsRecord["Mbox"] = rr.Mbox
		jsRecord["Serial"] = int64(rr.Serial)
	case *dns.TXT:
		jsRecord["Txt"] = rr.Txt
	case *dns.SRV:
		jsRecord["Port"] = int64(rr.Port)
		jsRecord["Priority"] = int64(rr.Priority)
		jsRecord["Target"] = rr.Target
		jsRecord["Weight"] = int64(rr.Weight)
	case *dns.PTR:
		jsRecord["Ptr"] = rr.Ptr
	case *dns.NS:
		jsRecord["Ns"] = rr.Ns
	case *dns.DNAME:
		jsRecord["Target"] = rr.Target
	case *dns.AFSDB:
		jsRecord["Subtype"] = int64(rr.Subtype)
		jsRecord["Hostname"] = rr.Hostname
	case *dns.CAA:
		jsRecord["Flag"] = int64(rr.Flag)
		jsRecord["Tag"] = rr.Tag
		jsRecord["Value"] = rr.Value
	case *dns.HINFO:
		jsRecord["Cpu"] = rr.Cpu
		jsRecord["Os"] = rr.Os
	case *dns.MINFO:
		jsRecord["Email"] = rr.Email
		jsRecord["Rmail"] = rr.Rmail
	case *dns.ISDN:
		jsRecord["Address"] = rr.Address
		jsRecord["SubAddress"] = rr.SubAddress
	case *dns.KX:
		jsRecord["Exchanger"] = rr.Exchanger
		jsRecord["Preference"] = int64(rr.Preference)
	case *dns.LOC:
		jsRecord["Altitude"] = int64(rr.Altitude)
		jsRecord["HorizPre"] = int64(rr.HorizPre)
		jsRecord["Latitude"] = int64(rr.Latitude)
		jsRecord["Longitude"] = int64(rr.Longitude)
		jsRecord["Size"] = int64(rr.Size)
		jsRecord["Version"] = int64(rr.Version)
		jsRecord["VertPre"] = int64(rr.VertPre)
	case *dns.SSHFP:
		jsRecord["Algorithm"] = int64(rr.Algorithm)
		jsRecord["FingerPrint"] = rr.FingerPrint
		jsRecord["Type"] = int64(rr.Type)
	case *dns.TLSA:
		jsRecord["Certificate"] = rr.Certificate
		jsRecord["MatchingType"] = int64(rr.MatchingType)
		jsRecord["Selector"] = int64(rr.Selector)
		jsRecord["Usage"] = int64(rr.Usage)
	case *dns.CERT:
		jsRecord["Algorithm"] = int64(rr.Algorithm)
		jsRecord["Certificate"] = rr.Certificate
		jsRecord["KeyTag"] = int64(rr.KeyTag)
		jsRecord["Type"] = int64(rr.Type)
	case *dns.DS:
		jsRecord["Algorithm"] = int64(rr.Algorithm)
		jsRecord["Digest"] = rr.Digest
		jsRecord["DigestType"] = int64(rr.DigestType)
		jsRecord["KeyTag"] = int64(rr.KeyTag)
	case *dns.NAPTR:
		jsRecord["Order"] = int64(rr.Order)
		jsRecord["Preference"] = int64(rr.Preference)
		jsRecord["Flags"] = rr.Flags
		jsRecord["Service"] = rr.Service
		jsRecord["Regexp"] = rr.Regexp
		jsRecord["Replacement"] = rr.Replacement
	case *dns.RRSIG:
		jsRecord["Algorithm"] = int64(rr.Algorithm)
		jsRecord["Expiration"] = int64(rr.Expiration)
		jsRecord["Inception"] = int64(rr.Inception)
		jsRecord["KeyTag"] = int64(rr.KeyTag)
		jsRecord["Labels"] = int64(rr.Labels)
		jsRecord["OrigTtl"] = int64(rr.OrigTtl)
		jsRecord["Signature"] = rr.Signature
		jsRecord["SignerName"] = rr.SignerName
		jsRecord["TypeCovered"] = int64(rr.TypeCovered)
	case *dns.NSEC:
		jsRecord["NextDomain"] = rr.NextDomain
		jsRecord["TypeBitMap"] = uint16ArrayToInt64Array(rr.TypeBitMap)
	case *dns.NSEC3:
		jsRecord["Flags"] = int64(rr.Flags)
		jsRecord["Hash"] = int64(rr.Hash)
		jsRecord["HashLength"] = int64(rr.HashLength)
		jsRecord["Iterations"] = int64(rr.Iterations)
		jsRecord["NextDomain"] = rr.NextDomain
		jsRecord["Salt"] = rr.Salt
		jsRecord["SaltLength"] = int64(rr.SaltLength)
		jsRecord["TypeBitMap"] = uint16ArrayToInt64Array(rr.TypeBitMap)
	case *dns.NSEC3PARAM:
		jsRecord["Flags"] = int64(rr.Flags)
		jsRecord["Hash"] = int64(rr.Hash)
		jsRecord["Iterations"] = int64(rr.Iterations)
		jsRecord["Salt"] = rr.Salt
		jsRecord["SaltLength"] = int64(rr.SaltLength)
	case *dns.TKEY:
		jsRecord["Algorithm"] = rr.Algorithm
		jsRecord["Error"] = int64(rr.Error)
		jsRecord["Expiration"] = int64(rr.Expiration)
		jsRecord["Inception"] = int64(rr.Inception)
		jsRecord["Key"] = rr.Key
		jsRecord["KeySize"] = int64(rr.KeySize)
		jsRecord["Mode"] = int64(rr.Mode)
		jsRecord["OtherData"] = rr.OtherData
		jsRecord["OtherLen"] = int64(rr.OtherLen)
	case *dns.TSIG:
		jsRecord["Algorithm"] = rr.Algorithm
		jsRecord["Error"] = int64(rr.Error)
		jsRecord["Fudge"] = int64(rr.Fudge)
		jsRecord["MACSize"] = int64(rr.MACSize)
		jsRecord["MAC"] = rr.MAC
		jsRecord["OrigId"] = int64(rr.OrigId)
		jsRecord["OtherData"] = rr.OtherData
		jsRecord["OtherLen"] = int64(rr.OtherLen)
		jsRecord["TimeSigned"] = int64(rr.TimeSigned)
	case *dns.IPSECKEY:
		jsRecord["Algorithm"] = int64(rr.Algorithm)
		jsRecord["GatewayAddr"] = rr.GatewayAddr.String()
		jsRecord["GatewayHost"] = rr.GatewayHost
		jsRecord["GatewayType"] = int64(rr.GatewayType)
		jsRecord["Precedence"] = int64(rr.Precedence)
		jsRecord["PublicKey"] = rr.PublicKey
	case *dns.KEY:
		jsRecord["Flags"] = int64(rr.Flags)
		jsRecord["Protocol"] = int64(rr.Protocol)
		jsRecord["Algorithm"] = int64(rr.Algorithm)
		jsRecord["PublicKey"] = rr.PublicKey
	case *dns.CDS:
		jsRecord["KeyTag"] = int64(rr.KeyTag)
		jsRecord["Algorithm"] = int64(rr.Algorithm)
		jsRecord["DigestType"] = int64(rr.DigestType)
		jsRecord["Digest"] = rr.Digest
	case *dns.CDNSKEY:
		jsRecord["Algorithm"] = int64(rr.Algorithm)
		jsRecord["Flags"] = int64(rr.Flags)
		jsRecord["Protocol"] = int64(rr.Protocol)
		jsRecord["PublicKey"] = rr.PublicKey
	case *dns.NID:
		jsRecord["NodeID"] = rr.NodeID
		jsRecord["Preference"] = int64(rr.Preference)
	case *dns.L32:
		jsRecord["Locator32"] = rr.Locator32.String()
		jsRecord["Preference"] = int64(rr.Preference)
	case *dns.L64:
		jsRecord["Locator64"] = rr.Locator64
		jsRecord["Preference"] = int64(rr.Preference)
	case *dns.LP:
		jsRecord["Fqdn"] = rr.Fqdn
		jsRecord["Preference"] = int16(rr.Preference)
	case *dns.GPOS:
		jsRecord["Altitude"] = rr.Altitude
		jsRecord["Latitude"] = rr.Latitude
		jsRecord["Longitude"] = rr.Longitude
	case *dns.RP:
		jsRecord["Mbox"] = rr.Mbox
		jsRecord["Txt"] = rr.Txt
	case *dns.RKEY:
		jsRecord["Algorithm"] = int64(rr.Algorithm)
		jsRecord["Flags"] = int64(rr.Flags)
		jsRecord["Protocol"] = int64(rr.Protocol)
		jsRecord["PublicKey"] = rr.PublicKey
	case *dns.SMIMEA:
		jsRecord["Certificate"] = rr.Certificate
		jsRecord["MatchingType"] = int64(rr.MatchingType)
		jsRecord["Selector"] = int64(rr.Selector)
		jsRecord["Usage"] = int64(rr.Usage)
	case *dns.AMTRELAY:
		jsRecord["GatewayAddr"] = rr.GatewayAddr.String()
		jsRecord["GatewayHost"] = rr.GatewayHost
		jsRecord["GatewayType"] = int64(rr.GatewayType)
		jsRecord["Precedence"] = int64(rr.Precedence)
	case *dns.AVC:
		jsRecord["Txt"] = rr.Txt
	case *dns.URI:
		jsRecord["Priority"] = int64(rr.Priority)
		jsRecord["Weight"] = int64(rr.Weight)
		jsRecord["Target"] = rr.Target
	case *dns.EUI48:
		jsRecord["Address"] = rr.Address
	case *dns.EUI64:
		jsRecord["Address"] = rr.Address
	case *dns.GID:
		jsRecord["Gid"] = int64(rr.Gid)
	case *dns.UID:
		jsRecord["Uid"] = int64(rr.Uid)
	case *dns.UINFO:
		jsRecord["Uinfo"] = rr.Uinfo
	case *dns.SPF:
		jsRecord["Txt"] = rr.Txt
	case *dns.HTTPS:
		jsRecord["Priority"] = int64(rr.Priority)
		jsRecord["Target"] = rr.Target
		kvs := rr.Value
		var jsKvs []map[string]interface{}
		for _, kv := range kvs {
			jsKv, err := NewJSSVCBKeyValue(kv)
			if err != nil {
				log.Error(err.Error())
				continue
			}
			jsKvs = append(jsKvs, jsKv)
		}
		jsRecord["Value"] = jsKvs
	case *dns.SVCB:
		jsRecord["Priority"] = int64(rr.Priority)
		jsRecord["Target"] = rr.Target
		kvs := rr.Value
		jsKvs := make([]map[string]interface{}, len(kvs))
		for i, kv := range kvs {
			jsKv, err := NewJSSVCBKeyValue(kv)
			if err != nil {
				log.Error(err.Error())
				continue
			}
			jsKvs[i] = jsKv
		}
		jsRecord["Value"] = jsKvs
	case *dns.ZONEMD:
		jsRecord["Digest"] = rr.Digest
		jsRecord["Hash"] = int64(rr.Hash)
		jsRecord["Scheme"] = int64(rr.Scheme)
		jsRecord["Serial"] = int64(rr.Serial)
	case *dns.CSYNC:
		jsRecord["Flags"] = int64(rr.Flags)
		jsRecord["Serial"] = int64(rr.Serial)
		jsRecord["TypeBitMap"] = uint16ArrayToInt64Array(rr.TypeBitMap)
	case *dns.OPENPGPKEY:
		jsRecord["PublicKey"] = rr.PublicKey
	case *dns.TALINK:
		jsRecord["NextName"] = rr.NextName
		jsRecord["PreviousName"] = rr.PreviousName
	case *dns.NINFO:
		jsRecord["ZSData"] = rr.ZSData
	case *dns.DHCID:
		jsRecord["Digest"] = rr.Digest
	case *dns.DNSKEY:
		jsRecord["Flags"] = int64(rr.Flags)
		jsRecord["Protocol"] = int64(rr.Protocol)
		jsRecord["Algorithm"] = int64(rr.Algorithm)
		jsRecord["PublicKey"] = rr.PublicKey
	case *dns.HIP:
		jsRecord["Hit"] = rr.Hit
		jsRecord["HitLength"] = int64(rr.HitLength)
		jsRecord["PublicKey"] = rr.PublicKey
		jsRecord["PublicKeyAlgorithm"] = int64(rr.PublicKeyAlgorithm)
		jsRecord["PublicKeyLength"] = int64(rr.PublicKeyLength)
		jsRecord["RendezvousServers"] = rr.RendezvousServers
	case *dns.OPT:
		options := rr.Option
		jsOptions := make([]map[string]interface{}, len(options))
		for i, option := range options {
			jsOption, err := NewJSEDNS0(option)
			if err != nil {
				log.Error(err.Error())
				continue
			}
			jsOptions[i] = jsOption
		}
		jsRecord["Option"] = jsOptions
	case *dns.NIMLOC:
		jsRecord["Locator"] = rr.Locator
	case *dns.EID:
		jsRecord["Endpoint"] = rr.Endpoint
	case *dns.NXT:
		jsRecord["NextDomain"] = rr.NextDomain
		jsRecord["TypeBitMap"] = uint16ArrayToInt64Array(rr.TypeBitMap)
	case *dns.PX:
		jsRecord["Mapx400"] = rr.Mapx400
		jsRecord["Map822"] = rr.Map822
		jsRecord["Preference"] = int64(rr.Preference)
	case *dns.SIG:
		jsRecord["Algorithm"] = int64(rr.Algorithm)
		jsRecord["Expiration"] = int64(rr.Expiration)
		jsRecord["Inception"] = int64(rr.Inception)
		jsRecord["KeyTag"] = int64(rr.KeyTag)
		jsRecord["Labels"] = int64(rr.Labels)
		jsRecord["OrigTtl"] = int64(rr.OrigTtl)
		jsRecord["Signature"] = rr.Signature
		jsRecord["SignerName"] = rr.SignerName
		jsRecord["TypeCovered"] = int64(rr.TypeCovered)
	case *dns.RT:
		jsRecord["Host"] = rr.Host
		jsRecord["Preference"] = int64(rr.Preference)
	case *dns.NSAPPTR:
		jsRecord["Ptr"] = rr.Ptr
	case *dns.X25:
		jsRecord["PSDNAddress"] = rr.PSDNAddress
	case *dns.RFC3597:
		jsRecord["Rdata"] = rr.Rdata
	// case *dns.ATMA:
	// case *dns.WKS:
	// case *dns.DOA:
	// case *dns.SINK:
	default:
		if header.Rrtype == dns.TypeNone {
			break
		}
		return nil, fmt.Errorf("error creating JSResourceRecord: unknown type: %d", header.Rrtype)
	}

	return jsRecord, nil
}

func ToRR(jsRecord map[string]interface{}) (rr dns.RR, err error) {
	jsHeader := jsPropToMap(jsRecord, "Header")

	header := dns.RR_Header{
		Class:  jsPropToUint16(jsHeader, "Class"),
		Name:   jsPropToString(jsHeader, "Name"),
		Rrtype: jsPropToUint16(jsHeader, "Rrtype"),
		Ttl:    jsPropToUint32(jsHeader, "Ttl"),
	}

	switch header.Rrtype {
	case dns.TypeNone:
		break
	case dns.TypeA:
		rr = &dns.A{
			Hdr: header,
			A:   net.ParseIP(jsPropToString(jsRecord, "A")),
		}
	case dns.TypeAAAA:
		rr = &dns.AAAA{
			Hdr:  header,
			AAAA: net.ParseIP(jsPropToString(jsRecord, "AAAA")),
		}
	case dns.TypeAPL:
		jsPrefixes := jsRecord["Prefixes"].([]map[string]interface{})
		prefixes := make([]dns.APLPrefix, len(jsPrefixes))
		for i, jsPrefix := range jsPrefixes {
			jsNetwork := jsPrefix["Network"].(string)
			_, network, err := net.ParseCIDR(jsNetwork)
			if err != nil {
				log.Error("error parsing CIDR: %s", jsNetwork)
				continue
			}
			prefixes[i] = dns.APLPrefix{
				Negation: jsPrefix["Negation"].(bool),
				Network:  *network,
			}
		}
		rr = &dns.APL{
			Hdr:      header,
			Prefixes: prefixes,
		}
	case dns.TypeCNAME:
		rr = &dns.CNAME{
			Hdr:    header,
			Target: jsPropToString(jsRecord, "Target"),
		}
	case dns.TypeMB:
		rr = &dns.MB{
			Hdr: header,
			Mb:  jsPropToString(jsRecord, "Mb"),
		}
	case dns.TypeMD:
		rr = &dns.MD{
			Hdr: header,
			Md:  jsPropToString(jsRecord, "Md"),
		}
	case dns.TypeMF:
		rr = &dns.MF{
			Hdr: header,
			Mf:  jsPropToString(jsRecord, "Mf"),
		}
	case dns.TypeMG:
		rr = &dns.MG{
			Hdr: header,
			Mg:  jsPropToString(jsRecord, "Mg"),
		}
	case dns.TypeMR:
		rr = &dns.MR{
			Hdr: header,
			Mr:  jsPropToString(jsRecord, "Mr"),
		}
	case dns.TypeMX:
		rr = &dns.MX{
			Hdr:        header,
			Mx:         jsPropToString(jsRecord, "Mx"),
			Preference: jsPropToUint16(jsRecord, "Preference"),
		}
	case dns.TypeNULL:
		rr = &dns.NULL{
			Hdr:  header,
			Data: jsPropToString(jsRecord, "Data"),
		}
	case dns.TypeSOA:
		rr = &dns.SOA{
			Hdr:     header,
			Expire:  jsPropToUint32(jsRecord, "Expire"),
			Mbox:    jsPropToString(jsRecord, "Mbox"),
			Minttl:  jsPropToUint32(jsRecord, "Minttl"),
			Ns:      jsPropToString(jsRecord, "Ns"),
			Refresh: jsPropToUint32(jsRecord, "Refresh"),
			Retry:   jsPropToUint32(jsRecord, "Retry"),
			Serial:  jsPropToUint32(jsRecord, "Serial"),
		}
	case dns.TypeTXT:
		rr = &dns.TXT{
			Hdr: header,
			Txt: jsPropToStringArray(jsRecord, "Txt"),
		}
	case dns.TypeSRV:
		rr = &dns.SRV{
			Hdr:      header,
			Port:     jsPropToUint16(jsRecord, "Port"),
			Priority: jsPropToUint16(jsRecord, "Priority"),
			Target:   jsPropToString(jsRecord, "Target"),
			Weight:   jsPropToUint16(jsRecord, "Weight"),
		}
	case dns.TypePTR:
		rr = &dns.PTR{
			Hdr: header,
			Ptr: jsPropToString(jsRecord, "Ptr"),
		}
	case dns.TypeNS:
		rr = &dns.NS{
			Hdr: header,
			Ns:  jsPropToString(jsRecord, "Ns"),
		}
	case dns.TypeDNAME:
		rr = &dns.DNAME{
			Hdr:    header,
			Target: jsPropToString(jsRecord, "Target"),
		}
	case dns.TypeAFSDB:
		rr = &dns.AFSDB{
			Hdr:      header,
			Hostname: jsPropToString(jsRecord, "Hostname"),
			Subtype:  jsPropToUint16(jsRecord, "Subtype"),
		}
	case dns.TypeCAA:
		rr = &dns.CAA{
			Hdr:   header,
			Flag:  jsPropToUint8(jsRecord, "Flag"),
			Tag:   jsPropToString(jsRecord, "Tag"),
			Value: jsPropToString(jsRecord, "Value"),
		}
	case dns.TypeHINFO:
		rr = &dns.HINFO{
			Hdr: header,
			Cpu: jsPropToString(jsRecord, "Cpu"),
			Os:  jsPropToString(jsRecord, "Os"),
		}
	case dns.TypeMINFO:
		rr = &dns.MINFO{
			Hdr:   header,
			Email: jsPropToString(jsRecord, "Email"),
			Rmail: jsPropToString(jsRecord, "Rmail"),
		}
	case dns.TypeISDN:
		rr = &dns.ISDN{
			Hdr:        header,
			Address:    jsPropToString(jsRecord, "Address"),
			SubAddress: jsPropToString(jsRecord, "SubAddress"),
		}
	case dns.TypeKX:
		rr = &dns.KX{
			Hdr:        header,
			Preference: jsPropToUint16(jsRecord, "Preference"),
			Exchanger:  jsPropToString(jsRecord, "Exchanger"),
		}
	case dns.TypeLOC:
		rr = &dns.LOC{
			Hdr:       header,
			Version:   jsPropToUint8(jsRecord, "Version"),
			Size:      jsPropToUint8(jsRecord, "Size"),
			HorizPre:  jsPropToUint8(jsRecord, "HorizPre"),
			VertPre:   jsPropToUint8(jsRecord, "VertPre"),
			Latitude:  jsPropToUint32(jsRecord, "Latitude"),
			Longitude: jsPropToUint32(jsRecord, "Longitude"),
			Altitude:  jsPropToUint32(jsRecord, "Altitude"),
		}
	case dns.TypeSSHFP:
		rr = &dns.SSHFP{
			Hdr:         header,
			Algorithm:   jsPropToUint8(jsRecord, "Algorithm"),
			FingerPrint: jsPropToString(jsRecord, "FingerPrint"),
			Type:        jsPropToUint8(jsRecord, "Type"),
		}
	case dns.TypeTLSA:
		rr = &dns.TLSA{
			Hdr:          header,
			Certificate:  jsPropToString(jsRecord, "Certificate"),
			MatchingType: jsPropToUint8(jsRecord, "MatchingType"),
			Selector:     jsPropToUint8(jsRecord, "Selector"),
			Usage:        jsPropToUint8(jsRecord, "Usage"),
		}
	case dns.TypeCERT:
		rr = &dns.CERT{
			Hdr:         header,
			Algorithm:   jsPropToUint8(jsRecord, "Algorithm"),
			Certificate: jsPropToString(jsRecord, "Certificate"),
			KeyTag:      jsPropToUint16(jsRecord, "KeyTag"),
			Type:        jsPropToUint16(jsRecord, "Type"),
		}
	case dns.TypeDS:
		rr = &dns.DS{
			Hdr:        header,
			Algorithm:  jsPropToUint8(jsRecord, "Algorithm"),
			Digest:     jsPropToString(jsRecord, "Digest"),
			DigestType: jsPropToUint8(jsRecord, "DigestType"),
			KeyTag:     jsPropToUint16(jsRecord, "KeyTag"),
		}
	case dns.TypeNAPTR:
		rr = &dns.NAPTR{
			Hdr:         header,
			Flags:       jsPropToString(jsRecord, "Flags"),
			Order:       jsPropToUint16(jsRecord, "Order"),
			Preference:  jsPropToUint16(jsRecord, "Preference"),
			Regexp:      jsPropToString(jsRecord, "Regexp"),
			Replacement: jsPropToString(jsRecord, "Replacement"),
			Service:     jsPropToString(jsRecord, "Service"),
		}
	case dns.TypeRRSIG:
		rr = &dns.RRSIG{
			Hdr:         header,
			Algorithm:   jsPropToUint8(jsRecord, "Algorithm"),
			Expiration:  jsPropToUint32(jsRecord, "Expiration"),
			Inception:   jsPropToUint32(jsRecord, "Inception"),
			KeyTag:      jsPropToUint16(jsRecord, "KeyTag"),
			Labels:      jsPropToUint8(jsRecord, "Labels"),
			OrigTtl:     jsPropToUint32(jsRecord, "OrigTtl"),
			Signature:   jsPropToString(jsRecord, "Signature"),
			SignerName:  jsPropToString(jsRecord, "SignerName"),
			TypeCovered: jsPropToUint16(jsRecord, "TypeCovered"),
		}
	case dns.TypeNSEC:
		rr = &dns.NSEC{
			Hdr:        header,
			NextDomain: jsPropToString(jsRecord, "NextDomain"),
			TypeBitMap: jsPropToUint16Array(jsRecord, "TypeBitMap"),
		}
	case dns.TypeNSEC3:
		rr = &dns.NSEC3{
			Hdr:        header,
			Flags:      jsPropToUint8(jsRecord, "Flags"),
			Hash:       jsPropToUint8(jsRecord, "Hash"),
			HashLength: jsPropToUint8(jsRecord, "HashLength"),
			Iterations: jsPropToUint16(jsRecord, "Iterations"),
			NextDomain: jsPropToString(jsRecord, "NextDomain"),
			Salt:       jsPropToString(jsRecord, "Salt"),
			SaltLength: jsPropToUint8(jsRecord, "SaltLength"),
			TypeBitMap: jsPropToUint16Array(jsRecord, "TypeBitMap"),
		}
	case dns.TypeNSEC3PARAM:
		rr = &dns.NSEC3PARAM{
			Hdr:        header,
			Flags:      jsPropToUint8(jsRecord, "Flags"),
			Hash:       jsPropToUint8(jsRecord, "Hash"),
			Iterations: jsPropToUint16(jsRecord, "Iterations"),
			Salt:       jsPropToString(jsRecord, "Salt"),
			SaltLength: jsPropToUint8(jsRecord, "SaltLength"),
		}
	case dns.TypeTKEY:
		rr = &dns.TKEY{
			Hdr:        header,
			Algorithm:  jsPropToString(jsRecord, "Algorithm"),
			Error:      jsPropToUint16(jsRecord, "Error"),
			Expiration: jsPropToUint32(jsRecord, "Expiration"),
			Inception:  jsPropToUint32(jsRecord, "Inception"),
			Key:        jsPropToString(jsRecord, "Key"),
			KeySize:    jsPropToUint16(jsRecord, "KeySize"),
			Mode:       jsPropToUint16(jsRecord, "Mode"),
			OtherData:  jsPropToString(jsRecord, "OtherData"),
			OtherLen:   jsPropToUint16(jsRecord, "OtherLen"),
		}
	case dns.TypeTSIG:
		rr = &dns.TSIG{
			Hdr:        header,
			Algorithm:  jsPropToString(jsRecord, "Algorithm"),
			Error:      jsPropToUint16(jsRecord, "Error"),
			Fudge:      jsPropToUint16(jsRecord, "Fudge"),
			MACSize:    jsPropToUint16(jsRecord, "MACSize"),
			MAC:        jsPropToString(jsRecord, "MAC"),
			OrigId:     jsPropToUint16(jsRecord, "OrigId"),
			OtherData:  jsPropToString(jsRecord, "OtherData"),
			OtherLen:   jsPropToUint16(jsRecord, "OtherLen"),
			TimeSigned: jsPropToUint64(jsRecord, "TimeSigned"),
		}
	case dns.TypeIPSECKEY:
		rr = &dns.IPSECKEY{
			Hdr:         header,
			Algorithm:   jsPropToUint8(jsRecord, "Algorithm"),
			GatewayAddr: net.IP(jsPropToString(jsRecord, "GatewayAddr")),
			GatewayHost: jsPropToString(jsRecord, "GatewayHost"),
			GatewayType: jsPropToUint8(jsRecord, "GatewayType"),
			Precedence:  jsPropToUint8(jsRecord, "Precedence"),
			PublicKey:   jsPropToString(jsRecord, "PublicKey"),
		}
	case dns.TypeKEY:
		rr = &dns.KEY{
			DNSKEY: dns.DNSKEY{
				Hdr:       header,
				Algorithm: jsPropToUint8(jsRecord, "Algorithm"),
				Flags:     jsPropToUint16(jsRecord, "Flags"),
				Protocol:  jsPropToUint8(jsRecord, "Protocol"),
				PublicKey: jsPropToString(jsRecord, "PublicKey"),
			},
		}
	case dns.TypeCDS:
		rr = &dns.CDS{
			DS: dns.DS{
				Hdr:        header,
				KeyTag:     jsPropToUint16(jsRecord, "KeyTag"),
				Algorithm:  jsPropToUint8(jsRecord, "Algorithm"),
				DigestType: jsPropToUint8(jsRecord, "DigestType"),
				Digest:     jsPropToString(jsRecord, "Digest"),
			},
		}
	case dns.TypeCDNSKEY:
		rr = &dns.CDNSKEY{
			DNSKEY: dns.DNSKEY{
				Hdr:       header,
				Algorithm: jsPropToUint8(jsRecord, "Algorithm"),
				Flags:     jsPropToUint16(jsRecord, "Flags"),
				Protocol:  jsPropToUint8(jsRecord, "Protocol"),
				PublicKey: jsPropToString(jsRecord, "PublicKey"),
			},
		}
	case dns.TypeNID:
		rr = &dns.NID{
			Hdr:        header,
			NodeID:     jsPropToUint64(jsRecord, "NodeID"),
			Preference: jsPropToUint16(jsRecord, "Preference"),
		}
	case dns.TypeL32:
		rr = &dns.L32{
			Hdr:        header,
			Locator32:  net.IP(jsPropToString(jsRecord, "Locator32")),
			Preference: jsPropToUint16(jsRecord, "Preference"),
		}
	case dns.TypeL64:
		rr = &dns.L64{
			Hdr:        header,
			Locator64:  jsPropToUint64(jsRecord, "Locator64"),
			Preference: jsPropToUint16(jsRecord, "Preference"),
		}
	case dns.TypeLP:
		rr = &dns.LP{
			Hdr:        header,
			Fqdn:       jsPropToString(jsRecord, "Fqdn"),
			Preference: jsPropToUint16(jsRecord, "Preference"),
		}
	case dns.TypeGPOS:
		rr = &dns.GPOS{
			Hdr:       header,
			Altitude:  jsPropToString(jsRecord, "Altitude"),
			Latitude:  jsPropToString(jsRecord, "Latitude"),
			Longitude: jsPropToString(jsRecord, "Longitude"),
		}
	case dns.TypeRP:
		rr = &dns.RP{
			Hdr:  header,
			Mbox: jsPropToString(jsRecord, "Mbox"),
			Txt:  jsPropToString(jsRecord, "Txt"),
		}
	case dns.TypeRKEY:
		rr = &dns.RKEY{
			Hdr:       header,
			Algorithm: jsPropToUint8(jsRecord, "Algorithm"),
			Flags:     jsPropToUint16(jsRecord, "Flags"),
			Protocol:  jsPropToUint8(jsRecord, "Protocol"),
			PublicKey: jsPropToString(jsRecord, "PublicKey"),
		}
	case dns.TypeSMIMEA:
		rr = &dns.SMIMEA{
			Hdr:          header,
			Certificate:  jsPropToString(jsRecord, "Certificate"),
			MatchingType: jsPropToUint8(jsRecord, "MatchingType"),
			Selector:     jsPropToUint8(jsRecord, "Selector"),
			Usage:        jsPropToUint8(jsRecord, "Usage"),
		}
	case dns.TypeAMTRELAY:
		rr = &dns.AMTRELAY{
			Hdr:         header,
			GatewayAddr: net.IP(jsPropToString(jsRecord, "GatewayAddr")),
			GatewayHost: jsPropToString(jsRecord, "GatewayHost"),
			GatewayType: jsPropToUint8(jsRecord, "GatewayType"),
			Precedence:  jsPropToUint8(jsRecord, "Precedence"),
		}
	case dns.TypeAVC:
		rr = &dns.AVC{
			Hdr: header,
			Txt: jsPropToStringArray(jsRecord, "Txt"),
		}
	case dns.TypeURI:
		rr = &dns.URI{
			Hdr:      header,
			Priority: jsPropToUint16(jsRecord, "Priority"),
			Weight:   jsPropToUint16(jsRecord, "Weight"),
			Target:   jsPropToString(jsRecord, "Target"),
		}
	case dns.TypeEUI48:
		rr = &dns.EUI48{
			Hdr:     header,
			Address: jsPropToUint64(jsRecord, "Address"),
		}
	case dns.TypeEUI64:
		rr = &dns.EUI64{
			Hdr:     header,
			Address: jsPropToUint64(jsRecord, "Address"),
		}
	case dns.TypeGID:
		rr = &dns.GID{
			Hdr: header,
			Gid: jsPropToUint32(jsRecord, "Gid"),
		}
	case dns.TypeUID:
		rr = &dns.UID{
			Hdr: header,
			Uid: jsPropToUint32(jsRecord, "Uid"),
		}
	case dns.TypeUINFO:
		rr = &dns.UINFO{
			Hdr:   header,
			Uinfo: jsPropToString(jsRecord, "Uinfo"),
		}
	case dns.TypeSPF:
		rr = &dns.SPF{
			Hdr: header,
			Txt: jsPropToStringArray(jsRecord, "Txt"),
		}
	case dns.TypeHTTPS:
		jsKvs := jsPropToMapArray(jsRecord, "Value")
		var kvs []dns.SVCBKeyValue
		for _, jsKv := range jsKvs {
			kv, err := ToSVCBKeyValue(jsKv)
			if err != nil {
				log.Error(err.Error())
				continue
			}
			kvs = append(kvs, kv)
		}
		rr = &dns.HTTPS{
			SVCB: dns.SVCB{
				Hdr:      header,
				Priority: jsPropToUint16(jsRecord, "Priority"),
				Target:   jsPropToString(jsRecord, "Target"),
				Value:    kvs,
			},
		}
	case dns.TypeSVCB:
		jsKvs := jsPropToMapArray(jsRecord, "Value")
		var kvs []dns.SVCBKeyValue
		for _, jsKv := range jsKvs {
			kv, err := ToSVCBKeyValue(jsKv)
			if err != nil {
				log.Error(err.Error())
				continue
			}
			kvs = append(kvs, kv)
		}
		rr = &dns.SVCB{
			Hdr:      header,
			Priority: jsPropToUint16(jsRecord, "Priority"),
			Target:   jsPropToString(jsRecord, "Target"),
			Value:    kvs,
		}
	case dns.TypeZONEMD:
		rr = &dns.ZONEMD{
			Hdr:    header,
			Digest: jsPropToString(jsRecord, "Digest"),
			Hash:   jsPropToUint8(jsRecord, "Hash"),
			Scheme: jsPropToUint8(jsRecord, "Scheme"),
			Serial: jsPropToUint32(jsRecord, "Serial"),
		}
	case dns.TypeCSYNC:
		rr = &dns.CSYNC{
			Hdr:        header,
			Flags:      jsPropToUint16(jsRecord, "Flags"),
			Serial:     jsPropToUint32(jsRecord, "Serial"),
			TypeBitMap: jsPropToUint16Array(jsRecord, "TypeBitMap"),
		}
	case dns.TypeOPENPGPKEY:
		rr = &dns.OPENPGPKEY{
			Hdr:       header,
			PublicKey: jsPropToString(jsRecord, "PublicKey"),
		}
	case dns.TypeTALINK:
		rr = &dns.TALINK{
			Hdr:          header,
			NextName:     jsPropToString(jsRecord, "NextName"),
			PreviousName: jsPropToString(jsRecord, "PreviousName"),
		}
	case dns.TypeNINFO:
		rr = &dns.NINFO{
			Hdr:    header,
			ZSData: jsPropToStringArray(jsRecord, "ZSData"),
		}
	case dns.TypeDHCID:
		rr = &dns.DHCID{
			Hdr:    header,
			Digest: jsPropToString(jsRecord, "Digest"),
		}
	case dns.TypeDNSKEY:
		rr = &dns.DNSKEY{
			Hdr:       header,
			Algorithm: jsPropToUint8(jsRecord, "Algorithm"),
			Flags:     jsPropToUint16(jsRecord, "Flags"),
			Protocol:  jsPropToUint8(jsRecord, "Protocol"),
			PublicKey: jsPropToString(jsRecord, "PublicKey"),
		}
	case dns.TypeHIP:
		rr = &dns.HIP{
			Hdr:                header,
			Hit:                jsPropToString(jsRecord, "Hit"),
			HitLength:          jsPropToUint8(jsRecord, "HitLength"),
			PublicKey:          jsPropToString(jsRecord, "PublicKey"),
			PublicKeyAlgorithm: jsPropToUint8(jsRecord, "PublicKeyAlgorithm"),
			PublicKeyLength:    jsPropToUint16(jsRecord, "PublicKeyLength"),
			RendezvousServers:  jsPropToStringArray(jsRecord, "RendezvousServers"),
		}
	case dns.TypeOPT:
		jsOptions := jsPropToMapArray(jsRecord, "Option")
		var options []dns.EDNS0
		for _, jsOption := range jsOptions {
			option, err := ToEDNS0(jsOption)
			if err != nil {
				log.Error(err.Error())
				continue
			}
			options = append(options, option)
		}
		rr = &dns.OPT{
			Hdr:    header,
			Option: options,
		}
	case dns.TypeNIMLOC:
		rr = &dns.NIMLOC{
			Hdr:     header,
			Locator: jsPropToString(jsRecord, "Locator"),
		}
	case dns.TypeEID:
		rr = &dns.EID{
			Hdr:      header,
			Endpoint: jsPropToString(jsRecord, "Endpoint"),
		}
	case dns.TypeNXT:
		rr = &dns.NXT{
			NSEC: dns.NSEC{
				Hdr:        header,
				NextDomain: jsPropToString(jsRecord, "NextDomain"),
				TypeBitMap: jsPropToUint16Array(jsRecord, "TypeBitMap"),
			},
		}
	case dns.TypePX:
		rr = &dns.PX{
			Hdr:        header,
			Mapx400:    jsPropToString(jsRecord, "Mapx400"),
			Map822:     jsPropToString(jsRecord, "Map822"),
			Preference: jsPropToUint16(jsRecord, "Preference"),
		}
	case dns.TypeSIG:
		rr = &dns.SIG{
			RRSIG: dns.RRSIG{
				Hdr:         header,
				Algorithm:   jsPropToUint8(jsRecord, "Algorithm"),
				Expiration:  jsPropToUint32(jsRecord, "Expiration"),
				Inception:   jsPropToUint32(jsRecord, "Inception"),
				KeyTag:      jsPropToUint16(jsRecord, "KeyTag"),
				Labels:      jsPropToUint8(jsRecord, "Labels"),
				OrigTtl:     jsPropToUint32(jsRecord, "OrigTtl"),
				Signature:   jsPropToString(jsRecord, "Signature"),
				SignerName:  jsPropToString(jsRecord, "SignerName"),
				TypeCovered: jsPropToUint16(jsRecord, "TypeCovered"),
			},
		}
	case dns.TypeRT:
		rr = &dns.RT{
			Hdr:        header,
			Host:       jsPropToString(jsRecord, "Host"),
			Preference: jsPropToUint16(jsRecord, "Preference"),
		}
	case dns.TypeNSAPPTR:
		rr = &dns.NSAPPTR{
			Hdr: header,
			Ptr: jsPropToString(jsRecord, "Ptr"),
		}
	case dns.TypeX25:
		rr = &dns.X25{
			Hdr:         header,
			PSDNAddress: jsPropToString(jsRecord, "PSDNAddress"),
		}
	// case dns.TypeATMA:
	// case dns.TypeWKS:
	// case dns.TypeDOA:
	// case dns.TypeSINK:
	default:
		if rdata, ok := jsRecord["Rdata"].(string); ok {
			rr = &dns.RFC3597{
				Hdr:   header,
				Rdata: rdata,
			}
		} else {
			return nil, fmt.Errorf("error converting to dns.RR: unknown type: %d", header.Rrtype)
		}
	}

	return rr, nil
}
