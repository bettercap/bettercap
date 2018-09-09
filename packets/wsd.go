package packets

import (
	"net"
)

const (
	WSDPort = 3702
)

var (
	WSDDestIP           = net.ParseIP("239.255.255.250")
	WSDDiscoveryPayload = []byte("<?xml version=\"1.0\" encoding=\"utf-8\" ?>" +
		"<soap:Envelope" +
		" xmlns:soap=\"http://www.w3.org/2003/05/soap-envelope\"" +
		" xmlns:wsa=\"http://schemas.xmlsoap.org/ws/2004/08/addressing\"" +
		" xmlns:wsd=\"http://schemas.xmlsoap.org/ws/2005/04/discovery\"" +
		" xmlns:wsdp=\"http://schemas.xmlsoap.org/ws/2006/02/devprof\">" +
		"<soap:Header>" +
		"<wsa:To>urn:schemas-xmlsoap-org:ws:2005:04:discovery</wsa:To>" +
		"<wsa:Action>http://schemas.xmlsoap.org/ws/2005/04/discovery/Probe</wsa:Action>" +
		"<wsa:MessageID>urn:uuid:05a0036e-dcc8-4db8-98b6-0ceeee60a6d9</wsa:MessageID>" +
		"</soap:Header>" +
		"<soap:Body>" +
		"<wsd:Probe/>" +
		"</soap:Body>" +
		"</env:Envelope>")
)
