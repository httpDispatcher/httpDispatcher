package MyError

//	"fmt"

const (
	ERROR_PARAM     = "ERROR_PARAM"
	ERROR_NORESULT  = "ERROR_NORESULT"
	ERROR_UNKNOWN   = "ERROR_UNKNOWN"
	ERROR_SUBDOMAIN = "ERROR_SUBDOMAIN"
	ERROR_TYPE      = "ERROR_TYPE"
	ERROR_NOTFOUND  = "ERROR_NOTFOUND"
	ERROR_NOTVALID  = "ERROR_NOTVALID"
	ERROR_CNAME     = "ERROR_CNAME"
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
