package response

import (
	"fmt"
	"net/http"
	"time"
)

const (
	ErrCodeUnknown = iota
	ErrCodeInvalidArguments
	ErrCodeAlreadyExists
	ErrCodeNotFound
)

type ApiError struct {
	HttpStatus int         `json:"http_status,omitempty" yaml:"http_status,omitempty"`
	ErrorCode  int         `json:"code" yaml:"code"`
	Message    interface{} `json:"message,omitempty" yaml:"message,omitempty"`
	Timestamp  interface{} `json:"timestamp,omitempty" yaml:"timestamp,omitempty"`
}

func (e ApiError) Error() string {
	return e.String()
}

func (e ApiError) String() string {
	msg := fmt.Sprintf("%s", e.Message)

	if e.HttpStatus > 0 {
		msg = fmt.Sprintf("[HTTP_STATUS:%d] %s", e.HttpStatus, msg)
	}

	if e.ErrorCode > 0 {
		msg = fmt.Sprintf("[CODE:%d] %s", e.ErrorCode, msg)
	}

	if e.Timestamp != nil {
		msg = fmt.Sprintf("[%s] %s", e.Timestamp, msg)
	}

	return fmt.Sprintf("%s", msg)
}

func HttpCode(errorCode int) int {
	switch errorCode {
	case ErrCodeInvalidArguments:
		return http.StatusBadRequest
	case ErrCodeAlreadyExists:
		return http.StatusBadRequest
	case ErrCodeNotFound:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

func NewApiError(code int, msg string) *ApiError {
	return &ApiError{
		ErrorCode:  code,
		HttpStatus: HttpCode(code),
		Message:    msg,
		Timestamp:  time.Now().Format(time.RFC1123),
	}
}
