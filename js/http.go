package js

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/robertkrimen/otto"
)

type httpPackage struct{}

type httpResponse struct {
	Error    error
	Response *http.Response
	Raw      []byte
	Body     string
}

// Encode encodes a string for use in a URL query.
func (c httpPackage) Encode(s string) string {
	return url.QueryEscape(s)
}

// Request sends an HTTP request with the specified method, URL, headers, form data, or JSON payload.
func (c httpPackage) Request(method, uri string, headers map[string]string, form map[string]string, json string) httpResponse {
	var reader io.Reader

	if form != nil {
		data := url.Values{}
		for k, v := range form {
			data.Set(k, v)
		}
		reader = strings.NewReader(data.Encode())
	} else if json != "" {
		reader = strings.NewReader(json)
	}

	req, err := http.NewRequest(method, uri, reader)
	if err != nil {
		return httpResponse{Error: fmt.Errorf("failed to create request: %w", err)}
	}

	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else if json != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	for name, value := range headers {
		req.Header.Set(name, value)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return httpResponse{Error: fmt.Errorf("request failed: %w", err)}
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return httpResponse{Error: fmt.Errorf("failed to read response body: %w", err)}
	}

	res := httpResponse{
		Response: resp,
		Raw:      raw,
		Body:     string(raw),
	}

	if resp.StatusCode >= 400 {
		res.Error = fmt.Errorf("HTTP error: %s", resp.Status)
	}

	return res
}

// Get sends a GET request.
func (c httpPackage) Get(url string, headers map[string]string) httpResponse {
	return c.Request(http.MethodGet, url, headers, nil, "")
}

// PostForm sends a POST request with form data.
func (c httpPackage) PostForm(url string, headers map[string]string, form map[string]string) httpResponse {
	return c.Request(http.MethodPost, url, headers, form, "")
}

// PostJSON sends a POST request with a JSON payload.
func (c httpPackage) PostJSON(url string, headers map[string]string, json string) httpResponse {
	return c.Request(http.MethodPost, url, headers, nil, json)
}

// httpRequest processes JavaScript calls for HTTP requests.
func httpRequest(call otto.FunctionCall) otto.Value {
	if len(call.ArgumentList) < 2 {
		return ReportError("httpRequest: expected at least 2 arguments, got %d", len(call.ArgumentList))
	}

	method := call.Argument(0).String()
	url := call.Argument(1).String()

	var reader io.Reader
	if len(call.ArgumentList) >= 3 {
		data := call.Argument(2).String()
		reader = bytes.NewBufferString(data)
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return ReportError("failed to create request for URL %s: %s", url, err)
	}

	if len(call.ArgumentList) > 3 {
		headers, _ := call.Argument(3).Export()
		if headerMap, ok := headers.(map[string]interface{}); ok {
			for key, value := range headerMap {
				if strValue, ok := value.(string); ok {
					req.Header.Set(key, strValue)
				}
			}
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return ReportError("failed to execute request to URL %s: %s", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ReportError("failed to read response body: %s", err)
	}

	responseObj, _ := otto.New().Object(`({})`)
	responseObj.Set("body", string(body))

	v, err := otto.ToValue(responseObj)
	if err != nil {
		return ReportError("failed to convert response to Otto value: %s", err)
	}
	return v
}

// ReportError formats and returns a JavaScript-compatible error.
func ReportError(format string, args ...interface{}) otto.Value {
	errMessage := fmt.Sprintf(format, args...)
	fmt.Println("Error:", errMessage) // Log the error
	val, _ := otto.ToValue(fmt.Errorf(errMessage))
	return val
}
