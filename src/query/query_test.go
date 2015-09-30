package query

import (
	//	"Testing"
	"fmt"
	"github.com/miekg/dns"
	//	"query"
)

const (
	config_file = "/etc/resolv.conf"
)

func TestQueryNS() {
	cf, err := dns.ClientConfigFromFile("/etc/resolv.conf")
	if err != nil {
		fmt.Println(err)
		return
	}
	r, err := queryNS("api.weibo.cn", cf)
	fmt.Println(r)
}

func main() {
	//	testing.RunTests()
	//	TestQueryNS()

}
