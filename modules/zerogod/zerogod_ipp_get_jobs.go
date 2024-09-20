package zerogod

import (
	"github.com/evilsocket/islazy/tui"
	"github.com/phin1x/go-ipp"
)

func ippOnGetJobs(ctx *HandlerContext, ipp_req *ipp.Request) {
	jobUser := "<unknown>"
	if value, found := ipp_req.OperationAttributes["requesting-user-name"]; found {
		jobUser = value.(string)
	}

	ctx.mod.Debug("responding with empty jobs list to requesting_user=%s", tui.Green(jobUser))

	// respond with an empty list of jobs, which probably breaks the rfc
	// if the client asked for completed jobs https://tools.ietf.org/html/rfc2911#section-3.2.6.2
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

	ippSendResponse(ctx, ipp_resp)
}
