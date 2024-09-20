package zerogod

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/evilsocket/islazy/fs"
	"github.com/evilsocket/islazy/ops"
	"github.com/evilsocket/islazy/str"
	"github.com/evilsocket/islazy/tui"

	"github.com/phin1x/go-ipp"
)

const IPP_CHUNK_MAX_LINE_SIZE = 1024

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
	ipp.AttributeTagMapping["job-state-reasons"] = ipp.TagKeyword
	ipp.AttributeTagMapping["job-state"] = ipp.TagEnum
	ipp.AttributeTagMapping["job-uri"] = ipp.TagUri
	ipp.AttributeTagMapping["job-id"] = ipp.TagInteger
	ipp.AttributeTagMapping["job-printer-uri"] = ipp.TagUri
	ipp.AttributeTagMapping["job-name"] = ipp.TagName
	ipp.AttributeTagMapping["job-originating-user-name"] = ipp.TagName
	ipp.AttributeTagMapping["time-at-creation"] = ipp.TagInteger
	ipp.AttributeTagMapping["time-at-completed"] = ipp.TagInteger
	ipp.AttributeTagMapping["job-printer-up-time"] = ipp.TagInteger
}

func ippReadChunkSizeHex(ctx *HandlerContext) string {
	var buf []byte

	for b := make([]byte, 1); ; {
		if n, err := ctx.client.Read(b); err != nil {
			ctx.mod.Error("could not read chunked byte: %v", err)
		} else if n == 0 {
			break
		} else if b[0] == '\n' {
			break
		} else {
			// ctx.mod.Info("buf += 0x%x (%c)", b[0], b[0])
			buf = append(buf, b[0])
		}

		if len(buf) >= IPP_CHUNK_MAX_LINE_SIZE {
			ctx.mod.Warning("buffer size exceeded %d bytes when reading chunk size", IPP_CHUNK_MAX_LINE_SIZE)
			break
		}
	}

	return str.Trim(string(buf))
}

func ippReadChunkSize(ctx *HandlerContext) (uint64, error) {
	if chunkSizeHex := ippReadChunkSizeHex(ctx); chunkSizeHex != "" {
		ctx.mod.Debug("got chunk size: 0x%s", chunkSizeHex)
		return strconv.ParseUint(chunkSizeHex, 16, 64)
	}
	return 0, nil
}

func ippReadChunkedBody(ctx *HandlerContext) ([]byte, error) {
	var chunkedBody []byte
	// read chunked loop
	for {
		// read the next chunk size
		if chunkSize, err := ippReadChunkSize(ctx); err != nil {
			return nil, fmt.Errorf("error reading next chunk size: %v", err)
		} else if chunkSize == 0 {
			break
		} else {
			chunk := make([]byte, chunkSize)
			if n, err := ctx.client.Read(chunk); err != nil {
				return nil, fmt.Errorf("error while reading chunk of %d bytes: %v", chunkSize, err)
			} else if n != int(chunkSize) {
				return nil, fmt.Errorf("expected chunk of size %d, got %d bytes", chunkSize, n)
			} else {
				chunkedBody = append(chunkedBody, chunk...)
			}
		}
	}

	return chunkedBody, nil
}

func ippReadRequestBody(ctx *HandlerContext, http_req *http.Request) (io.ReadCloser, error) {
	ipp_body := http_req.Body

	// check for an Expect 100-continue
	if http_req.Header.Get("Expect") == "100-continue" {
		buf := make([]byte, 4096)

		// inform the client we're ready to read the request body
		ctx.client.Write([]byte("HTTP/1.1 100 Continue\r\n\r\n"))

		if slices.Contains(http_req.TransferEncoding, "chunked") {
			ctx.mod.Debug("detected chunked encoding")
			if body, err := ippReadChunkedBody(ctx); err != nil {
				return nil, err
			} else {
				ipp_body = io.NopCloser(bytes.NewReader(body))
			}
		} else {
			// read the body in a single step
			read, err := ctx.client.Read(buf)
			if err != nil {
				if err == io.EOF {
					return nil, nil
				}
				return nil, fmt.Errorf("error while reading ipp body from %v: %v", ctx.client.RemoteAddr(), err)
			} else if read == 0 {
				return nil, fmt.Errorf("error while reading ipp body from %v: no data", ctx.client.RemoteAddr())
			}

			ipp_body = io.NopCloser(bytes.NewReader(buf[0:read]))
		}
	}

	return ipp_body, nil
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
	ctx.mod.Warning("unhandled request from %v: operation=%s - %++v", ctx.client.RemoteAddr(), ipp_op_name, *ipp_req)

	ippSendResponse(ctx, ipp.NewResponse(
		ipp.StatusErrorOperationNotSupported,
		ipp_req.RequestId))
}

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
			Value: "bettercap", // TODO: check if this must match the actual job user from a print operation
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
