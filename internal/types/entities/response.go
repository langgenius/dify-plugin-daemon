package entities

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func NewSuccessResponse(data any) *Response {
	return &Response{
		Code:    0,
		Message: "success",
		Data:    data,
	}
}

func NewDaemonErrorResponse(code int, message string, args ...any) *Response {
	resp := &Response{
		Code:    code,
		Message: message,
		Data:    nil,
	}
	if len(args) > 0 {
		resp.Data = args[0]
	}
	return resp
}

type GenericResponse[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}
