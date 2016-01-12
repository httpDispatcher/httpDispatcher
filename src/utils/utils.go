package utils

import (
	"MyError"
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
)

var Logger log.Logger

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
	buf := new(bytes.Buffer)
	// Every line is indented at least one tab.
	buf.WriteByte('\t')
	fmt.Fprintf(buf, "%s:%d: ", file, line)
	return buf.String()
}

func CheckIPv4(ip string) {

}

func InitUitls() {
	Logger := log.New(os.Stdout, "httpDispacher", log.Ldate|log.Ltime|log.Llongfile)
	Logger.Println("Starting httpDispacher...")
}

// Bigger than we need, not too big to worry about overflow
const big = 0xFFFFFF

func dtoi(s string, i0 int) (n int, i int, ok bool) {
	n = 0
	for i = i0; i < len(s) && '0' <= s[i] && s[i] <= '9'; i++ {
		n = n*10 + int(s[i]-'0')
		if n >= big {
			return 0, i, false
		}
	}
	if i == i0 {
		return 0, i, false
	}
	return n, i, true
}

//func IP4toInt(IPv4Address net.IP) int64 {
//	IPv4Int := big.NewInt(0)
//
//	IPv4Int.SetBytes(IPv4Address.To4())
//	fmt.Println(IPv4Int.Bytes())
//	fmt.Printf("%v", IPv4Int.Bytes())
//	return IPv4Int.Int64()
//}

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
	//	fmt.Println(family)
	//	iplen := 0
	//	switch family {
	//	case 1:
	//		iplen = net.IPv4len
	//	case 2:
	//		iplen = net.IPv6len
	//	}
	//	n, i, ok := dtoi(string(mask), 0)
	//	if ip == nil || !ok || i != len(string(mask)) || n < 0 || n > 8*iplen {
	//		return nil, MyError.NewError(MyError.ERROR_NOTVALID, "ParseEdnsIPNet error, param: "+ip.String()+"/"+string(mask))
	//	}
	//	m := net.CIDRMask(n, 8*iplen)
	//	return &net.IPNet{IP: ip.Mask(m), Mask: m}, nil
	cidr := strings.Join([]string{ip.String(), strconv.Itoa(int(mask))}, "/")
	_, ipnet, e := net.ParseCIDR(cidr)
	if e == nil {
		return ipnet, nil
	}
	return nil, MyError.NewError(MyError.ERROR_NOTVALID, e.Error())
}

//Parse *net.IPNet to ip(uint32) and mask(int)
func IpNetToInt32(ipnet *net.IPNet) (ip uint32, mask int) {
	//	ai, bi := uint32(0), uint32(0)
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
			//		fmt.Println(key)
		}
		fmt.Println(GetDebugLine(), " GetCIDRMaskWithUint32Range : ",
			" startIP: ", startIp, " endIp: ", endIp, " Result: ", 32-x)
		return int(32 - x)
	}
	return 0
}
