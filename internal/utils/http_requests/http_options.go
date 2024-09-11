package http_requests

import "io"

type HttpOptions struct {
	Type  string
	Value interface{}
}

// milliseconds
func HttpWriteTimeout(timeout int64) HttpOptions {
	return HttpOptions{"write_timeout", timeout}
}

// milliseconds
func HttpReadTimeout(timeout int64) HttpOptions {
	return HttpOptions{"read_timeout", timeout}
}

func HttpHeader(header map[string]string) HttpOptions {
	return HttpOptions{"header", header}
}

// which is used for params with in url
func HttpParams(params map[string]string) HttpOptions {
	return HttpOptions{"params", params}
}

// which is used for POST method only
func HttpPayload(payload map[string]string) HttpOptions {
	return HttpOptions{"payload", payload}
}

// which is used for POST method only
func HttpPayloadText(payload string) HttpOptions {
	return HttpOptions{"payloadText", payload}
}

// which is used for POST method only
func HttpPayloadJson(payload interface{}) HttpOptions {
	return HttpOptions{"payloadJson", payload}
}

// which is used for POST method only
// payload follows the form data format, and files is a map from filename to file
func HttpPayloadMultipart(payload map[string]string, files map[string]io.Reader) HttpOptions {
	return HttpOptions{"payloadMultipart", map[string]interface{}{
		"payload": payload,
		"files":   files,
	}}
}

func HttpRaiseErrorWhenStreamDataNotMatch(raise bool) HttpOptions {
	return HttpOptions{"raiseErrorWhenStreamDataNotMatch", raise}
}

func HttpWithDirectReferer() HttpOptions {
	return HttpOptions{"directReferer", true}
}

func HttpWithRetCode(retCode *int) HttpOptions {
	return HttpOptions{"retCode", retCode}
}
