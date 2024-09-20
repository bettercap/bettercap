package zerogod

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/evilsocket/islazy/ops"
	"github.com/evilsocket/islazy/tui"

	"github.com/phin1x/go-ipp"
)

var IPP_REQUEST_NAMES = map[int16]string{
	// https://tools.ietf.org/html/rfc2911#section-4.4.15
	0x0002: "Print-Job",
	0x0003: "Print-URI",
	0x0004: "Validate-Job",
	0x0005: "Create-Job",
	0x0006: "Send-Document",
	0x0007: "Send-URI",
	0x0008: "Cancel-Job",
	0x0009: "Get-Job-Attributes",
	0x000A: "Get-Jobs",
	0x000B: "Get-Printer-Attributes",
	0x000C: "Hold-Job",
	0x000D: "Release-Job",
	0x000E: "Restart-Job",
	0x0010: "Pause-Printer",
	0x0011: "Resume-Printer",
	0x0012: "Purge-Jobs",
	// https://web.archive.org/web/20061024184939/http://uw714doc.sco.com/en/cups/ipp.html
	0x4001: "CUPS-Get-Default",
	0x4002: "CUPS-Get-Printers",
	0x4003: "CUPS-Add-Modify-Printer",
	0x4004: "CUPS-Delete-Printer",
	0x4005: "CUPS-Get-Classes",
	0x4006: "CUPS-Add-Modify-Class",
	0x4007: "CUPS-Delete-Class",
	0x4008: "CUPS-Accept-Jobs",
	0x4009: "CUPS-Reject-Jobs",
	0x400A: "CUPS-Set-Default",
	0x400B: "CUPS-Get-Devices",
	0x400C: "CUPS-Get-PPDs",
	0x400D: "CUPS-Move-Job",
}

var IPP_USER_ATTRIBUTES = map[string]string{
	"printer-name":               "PRINTER_NAME",
	"printer-info":               "PRINTER_INFO",
	"printer-make-and-model":     "PRINTER_MAKE PRINTER_MODEL",
	"printer-location":           "PRINTER_LOCATION",
	"printer-privacy-policy-uri": "https://www.bettercap.org/",
	"ppd-name":                   "everywhere",
}

func init() {
	ipp.AttributeTagMapping["printer-uri-supported"] = ipp.TagUri
	ipp.AttributeTagMapping["uri-authentication-supported"] = ipp.TagKeyword
	ipp.AttributeTagMapping["uri-security-supported"] = ipp.TagKeyword
	ipp.AttributeTagMapping["printer-name"] = ipp.TagName
	ipp.AttributeTagMapping["printer-info"] = ipp.TagText
	ipp.AttributeTagMapping["printer-make-and-model"] = ipp.TagText
	ipp.AttributeTagMapping["printer-state"] = ipp.TagEnum
	ipp.AttributeTagMapping["printer-state-reasons"] = ipp.TagKeyword
	ipp.AttributeTagMapping["ipp-versions-supported"] = ipp.TagKeyword
	ipp.AttributeTagMapping["operations-supported"] = ipp.TagEnum
	ipp.AttributeTagMapping["multiple-document-jobs-supported"] = ipp.TagBoolean
	ipp.AttributeTagMapping["charset-configured"] = ipp.TagCharset
	ipp.AttributeTagMapping["charset-supported"] = ipp.TagCharset
	ipp.AttributeTagMapping["natural-language-configured"] = ipp.TagLanguage
	ipp.AttributeTagMapping["generated-natural-language-supported"] = ipp.TagLanguage
	ipp.AttributeTagMapping["document-format-default"] = ipp.TagMimeType
	ipp.AttributeTagMapping["document-format-supported"] = ipp.TagMimeType
	ipp.AttributeTagMapping["printer-is-accepting-jobs"] = ipp.TagBoolean
	ipp.AttributeTagMapping["queued-job-count"] = ipp.TagInteger
	ipp.AttributeTagMapping["pdl-override-supported"] = ipp.TagKeyword
	ipp.AttributeTagMapping["printer-up-time"] = ipp.TagInteger
	ipp.AttributeTagMapping["compression-supported"] = ipp.TagKeyword
	ipp.AttributeTagMapping["printer-privacy-policy-uri"] = ipp.TagUri
	ipp.AttributeTagMapping["printer-location"] = ipp.TagText
	ipp.AttributeTagMapping["ppd-name"] = ipp.TagName
}

