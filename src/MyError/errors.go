package MyError

//	"fmt"

const (
	ERROR_PARAM        = "ERROR_PARAM"
	ERROR_NORESULT     = "ERROR_NORESULT"
	ERROR_UNKNOWN      = "ERROR_UNKNOWN"
	ERROR_SUBDOMAIN    = "ERROR_SUBDOMAIN"
	ERROR_TYPE         = "ERROR_TYPE"
	ERROR_NOTFOUND     = "ERROR_NOTFOUND"
	ERROR_NOTVALID     = "ERROR_NOTVALID"
	ERROR_CNAME        = "ERROR_CNAME"
	ERROR_NORESOLVCONF = "ERROR_NORESOLVCONF"
)

type MyError struct {
	errorNo string
	msg     string
}

func NewError(errno, Msg string) *MyError {
	return &MyError{errorNo: errno, msg: Msg}
}

func (e *MyError) Error() string {
	return "Error -> : " + e.errorNo + " .. " + e.msg
}
