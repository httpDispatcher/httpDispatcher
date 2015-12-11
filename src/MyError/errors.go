package MyError

import (
//	"fmt"
)

const (
	ERROR_PARAM     = "ERROR_PARAM"
	ERROR_NORESULT  = "ERROR_NORESULT"
	ERROR_UNKNOWN   = "ERROR_UNKNOWN"
	ERROR_SUBDOMAIN = "ERROR_SUBDOMAIN"
)

type MyError struct {
	ErrorNo string
	Msg     string
}

func NewError(errno, Msg string) *MyError {
	return &MyError{ErrorNo: errno, Msg: Msg}
}

func (e *MyError) Error() string {
	return "Error -> : " + e.ErrorNo + " .. " + e.Msg
}
