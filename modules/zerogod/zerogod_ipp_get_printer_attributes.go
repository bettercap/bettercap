package zerogod

import (
	"fmt"
	"time"

	"github.com/evilsocket/islazy/ops"
	"github.com/phin1x/go-ipp"
)

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
