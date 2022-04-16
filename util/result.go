package util

type RequestError struct {
	Code    string
	Message string
}
type RequestResult struct {
	Error   RequestError
	Success bool
	Data    any
}

var Result *RequestResult = &RequestResult{}

func (*RequestResult) OK(data any) *RequestResult {
	return &RequestResult{
		Success: true,
		Data:    data,
	}
}

func (*RequestResult) Fail(code, message string) *RequestResult {
	return &RequestResult{
		Success: false,
		Error: RequestError{
			Code:    code,
			Message: message,
		},
	}
}
