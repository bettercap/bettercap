package zerogod

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
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

	clientIP := strings.SplitN(ctx.client.RemoteAddr().String(), ":", 2)[0]

	buf := make([]byte, 4096)

	// read raw request
	read, err := ctx.client.Read(buf)
	if err != nil {
		if err == io.EOF {
			return
		}
		ctx.mod.Warning("error while reading from %v: %v", clientIP, err)
		return
	} else if read == 0 {
		ctx.mod.Warning("error while reading from %v: no data", clientIP)
		return
	}

	raw_req := buf[0:read]

	ctx.mod.Debug("read %d bytes from %v:\n%s\n", read, clientIP, Dump(raw_req))

	// parse as http
	reader := bufio.NewReader(bytes.NewReader(raw_req))
	http_req, err := http.ReadRequest(reader)
	if err != nil {
		ctx.mod.Error("error while parsing http request from %v: %v", clientIP, err)
		return
	}

	clientUA := http_req.UserAgent()
	ctx.mod.Debug("%v -> %s", clientIP, tui.Green(clientUA))

	ipp_body, err := ippReadRequestBody(ctx, http_req)
	if err != nil {
		ctx.mod.Error("%v", err)
		return
	}

	// parse as IPP
	ipp_req, err := ipp.NewRequestDecoder(ipp_body).Decode(nil)
	if err != nil {
		ctx.mod.Error("error while parsing ipp request from %v: %v -> %++v", clientIP, err, *http_req)
		return
	}

	ipp_op_name := fmt.Sprintf("<unknown 0x%x>", ipp_req.Operation)
	if name, found := IPP_REQUEST_NAMES[ipp_req.Operation]; found {
		ipp_op_name = name
	}

	ctx.mod.Info("%s <- %s (%s) %s",
		tui.Yellow(ctx.service),
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
		ippOnPrintJob(ctx, http_req, ipp_req)
	// Get-Job-Attributes
	case 0x0009:
		ippOnGetJobAttributes(ctx, ipp_req)

	default:
		ippOnUnhandledRequest(ctx, ipp_req, ipp_op_name)
	}
}
