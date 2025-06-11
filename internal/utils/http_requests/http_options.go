package http_requests

import "io"

type HttpOptions struct {
	Type  string
	Value interface{}
}

const (
	HttpOptionTypeWriteTimeout                     = "write_timeout"
	HttpOptionTypeReadTimeout                      = "read_timeout"
	HttpOptionTypeHeader                           = "header"
	HttpOptionTypeParams                           = "params"
	HttpOptionTypePayload                          = "payload"
	HttpOptionTypePayloadText                      = "payloadText"
	HttpOptionTypePayloadJson                      = "payloadJson"
	HttpOptionTypePayloadMultipart                 = "payloadMultipart"
	HttpOptionTypeRaiseErrorWhenStreamDataNotMatch = "raiseErrorWhenStreamDataNotMatch"
	HttpOptionTypeDirectReferer                    = "directReferer"
	HttpOptionTypeRetCode                          = "retCode"
	HttpOptionTypeUsingLengthPrefixed              = "usingLengthPrefixed"
)

// milliseconds
func HttpWriteTimeout(timeout int64) HttpOptions {
	return HttpOptions{HttpOptionTypeWriteTimeout, timeout}
}

// milliseconds
func HttpReadTimeout(timeout int64) HttpOptions {
	return HttpOptions{HttpOptionTypeReadTimeout, timeout}
}

func HttpHeader(header map[string]string) HttpOptions {
	return HttpOptions{HttpOptionTypeHeader, header}
}

// which is used for params with in url
func HttpParams(params map[string]string) HttpOptions {
	return HttpOptions{HttpOptionTypeParams, params}
}

// which is used for POST method only
func HttpPayload(payload map[string]string) HttpOptions {
	return HttpOptions{HttpOptionTypePayload, payload}
}

// which is used for POST method only
func HttpPayloadText(payload string) HttpOptions {
	return HttpOptions{HttpOptionTypePayloadText, payload}
}

// which is used for POST method only
func HttpPayloadReader(reader io.ReadCloser) HttpOptions {
	return HttpOptions{"payloadReader", reader}
}

// which is used for POST method only
func HttpPayloadJson(payload interface{}) HttpOptions {
	return HttpOptions{HttpOptionTypePayloadJson, payload}
}

type HttpPayloadMultipartFile struct {
	Filename string
	Reader   io.Reader
}

// which is used for POST method only
// payload follows the form data format, and files is a map from filename to file
func HttpPayloadMultipart(payload map[string]string, files map[string]HttpPayloadMultipartFile) HttpOptions {
	return HttpOptions{HttpOptionTypePayloadMultipart, map[string]interface{}{
		"payload": payload,
		"files":   files,
	}}
}

func HttpRaiseErrorWhenStreamDataNotMatch(raise bool) HttpOptions {
	return HttpOptions{HttpOptionTypeRaiseErrorWhenStreamDataNotMatch, raise}
}

func HttpWithDirectReferer() HttpOptions {
	return HttpOptions{HttpOptionTypeDirectReferer, true}
}

func HttpWithRetCode(retCode *int) HttpOptions {
	return HttpOptions{HttpOptionTypeRetCode, retCode}
}

// For standard SSE protocol, response are split by \n\n
// Which leads a bad performance when decoding, we need a larger chunk to store temporary data
// This option is used to enable length-prefixed mode, which is faster but less memory-friendly
// We uses following format:
//
//	| Field         | Size     | Description                     |
//	|---------------|----------|---------------------------------|
//	| Magic Number  | 1 byte   | Magic number identifier         |
//	| Reserved      | 1 byte   | Reserved field                  |
//	| Header Length | 2 bytes  | Header length (usually 0xa)    |
//	| Data Length   | 4 bytes  | Length of the data              |
//	| Reserved      | 6 bytes  | Reserved fields                 |
//	| Data          | Variable | Actual data content             |
//
//	| Reserved Fields | Header   | Data     |
//	|-----------------|----------|----------|
//	| 4 bytes total   | Variable | Variable |
//
// with the above format, we can achieve a better performance, avoid unexpected memory growth
func HttpUsingLengthPrefixed(using bool) HttpOptions {
	return HttpOptions{HttpOptionTypeUsingLengthPrefixed, using}
}
