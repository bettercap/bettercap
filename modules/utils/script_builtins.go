package utils

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io/ioutil"
	"net/http"

	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/session"

	"github.com/evilsocket/islazy/plugin"

	"github.com/robertkrimen/otto"
)

var nullOtto = otto.Value{}

func errOtto(format string, args ...interface{}) otto.Value {
	log.Error(format, args...)
	return nullOtto
}

func init() {
	// used to read a directory (returns string array)
	plugin.Defines["readDir"] = func(call otto.FunctionCall) otto.Value {
		argv := call.ArgumentList
		argc := len(argv)
		if argc != 1 {
			return errOtto("readDir: expected 1 argument, %d given instead.", argc)
		}

		path := argv[0].String()
		dir, err := ioutil.ReadDir(path)
		if err != nil {
			return errOtto("Could not read directory %s: %s", path, err)
		}

		entry_list := []string{}
		for _, file := range dir {
			entry_list = append(entry_list, file.Name())
		}

		v, err := otto.Otto.ToValue(*call.Otto, entry_list)
		if err != nil {
			return errOtto("Could not convert to array: %s", err)
		}

		return v
	}

	// used to read a file ... doh
	plugin.Defines["readFile"] = func(call otto.FunctionCall) otto.Value {
		argv := call.ArgumentList
		argc := len(argv)
		if argc != 1 {
			return errOtto("readFile: expected 1 argument, %d given instead.", argc)
		}

		filename := argv[0].String()
		raw, err := ioutil.ReadFile(filename)
		if err != nil {
			return errOtto("Could not read file %s: %s", filename, err)
		}

		v, err := otto.ToValue(string(raw))
		if err != nil {
			return errOtto("Could not convert to string: %s", err)
		}
		return v
	}

	plugin.Defines["writeFile"] = func(call otto.FunctionCall) otto.Value {
		argv := call.ArgumentList
		argc := len(argv)
		if argc != 2 {
			return errOtto("writeFile: expected 2 arguments, %d given instead.", argc)
		}

		filename := argv[0].String()
		data := argv[1].String()

		err := ioutil.WriteFile(filename, []byte(data), 0644)
		if err != nil {
			return errOtto("Could not write %d bytes to %s: %s", len(data), filename, err)
		}

		return otto.NullValue()
	}

	// log something
	plugin.Defines["log"] = func(call otto.FunctionCall) otto.Value {
		for _, v := range call.ArgumentList {
			log.Info("%s", v.String())
		}
		return otto.Value{}
	}

	// log debug
	plugin.Defines["log_debug"] = func(call otto.FunctionCall) otto.Value {
		for _, v := range call.ArgumentList {
			log.Debug("%s", v.String())
		}
		return otto.Value{}
	}

	// log info
	plugin.Defines["log_info"] = func(call otto.FunctionCall) otto.Value {
		for _, v := range call.ArgumentList {
			log.Info("%s", v.String())
		}
		return otto.Value{}
	}

	// log warning
	plugin.Defines["log_warn"] = func(call otto.FunctionCall) otto.Value {
		for _, v := range call.ArgumentList {
			log.Warning("%s", v.String())
		}
		return otto.Value{}
	}

	// log error
	plugin.Defines["log_error"] = func(call otto.FunctionCall) otto.Value {
		for _, v := range call.ArgumentList {
			log.Error("%s", v.String())
		}
		return otto.Value{}
	}

	// log fatal
	plugin.Defines["log_fatal"] = func(call otto.FunctionCall) otto.Value {
		for _, v := range call.ArgumentList {
			log.Fatal("%s", v.String())
		}
		return otto.Value{}
	}

	// javascript btoa function
	plugin.Defines["btoa"] = func(call otto.FunctionCall) otto.Value {
		varValue := base64.StdEncoding.EncodeToString([]byte(call.Argument(0).String()))
		v, err := otto.ToValue(varValue)
		if err != nil {
			return errOtto("Could not convert to string: %s", varValue)
		}
		return v
	}

	// javascript atob function
	plugin.Defines["atob"] = func(call otto.FunctionCall) otto.Value {
		varValue, err := base64.StdEncoding.DecodeString(call.Argument(0).String())
		if err != nil {
			return errOtto("Could not decode string: %s", call.Argument(0).String())
		}
		v, err := otto.ToValue(string(varValue))
		if err != nil {
			return errOtto("Could not convert to string: %s", varValue)
		}
		return v
	}

	// compress data with gzip
	plugin.Defines["gzipCompress"] = func(call otto.FunctionCall) otto.Value {
		argv := call.ArgumentList
		argc := len(argv)
		if argc != 1 {
			return errOtto("gzipCompress: expected 1 argument, %d given instead.", argc)
		}

		uncompressed_bytes := []byte(argv[0].String())

		var writer_buffer bytes.Buffer
		gzip_writer := gzip.NewWriter(&writer_buffer)
		_, err := gzip_writer.Write(uncompressed_bytes)
		if err != nil {
			return errOtto("gzipCompress: could not compress data: %s", err.Error())
		}
		gzip_writer.Close()

		compressed_bytes := writer_buffer.Bytes()

		v, err := otto.ToValue(string(compressed_bytes))
		if err != nil {
			return errOtto("Could not convert to string: %s", err.Error())
		}

		return v
	}

	// decompress data with gzip
	plugin.Defines["gzipDecompress"] = func(call otto.FunctionCall) otto.Value {
		argv := call.ArgumentList
		argc := len(argv)
		if argc != 1 {
			return errOtto("gzipDecompress: expected 1 argument, %d given instead.", argc)
		}

		compressed_bytes := []byte(argv[0].String())
		reader_buffer := bytes.NewBuffer(compressed_bytes)

		gzip_reader, err := gzip.NewReader(reader_buffer)
		if err != nil {
			return errOtto("gzipDecompress: could not create gzip reader: %s", err.Error())
		}

		var decompressed_buffer bytes.Buffer
		_, err = decompressed_buffer.ReadFrom(gzip_reader)
		if err != nil {
			return errOtto("gzipDecompress: could not decompress data: %s", err.Error())
		}

		decompressed_bytes := decompressed_buffer.Bytes()
		v, err := otto.ToValue(string(decompressed_bytes))
		if err != nil {
			return errOtto("Could not convert to string: %s", err.Error())
		}

		return v
	}

	// read or write environment variable
	plugin.Defines["env"] = func(call otto.FunctionCall) otto.Value {
		argv := call.ArgumentList
		argc := len(argv)

		if argc == 1 {
			// get
			varName := call.Argument(0).String()
			if found, varValue := session.I.Env.Get(varName); found {
				v, err := otto.ToValue(varValue)
				if err != nil {
					return errOtto("Could not convert to string: %s", varValue)
				}
				return v
			}

		} else if argc == 2 {
			// set
			varName := call.Argument(0).String()
			varValue := call.Argument(1).String()
			session.I.Env.Set(varName, varValue)
		} else {
			return errOtto("env: expected 1 or 2 arguments, %d given instead.", argc)
		}

		return nullOtto
	}

	// send http request
	plugin.Defines["httpRequest"] = func(call otto.FunctionCall) otto.Value {
		argv := call.ArgumentList
		argc := len(argv)
		if argc < 2 {
			return errOtto("httpRequest: expected 2 or more, %d given instead.", argc)
		}

		method := argv[0].String()
		url := argv[1].String()

		client := &http.Client{}
		req, err := http.NewRequest(method, url, nil)
		if argc >= 3 {
			data := argv[2].String()
			req, err = http.NewRequest(method, url, bytes.NewBuffer([]byte(data)))
			if err != nil {
				return errOtto("Could create request to url %s: %s", url, err)
			}

			if argc > 3 {
				headers := argv[3].Object()
				for _, key := range headers.Keys() {
					v, err := headers.Get(key)
					if err != nil {
						return errOtto("Could add header %s to request: %s", key, err)
					}
					req.Header.Add(key, v.String())
				}
			}
		} else if err != nil {
			return errOtto("Could create request to url %s: %s", url, err)
		}

		resp, err := client.Do(req)
		if err != nil {
			return errOtto("Could not request url %s: %s", url, err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errOtto("Could not read response: %s", err)
		}

		object, err := otto.New().Object("({})")
		if err != nil {
			return errOtto("Could not create response object: %s", err)
		}

		err = object.Set("body", string(body))
		if err != nil {
			return errOtto("Could not populate response object: %s", err)
		}

		v, err := otto.ToValue(object)
		if err != nil {
			return errOtto("Could not convert to object: %s", err)
		}
		return v
	}
}
