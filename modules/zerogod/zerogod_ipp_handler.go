package zerogod

import (
	"fmt"
	"time"

	"github.com/evilsocket/islazy/tui"

	"github.com/phin1x/go-ipp"
)

type ClientData struct {
	IP string `json:"ip"`
	UA string `json:"user_agent"`
}

type JobData struct {
	Name string `json:"name"`
	UUID string `json:"uuid"`
	User string `json:"username"`
}

type DocumentData struct {
	Name   string `json:"name"`
	Format string `json:"format"`
	Data   []byte `json:"data"`
}

type PrintData struct {
	CreatedAt time.Time    `json:"created_at"`
	Service   string       `json:"service"`
	Client    ClientData   `json:"client"`
	Job       JobData      `json:"job"`
	Document  DocumentData `json:"document"`
}

func ippClientHandler(ctx *HandlerContext) {
	defer ctx.client.Close()

	shouldQuit, clientIP, httpRequest, err := httpGenericHandler(ctx)
	if shouldQuit {
		return
	} else if err != nil {
		ctx.mod.Error("%v", err)
	}

	clientUA := httpRequest.UserAgent()
	ctx.mod.Info("%v -> %s", clientIP, tui.Green(clientUA))

	ipp_body, err := ippReadRequestBody(ctx, httpRequest)
	if err != nil {
		ctx.mod.Error("%v", err)
		return
	} else if ipp_body == nil {
		ctx.mod.Warning("no ipp request body from %v (%s)", clientIP, clientUA)
		return
	}

	// parse as IPP
	ipp_req, err := ipp.NewRequestDecoder(ipp_body).Decode(nil)
	if err != nil {
		ctx.mod.Error("error while parsing ipp request from %v: %v -> %++v", clientIP, err, *httpRequest)
		return
	}

	ipp_op_name := fmt.Sprintf("<unknown 0x%x>", ipp_req.Operation)
	if name, found := IPP_REQUEST_NAMES[ipp_req.Operation]; found {
		ipp_op_name = name
	}

	reqUsername := tui.Dim("<unknown>")
	if value, found := ipp_req.OperationAttributes["requesting-user-name"]; found {
		reqUsername = tui.Blue(value.(string))
	}

	ctx.mod.Info("%s <- %s@%s (%s) %s",
		tui.Yellow(ctx.service),
		reqUsername,
		clientIP,
		tui.Green(clientUA),
		tui.Bold(ipp_op_name))
	ctx.mod.Debug("  %++v", *ipp_req)

	switch ipp_req.Operation {
	// Get-Printer-Attributes
	case 0x000B:
		ippOnGetPrinterAttributes(ctx, ipp_req)
	// Validate-Job
	case 0x0004:
		ippOnValidateJob(ctx, ipp_req)
	// Get-Jobs
	case 0x000A:
		ippOnGetJobs(ctx, ipp_req)
	// Print-Job
	case 0x0002:
		ippOnPrintJob(ctx, httpRequest, ipp_req)
	// Get-Job-Attributes
	case 0x0009:
		ippOnGetJobAttributes(ctx, ipp_req)

	default:
		ippOnUnhandledRequest(ctx, ipp_req, ipp_op_name)
	}
}
