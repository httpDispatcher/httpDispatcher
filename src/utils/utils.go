package utils

import (
	"MyError"
	"encoding/binary"
	"fmt"
	"log"
	"net"
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

//func IP4toInt(IPv4Address net.IP) int64 {
//	IPv4Int := big.NewInt(0)
//
//	IPv4Int.SetBytes(IPv4Address.To4())
//	fmt.Println(IPv4Int.Bytes())
//	fmt.Printf("%v", IPv4Int.Bytes())
//	return IPv4Int.Int64()
//}

func NetworkRange(network *net.IPNet) (net.IP, net.IP) {
	fmt.Println(network)
	//	os.Exit(2)
	netIP := network.IP.To4()
	firstIP := netIP.Mask(network.Mask)
	lastIP := net.IPv4(0, 0, 0, 0).To4()
	for i := 0; i < len(lastIP); i++ {
		lastIP[i] = netIP[i] | ^network.Mask[i]
	}
	return firstIP, lastIP
}

// Converts a 4 bytes IP into a 32 bit integer
func Ip4ToInt32(ip net.IP) int32 {
	return int32(binary.BigEndian.Uint32(ip.To4()))
}

// Converts 32 bit integer into a 4 bytes IP address
func Int32ToIP4(n int32) net.IP {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(n))
	return net.IP(b)
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

func ParseEdnsIPNet(ip net.IP, mask uint8, family uint16) (*net.IPNet, *MyError.MyError) {
	fmt.Println(mask)
	iplen := 0
	switch family {
	case 1:
		iplen = net.IPv4len
	case 2:
		iplen = net.IPv6len
	}
	n, i, ok := dtoi(string(mask), 0)
	if ip == nil || !ok || i != len(string(mask)) || n < 0 || n > 8*iplen {
		return nil, MyError.NewError(MyError.ERROR_NOTVALID, "ParseEdnsIPNet error, param: "+ip.String()+"/"+string(mask))
	}
	m := net.CIDRMask(n, 8*iplen)
	return &net.IPNet{IP: ip.Mask(m), Mask: m}, nil
}
