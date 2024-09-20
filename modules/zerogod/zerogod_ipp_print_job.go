package zerogod

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/evilsocket/islazy/fs"
	"github.com/evilsocket/islazy/ops"
	"github.com/evilsocket/islazy/tui"
	"github.com/phin1x/go-ipp"
)

func ippOnPrintJob(ctx *HandlerContext, http_req *http.Request, ipp_req *ipp.Request) {
	var err error

	createdAt := time.Now()

	data := PrintData{
		CreatedAt: createdAt,
		Service:   ctx.service,
		Client: ClientData{
			UA: http_req.UserAgent(),
			IP: strings.SplitN(ctx.client.RemoteAddr().String(), ":", 2)[0],
		},
		Job:      JobData{},
		Document: DocumentData{},
	}

	if value, found := ipp_req.OperationAttributes["job-name"]; found {
		data.Job.Name = value.(string)
	}
	if value, found := ipp_req.OperationAttributes["requesting-user-name"]; found {
		data.Job.User = value.(string)
	}
	if value, found := ipp_req.JobAttributes["job-uuid"]; found {
		data.Job.UUID = value.(string)
	}
	if value, found := ipp_req.JobAttributes["document-name-supplied"]; found {
		data.Document.Name = value.(string)
	}
	if value, found := ipp_req.OperationAttributes["document-format"]; found {
		data.Document.Format = value.(string)
	}

	// TODO: check if not chunked
	data.Document.Data, err = ippReadChunkedBody(ctx)
	if err != nil {
		ctx.mod.Error("could not read document body: %v", err)
	}

	var docPath string
	if err, docPath = ctx.mod.StringParam("zerogod.ipp.save_path"); err != nil {
		ctx.mod.Error("can't read parameter zerogod.ipp.save_path: %v", err)
	} else if docPath, err = fs.Expand(docPath); err != nil {
		ctx.mod.Error("can't expand %s: %v", docPath, err)
	} else {
		// make sure the path exists
		if err := os.MkdirAll(docPath, 0755); err != nil {
			ctx.mod.Error("could not create directory %s: %v", docPath, err)
		}

		docName := path.Join(docPath, fmt.Sprintf("%d.json", createdAt.UnixMicro()))
		ctx.mod.Debug("saving to %s: %++v", docName, data)
		jsonData, err := json.Marshal(data)
		if err != nil {
			ctx.mod.Error("could not marshal data to json: %v", err)
		} else if err := ioutil.WriteFile(docName, jsonData, 0644); err != nil {
			ctx.mod.Error("could not write data to %s: %v", docName, err)
		} else {
			ctx.mod.Info("  document saved to %s", tui.Yellow(docName))
		}
	}

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
			Value: 3, // 3=pending https://tools.ietf.org/html/rfc2911#section-4.3.7
			Tag:   ipp.TagEnum,
		},
	}
	ipp_resp.OperationAttributes["job-state-reasons"] = []ipp.Attribute{
		{
			Value: []string{
				"job-incoming",
				"job-data-insufficient",
			},
			Tag: ipp.TagKeyword,
		},
	}

	ippSendResponse(ctx, ipp_resp)
}