func ippClientHandler(ctx *HandlerContext) {
	defer ctx.client.Close()

	buf := make([]byte, 4096)

	// read raw request
	read, err := ctx.client.Read(buf)
	if err != nil {
		if err == io.EOF {
			return
		}
		ctx.mod.Error("error while reading from %v: %v", ctx.client.RemoteAddr(), err)
		return
	} else if read == 0 {
		ctx.mod.Error("error while reading from %v: no data", ctx.client.RemoteAddr())
		return
	}

	raw_req := buf[0:read]

	ctx.mod.Debug("read %d bytes from %v:\n%s\n", read, ctx.client.RemoteAddr(), Dump(raw_req))

	// parse as http
	reader := bufio.NewReader(bytes.NewReader(raw_req))
	http_req, err := http.ReadRequest(reader)
	if err != nil {
		ctx.mod.Error("error while parsing http request from %v: %v", ctx.client.RemoteAddr(), err)
		return
	}

	ctx.mod.Info("%v -> %s", ctx.client.RemoteAddr(), tui.Green(http_req.UserAgent()))

	ipp_body := http_req.Body

	// check for an Expect 100-continue
	if http_req.Header.Get("Expect") == "100-continue" {
		// inform the client we're ready to read the request body
		ctx.client.Write([]byte("HTTP/1.1 100 Continue\r\n\r\n"))
		// read the body
		read, err := ctx.client.Read(buf)
		if err != nil {
			if err == io.EOF {
				return
			}
			ctx.mod.Error("error while reading ipp body from %v: %v", ctx.client.RemoteAddr(), err)
			return
		} else if read == 0 {
			ctx.mod.Error("error while reading ipp body from %v: no data", ctx.client.RemoteAddr())
			return
		}

		ipp_body = io.NopCloser(bytes.NewReader(buf[0:read]))
	}

	// parse as IPP
	ipp_req, err := ipp.NewRequestDecoder(ipp_body).Decode(nil)
	if err != nil {
		ctx.mod.Error("error while parsing ip request from %v: %v", ctx.client.RemoteAddr(), err)
		return
	}

	ipp_op_name := fmt.Sprintf("<unknown 0x%x>", ipp_req.Operation)
	if name, found := IPP_REQUEST_NAMES[ipp_req.Operation]; found {
		ipp_op_name = name
	}

	ctx.mod.Info("%v op=%s attributes=%v", ctx.client.RemoteAddr(), tui.Bold(ipp_op_name), ipp_req.OperationAttributes)

	switch ipp_req.Operation {
	// Get-Printer-Attributes
	case 0x000B:
		ippOnGetPrinterAttributes(ctx, ipp_req)

	default:
		ippOnUnhandledRequest(ctx, ipp_req, ipp_op_name)
	}
}

func ippSendResponse(ctx *HandlerContext, response *ipp.Response) {
	ctx.mod.Debug("SENDING %++v", *response)

	resp_data, err := response.Encode()
	if err != nil {
		ctx.mod.Error("error while encoding ipp response: %v", err)
		return
	}

	headers := [][]byte{
		[]byte("HTTP/1.1 200 OK\r\n"),
		[]byte("Content-Type: application/ipp\r\n"),
		[]byte(fmt.Sprintf("Content-Length: %d\r\n", len(resp_data))),
		[]byte("Connection: close\r\n"),
		[]byte("\r\n"),
	}

	for _, header := range headers {
		if _, err := ctx.client.Write(header); err != nil {
			ctx.mod.Error("error while writing header: %v", err)
			return
		}
	}

	if _, err = ctx.client.Write(resp_data); err != nil {
		ctx.mod.Error("error while writing ipp response data: %v", err)
		return
	}

	ctx.mod.Debug("sent %d of ipp response to %v", len(resp_data), ctx.client.RemoteAddr())
}

func ippOnUnhandledRequest(ctx *HandlerContext, ipp_req *ipp.Request, ipp_op_name string) {
	ctx.mod.Warning("unhandled request from %v: operation=%s", ctx.client.RemoteAddr(), ipp_op_name)

	ippSendResponse(ctx, ipp.NewResponse(
		ipp.StatusErrorOperationNotSupported,
		ipp_req.RequestId))
}

