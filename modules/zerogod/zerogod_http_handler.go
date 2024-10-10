package zerogod

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/evilsocket/islazy/tui"
)

func httpGenericHandler(ctx *HandlerContext) (shouldQuit bool, clientIP string, req *http.Request, err error) {
	clientIP = strings.SplitN(ctx.client.RemoteAddr().String(), ":", 2)[0]

	buf := make([]byte, 4096)
	// read raw request
	read, err := ctx.client.Read(buf)
	if err != nil {
		if err == io.EOF {
			ctx.mod.Debug("EOF, client %s disconnected", clientIP)
			return true, clientIP, nil, nil
		}
		ctx.mod.Warning("error while reading from %v: %v", clientIP, err)
		return true, clientIP, nil, err
	} else if read == 0 {
		ctx.mod.Warning("error while reading from %v: no data", clientIP)
		return true, clientIP, nil, err
	}

	raw_req := buf[0:read]

	ctx.mod.Debug("read %d bytes from %v:\n%s\n", read, clientIP, Dump(raw_req))

	// parse as http
	reader := bufio.NewReader(bytes.NewReader(raw_req))
	httpRequest, err := http.ReadRequest(reader)
	if err != nil {
		ctx.mod.Error("error while parsing http request from %v: %v\n%s", clientIP, err, Dump(raw_req))
		return true, clientIP, nil, err
	}

	return false, clientIP, httpRequest, nil
}

func httpLogRequest(ctx *HandlerContext, clientIP string, httpRequest *http.Request) {
	clientUA := httpRequest.UserAgent()
	ctx.mod.Info("%v (%s) > %s %s",
		clientIP,
		tui.Green(clientUA),
		tui.Bold(httpRequest.Method),
		tui.Yellow(httpRequest.RequestURI))
}

func httpClientHandler(ctx *HandlerContext) {
	defer ctx.client.Close()

	shouldQuit, clientIP, httpRequest, err := httpGenericHandler(ctx)
	if shouldQuit {
		return
	} else if err != nil {
		ctx.mod.Error("%v", err)
	}

	httpLogRequest(ctx, clientIP, httpRequest)

	respStatusCode := 404
	respStatus := "Not Found"
	respBody := `<html>
<head><title>Not Found</title></head>
<body>
<center><h1>Not Found</h1></center>
</body>
</html>`

	// see if anything in config matches
	for path, body := range ctx.httpPaths {
		if httpRequest.RequestURI == path {
			respStatusCode = 200
			respStatus = "OK"
			respBody = body
			break
		}
	}

	response := fmt.Sprintf(`HTTP/1.1 %d %s
Content-Type: text/html; charset=utf-8
Content-Length: %d
Connection: close

%s`,
		respStatusCode,
		respStatus,
		len(respBody),
		respBody,
	)

	if _, err = ctx.client.Write([]byte(response)); err != nil {
		ctx.mod.Error("error while writing http response data: %v", err)
	} else {
		ctx.mod.Debug("sent %d of http response to %v", len(response), ctx.client.RemoteAddr())
	}
}
