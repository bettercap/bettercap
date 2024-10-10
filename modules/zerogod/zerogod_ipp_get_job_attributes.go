package zerogod

import (
	"fmt"
	"time"

	"github.com/evilsocket/islazy/ops"
	"github.com/phin1x/go-ipp"
)

func ippOnGetJobAttributes(ctx *HandlerContext, ipp_req *ipp.Request) {
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

	jobID := 666

	ipp_resp.OperationAttributes["job-uri"] = []ipp.Attribute{
		{
			Value: fmt.Sprintf("%s://%s:%d/jobs/%d", ops.Ternary(ctx.srvTLS, "ipps", "ipp"), ctx.srvHost, ctx.srvPort, jobID),
			Tag:   ipp.TagUri,
		},
	}
	ipp_resp.OperationAttributes["job-id"] = []ipp.Attribute{
		{
			Value: jobID,
			Tag:   ipp.TagInteger,
		},
	}
	ipp_resp.OperationAttributes["job-state"] = []ipp.Attribute{
		{
			Value: 9, // 9=completed https://tools.ietf.org/html/rfc2911#section-4.3.7
			Tag:   ipp.TagEnum,
		},
	}
	ipp_resp.OperationAttributes["job-state-reasons"] = []ipp.Attribute{
		{
			Value: []string{
				"job-completed-successfully",
			},
			Tag: ipp.TagKeyword,
		},
	}
	ipp_resp.OperationAttributes["job-printer-uri"] = []ipp.Attribute{
		{
			Value: fmt.Sprintf("%s://%s:%d/printer", ops.Ternary(ctx.srvTLS, "ipps", "ipp"), ctx.srvHost, ctx.srvPort),
			Tag:   ipp.TagUri,
		},
	}
	ipp_resp.OperationAttributes["job-name"] = []ipp.Attribute{
		{
			Value: "Print job 666",
			Tag:   ipp.TagName,
		},
	}
	ipp_resp.OperationAttributes["job-originating-user-name"] = []ipp.Attribute{
		{
			Value: "bettercap",
			Tag:   ipp.TagName,
		},
	}
	ipp_resp.OperationAttributes["time-at-creation"] = []ipp.Attribute{
		{
			Value: 0,
			Tag:   ipp.TagInteger,
		},
	}
	ipp_resp.OperationAttributes["time-at-completed"] = []ipp.Attribute{
		{
			Value: 0,
			Tag:   ipp.TagInteger,
		},
	}
	ipp_resp.OperationAttributes["job-printer-up-time"] = []ipp.Attribute{
		{
			Value: time.Now().Unix(),
			Tag:   ipp.TagInteger,
		},
	}

	ippSendResponse(ctx, ipp_resp)
}