func ippOnGetPrinterAttributes(ctx *HandlerContext, ipp_req *ipp.Request) {
	ipp_resp := ipp.NewResponse(ipp.StatusOk, ipp_req.RequestId)

	// https://tools.ietf.org/html/rfc2911 section 3.1.4.2 Response Operation Attributes
	ipp_resp.OperationAttributes["attributes-charset"] = []ipp.Attribute{
		{
			Value: "utf-8",
			Tag:   ipp.TagCharset,
		},
	}
	ipp_resp.OperationAttributes["attributes-natural-language"] = []ipp.Attribute{
		{
			Value: "en",
			Tag:   ipp.TagLanguage,
		},
	}

	// collect user attributes
	userProps := make(map[string]string)
	for name, defaultValue := range IPP_USER_ATTRIBUTES {
		if value, found := ctx.ippAttributes[name]; found {
			userProps[name] = value
		} else {
			userProps[name] = defaultValue
		}
	}

	// rfc2911 section 4.4
	ipp_resp.PrinterAttributes = []ipp.Attributes{
		{
			// custom
			"printer-name": []ipp.Attribute{
				{
					Value: userProps["printer-name"],
					Tag:   ipp.TagName,
				},
			},
			"printer-info": []ipp.Attribute{
				{
					Value: userProps["printer-info"],
					Tag:   ipp.TagText,
				},
			},
			"printer-make-and-model": []ipp.Attribute{
				{
					Value: userProps["printer-make-and-model"],
					Tag:   ipp.TagText,
				},
			},
			"printer-location": []ipp.Attribute{
				{
					Value: userProps["printer-location"],
					Tag:   ipp.TagText,
				},
			},
			"printer-privacy-policy-uri": []ipp.Attribute{
				{
					Value: userProps["printer-privacy-policy-uri"],
					Tag:   ipp.TagUri,
				},
			},
			"ppd-name": []ipp.Attribute{
				{
					Value: userProps["ppd-name"],
					Tag:   ipp.TagName,
				},
			},
			"printer-uri-supported": []ipp.Attribute{
				{
					Value: fmt.Sprintf("%s://%s:%d/printer", ops.Ternary(ctx.srvTLS, "ipps", "ipp"), ctx.srvHost, ctx.srvPort),
					Tag:   ipp.TagUri,
				},
			},
			"uri-security-supported": []ipp.Attribute{
				{
					Value: ops.Ternary(ctx.srvTLS, "tls", "none"),
					Tag:   ipp.TagKeyword,
				},
			},
			"uri-authentication-supported": []ipp.Attribute{
				{
					Value: "none",
					Tag:   ipp.TagKeyword,
				},
			},
			"printer-state": []ipp.Attribute{
				{
					Value: 3, // idle
					Tag:   ipp.TagEnum,
				},
			},
			"printer-state-reasons": []ipp.Attribute{
				{
					Value: "none",
					Tag:   ipp.TagKeyword,
				},
			},
			"ipp-versions-supported": []ipp.Attribute{
				{
					Value: "1.1",
					Tag:   ipp.TagKeyword,
				},
			},
			"operations-supported": []ipp.Attribute{
				{
					Value: []int{
						0x0002, // print job (required by cups)
						0x0004, // validate job (required by cups)
						0x0008, // cancel job (required by cups)
						0x0009, // get job attributes (required by cups)
						0x000b, // get printer attributes
					},
					Tag: ipp.TagEnum,
				},
			},
			"multiple-document-jobs-supported": []ipp.Attribute{
				{
					Value: false,
					Tag:   ipp.TagBoolean,
				},
			},
			"charset-configured": []ipp.Attribute{
				{
					Value: "utf-8",
					Tag:   ipp.TagCharset,
				},
			},
			"charset-supported": []ipp.Attribute{
				{
					Value: "utf-8",
					Tag:   ipp.TagCharset,
				},
			},
			"natural-language-configured": []ipp.Attribute{
				{
					Value: "en",
					Tag:   ipp.TagLanguage,
				},
			},
			"generated-natural-language-supported": []ipp.Attribute{
				{
					Value: "en",
					Tag:   ipp.TagLanguage,
				},
			},
			"document-format-default": []ipp.Attribute{
				{
					Value: "application/pdf",
					Tag:   ipp.TagMimeType,
				},
			},
			"document-format-supported": []ipp.Attribute{
				{
					Value: "application/pdf",
					Tag:   ipp.TagMimeType,
				},
			},
			"printer-is-accepting-jobs": []ipp.Attribute{
				{
					Value: true,
					Tag:   ipp.TagBoolean,
				},
			},
			"queued-job-count": []ipp.Attribute{
				{
					Value: 0,
					Tag:   ipp.TagInteger,
				},
			},
			"pdl-override-supported": []ipp.Attribute{
				{
					Value: "not-attempted",
					Tag:   ipp.TagKeyword,
				},
			},
			"printer-up-time": []ipp.Attribute{
				{
					Value: time.Now().Unix(),
					Tag:   ipp.TagInteger,
				},
			},
			"compression-supported": []ipp.Attribute{
				{
					Value: "none",
					Tag:   ipp.TagKeyword,
				},
			},
		},
	}

	ippSendResponse(ctx, ipp_resp)
}
