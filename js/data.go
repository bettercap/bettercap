package js

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"

	"github.com/robertkrimen/otto"
)

func textEncode(call otto.FunctionCall) otto.Value {
	argv := call.ArgumentList
	argc := len(argv)
	if argc != 1 {
		return ReportError("textEncode: expected 1 argument, %d given instead.", argc)
	}

	arg := argv[0]
	if (!arg.IsString()) {
		return ReportError("textEncode: single argument must be a string.")
	}

	encoded := []byte(arg.String())
	vm := otto.New()
	v, err := vm.ToValue(encoded)
	if err != nil {
		return ReportError("textEncode: could not convert to []uint8: %s", err.Error())
	}

	return v
}

func textDecode(call otto.FunctionCall) otto.Value {
	argv := call.ArgumentList
	argc := len(argv)
	if argc != 1 {
		return ReportError("textDecode: expected 1 argument, %d given instead.", argc)
	}

	arg, err := argv[0].Export()
	if err != nil {
		return ReportError("textDecode: could not export argument value: %s", err.Error())
	}
	byteArr, ok := arg.([]uint8)
	if !ok {
		return ReportError("textDecode: single argument must be of type []uint8.")
	}

	decoded := string(byteArr)
	v, err := otto.ToValue(decoded)
	if err != nil {
		return ReportError("textDecode: could not convert to string: %s", err.Error())
	}

	return v
}

func btoa(call otto.FunctionCall) otto.Value {
	argv := call.ArgumentList
	argc := len(argv)
	if argc != 1 {
		return ReportError("btoa: expected 1 argument, %d given instead.", argc)
	}

	arg := argv[0]
	if (!arg.IsString()) {
		return ReportError("btoa: single argument must be a string.")
	}

	encoded := base64.StdEncoding.EncodeToString([]byte(arg.String()))
	v, err := otto.ToValue(encoded)
	if err != nil {
		return ReportError("btoa: could not convert to string: %s", err.Error())
	}

	return v
}

func atob(call otto.FunctionCall) otto.Value {
	argv := call.ArgumentList
	argc := len(argv)
	if argc != 1 {
		return ReportError("atob: expected 1 argument, %d given instead.", argc)
	}

	arg := argv[0]
	if (!arg.IsString()) {
		return ReportError("atob: single argument must be a string.")
	}

	decoded, err := base64.StdEncoding.DecodeString(arg.String())
	if err != nil {
		return ReportError("atob: could not decode string: %s", err.Error())
	}

	v, err := otto.ToValue(string(decoded))
	if err != nil {
		return ReportError("atob: could not convert to string: %s", err.Error())
	}

	return v
}

func gzipCompress(call otto.FunctionCall) otto.Value {
	argv := call.ArgumentList
	argc := len(argv)
	if argc != 1 {
		return ReportError("gzipCompress: expected 1 argument, %d given instead.", argc)
	}

	arg := argv[0]
	if (!arg.IsString()) {
		return ReportError("gzipCompress: single argument must be a string.")
	}

	uncompressedBytes := []byte(arg.String())

	var writerBuffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&writerBuffer)
	_, err := gzipWriter.Write(uncompressedBytes)
	if err != nil {
		return ReportError("gzipCompress: could not compress data: %s", err.Error())
	}
	gzipWriter.Close()

	compressedBytes := writerBuffer.Bytes()

	v, err := otto.ToValue(string(compressedBytes))
	if err != nil {
		return ReportError("gzipCompress: could not convert to string: %s", err.Error())
	}

	return v
}

func gzipDecompress(call otto.FunctionCall) otto.Value {
	argv := call.ArgumentList
	argc := len(argv)
	if argc != 1 {
		return ReportError("gzipDecompress: expected 1 argument, %d given instead.", argc)
	}

	compressedBytes := []byte(argv[0].String())
	readerBuffer := bytes.NewBuffer(compressedBytes)

	gzipReader, err := gzip.NewReader(readerBuffer)
	if err != nil {
		return ReportError("gzipDecompress: could not create gzip reader: %s", err.Error())
	}

	var decompressedBuffer bytes.Buffer
	_, err = decompressedBuffer.ReadFrom(gzipReader)
	if err != nil {
		return ReportError("gzipDecompress: could not decompress data: %s", err.Error())
	}

	decompressedBytes := decompressedBuffer.Bytes()
	v, err := otto.ToValue(string(decompressedBytes))
	if err != nil {
		return ReportError("gzipDecompress: could not convert to string: %s", err.Error())
	}

	return v
}
