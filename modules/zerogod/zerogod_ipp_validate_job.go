package zerogod

import (
	"github.com/evilsocket/islazy/tui"
	"github.com/phin1x/go-ipp"
)

func ippOnValidateJob(ctx *HandlerContext, ipp_req *ipp.Request) {
	jobName := "<unknown>"
	jobUUID := "<unknown>"
	jobUser := "<unknown>"

	if value, found := ipp_req.OperationAttributes["job-name"]; found {
		jobName = value.(string)
	}

	if value, found := ipp_req.OperationAttributes["requesting-user-name"]; found {
		jobUser = value.(string)
	}

	if value, found := ipp_req.JobAttributes["job-uuid"]; found {
		jobUUID = value.(string)
	}

	ctx.mod.Debug("validating job_name=%s job_uuid=%s job_user=%s", tui.Yellow(jobName), tui.Dim(jobUUID), tui.Green(jobUser))

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
