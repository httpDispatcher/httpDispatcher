package utils

import (
	"MyError"
	"config"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/op/go-logging"
)

var QueryLogger = logging.MustGetLogger("query")
var ServerLogger = logging.MustGetLogger("server")

var queryformat = logging.MustStringFormatter(
	`%{time:2006-01-02T15:04:05} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{message}`,
)

var serverformat = logging.MustStringFormatter(
	`%{time:2006-01-02T15:04:05} %{longfile} ▶ %{level:.4s} %{id:03x}%{message}`,
)

func GetDebugLine() string {
	_, file, line, ok := runtime.Caller(1)
	if ok {
		// Truncate file name at last file name separator.
		if index := strings.LastIndex(file, "/"); index >= 0 {
			file = file[index+1:]
		} else if index = strings.LastIndex(file, "\\"); index >= 0 {
			file = file[index+1:]
		}
	} else {
		file = "???"
		line = 1
	}
	return "\t" + file + ":" + strconv.Itoa(line)
}

func CheckIPv4(ip string) {

}

func InitLogger() {
	loglevel := map[string]logging.Level{
		"DEBUG":    logging.DEBUG,
		"INFO":     logging.INFO,
		"NOTICE":   logging.NOTICE,
		"WARNING":  logging.WARNING,
		"ERROR":    logging.ERROR,
		"CRITICAL": logging.CRITICAL}
	qfd, e := os.OpenFile(config.RC.QueryLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if e != nil {
		fmt.Println("Open log file ", config.RC.QueryLog, " error: ", e.Error())
		os.Exit(1)
	}

	sfd, e := os.OpenFile(config.RC.ServerLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if e != nil {
		fmt.Println("Open log file ", config.RC.ServerLog, " error: ", e.Error())
		os.Exit(1)
	}

	backend1 := logging.NewLogBackend(qfd, "", 0)
	backend2 := logging.NewLogBackend(sfd, "", 0)

	backend1Formatter := logging.NewBackendFormatter(backend1, queryformat)
	backend2Formatter := logging.NewBackendFormatter(backend2, serverformat)

	backend1Leveled := logging.AddModuleLevel(backend1Formatter)
	backend1Leveled.SetLevel(loglevel[config.RC.LogLevel], "querylog")

	backend2Leveled := logging.AddModuleLevel(backend2Formatter)
	backend2Leveled.SetLevel(loglevel[config.RC.LogLevel], "serverlog")

	QueryLogger.SetBackend(backend1Leveled)
	ServerLogger.SetBackend(backend2Leveled)
}

//Convert net.IPNet to  startIP & endIP
func NetworkRange(network *net.IPNet) (net.IP, net.IP) {
	fmt.Println(network)
	//	os.Exit(2)
	if network == nil {
		return net.IPv4(0, 0, 0, 0), net.IPv4(0, 0, 0, 0)
	}
	netIP := network.IP.To4()
	firstIP := netIP.Mask(network.Mask)
	lastIP := net.IPv4(0, 0, 0, 0).To4()
	for i := 0; i < len(lastIP); i++ {
		lastIP[i] = netIP[i] | ^network.Mask[i]
	}
	return firstIP, lastIP
}

// Converts a 4 bytes IP into a 32 bit integer
func Ip4ToInt32(ip net.IP) uint32 {
	return uint32(binary.BigEndian.Uint32(ip.To4()))
}

// Converts 32 bit integer into a 4 bytes IP address
func Int32ToIP4(n uint32) net.IP {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(n))
	return net.IP(b)
}

//ParseEdnsIPNet, Parse ends data to *net.IPNet
func ParseEdnsIPNet(ip net.IP, mask uint8, family uint16) (*net.IPNet, *MyError.MyError) {
	cidr := strings.Join([]string{ip.String(), strconv.Itoa(int(mask))}, "/")
	_, ipnet, e := net.ParseCIDR(cidr)
	if e == nil {
		return ipnet, nil
	}
	return nil, MyError.NewError(MyError.ERROR_NOTVALID, e.Error())
}

//Parse *net.IPNet to ip(uint32) and mask(int)
func IpNetToInt32(ipnet *net.IPNet) (ip uint32, mask int) {
	if ipnet == nil {
		return uint32(1), int(1)
	}
	ip = Ip4ToInt32(ipnet.IP)
	mask, _ = ipnet.Mask.Size()
	return ip, mask
}

//Parse ip(uint32) and mask(int) to *net.IPNe
func Int32ToIpNet(ip uint32, mask int) (*net.IPNet, *MyError.MyError) {
	if mask < 0 || mask > 32 {
		return nil, MyError.NewError(MyError.ERROR_NOTVALID, "invalid mask error, param: "+strconv.Itoa(mask))
	}
	ipaddr := Int32ToIP4(ip)

	cidr := strings.Join([]string{ipaddr.String(), strconv.Itoa(mask)}, "/")
	_, ipnet, ok := net.ParseCIDR(cidr)
	if ok != nil {
		return nil, MyError.NewError(MyError.ERROR_NOTVALID, "ParseCIDR error, param: "+cidr)
	}
	return ipnet, nil
}

func StrToIP(s string) net.IP {
	return net.ParseIP(s)
}

func GetCIDRMaskWithUint32Range(startIp, endIp uint32) int {
	n := endIp - startIp
	x := 0
	key := uint32(1)
	if n != uint32(0) {
		for key > 0 {
			if (n & key) > 0 {
				x++
			}
			key = key << 1
		}
		return int(32 - x)
	}
	return 0
}
