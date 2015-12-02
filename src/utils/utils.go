package utils

import (
	"log"
	"os"
)

const (
	DEFAULT_SOURCEMASK  = 32
	DEFAULT_SOURCESCOPE = 0
)

var Logger log.Logger

func CheckIPv4(ip string) {

}

func GetClientIP() string {
	return "124.207.129.171"
}

func InitUitls() {
	Logger := log.New(os.Stdout, "httpDispacher", log.Ldate|log.Ltime|log.Llongfile)
	Logger.Println("Starting httpDispacher...")
}

func UnpackIPCIDR(i string) {

}
