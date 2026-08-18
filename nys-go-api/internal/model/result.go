package model

// Result matches the JSON contract used by the Java services.
type Result struct {
	IsError bool   `json:"isError"`
	ErrMsg  string `json:"errMsg"`
	Result  any    `json:"result"`
}

func Success(value any) Result {
	return Result{IsError: false, ErrMsg: "", Result: value}
}

func Failure(message string) Result {
	return Result{IsError: true, ErrMsg: message, Result: nil}
}
