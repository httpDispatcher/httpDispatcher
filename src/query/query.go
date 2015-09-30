package query

import (
	"errors"
	"strconv"
	//	"flag"
	"fmt"
	"github.com/miekg/dns"
	//	"log"
	"net"
)

const (
	NS_SERVER_PORT = "53"
)

type DomainConfig struct {
	DomainName           string
	AuthoritativeServers []string
	Port                 string
	Ttl                  string
}

type DomainRecord struct {
	dc      *DomainConfig
	Records []string
	IpStart uint64
	IpEnd   uint64
	NetMask uint8
}

func (dc *DomainConfig) SetDomain(d string) (bool, error) {
	if _, ok := dns.IsDomainName(d); ok {
		dc.DomainName = dns.Fqdn(d)
		return true, nil
	} else {
		return false, errors.New(d + " is not domain name")
	}
}

func (dc *DomainConfig) SetTtl(t string) error {
	if ti, e := strconv.Atoi(t); e == nil {
		if ti > 0 && ti < 1024 {
			dc.Ttl = t
		} else {
			return errors.New(t + "is not permited")
		}
	}
	return nil
}

func (dc *DomainConfig) SetAS(as []string) error {
	for _, s := range as {
		if ip := net.ParseIP(s); ip != nil {
			dc.AuthoritativeServers = append(dc.AuthoritativeServers, s)
		} else {
			return errors.New(s + " is not standard ip string")
		}
	}
	return nil
}

func QueryNS(domain string, dc *DomainConfig) ([]string, error) {
	var ns_rr []string
	c := new(dns.Client)
	m := new(dns.Msg)
	m.SetQuestion(domain, dns.TypeNS)
	//@TODO: randow cf.Servers for load banlance
	r, _, e := c.Exchange(m, dc.AuthoritativeServers[0]+":"+dc.Port)

	if e != nil {
		return nil, e
	}
	for _, a := range r.Answer {
		if ns, ok := a.(*dns.NS); ok {
			ns_rr = append(ns_rr, dns.Field(ns, 1))
		}
	}
	return ns_rr, nil
}

func QueryA(domain string, dc *DomainConfig, edns0subnet dns.EDNS0) ([]string, *dns.EDNS0_SUBNET, error) {
	var a_rr []string
	c := new(dns.Client)
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), dns.TypeA)
	o := new(dns.OPT)
	o.Hdr.Name = "."
	o.Hdr.Rrtype = dns.TypeOPT
	o.Option = append(o.Option, edns0subnet)
	m.Extra = append(m.Extra, o)
	et := new(dns.EDNS0_SUBNET)
	r, _, e := c.Exchange(m, dc.AuthoritativeServers[0]+":"+dc.Port)
	if e != nil {
		return nil, et, e
	}
	//	fmt.Println("------Response.Answer------")
	//	fmt.Println(r.Answer)
	//	fmt.Println("------Response.Ns-------")
	//	fmt.Println(r.Ns)
	//	fmt.Println("-----Response.String()-------")
	//	fmt.Println(r.String())
	//	fmt.Println("------------>>>>>")
	//	for _, ot := range r.Extra {
	//		fmt.Print("1:")
	//		fmt.Println(ot.Header().Name)
	//		fmt.Print("2:")
	//		fmt.Println(ot.Header().Class)
	//		fmt.Print("3:")
	//		fmt.Println(ot.Header().Rdlength)
	//		fmt.Print("4:")
	//		fmt.Println(ot.Header().Rrtype)
	//		fmt.Print("5:")
	//		fmt.Println(ot.Header().Ttl)
	//		fmt.Println(".........................")
	//	}

	if r_opt := r.IsEdns0(); r_opt != nil {
		//		fmt.Print("Return edns0subnet msg: ")
		//		fmt.Println(r.Extra)
		//		fmt.Println("------------")
		for _, ot1 := range r_opt.Option {
			fmt.Println(ot1.Option())
			if ot1.Option() == dns.EDNS0SUBNET || ot1.Option() == dns.EDNS0SUBNETDRAFT {

				et = ot1.(*dns.EDNS0_SUBNET)
				//				fmt.Println(et.Address)
				//				fmt.Println(et.Code)
				//				fmt.Println(et.DraftOption)
				//				fmt.Println(et.Option())
				//				fmt.Println(et.SourceNetmask)
				//				fmt.Println(et.SourceScope)
				//				fmt.Println(et.Family)
			} else {
				et = nil
			}
		}
		//		fmt.Println("===============")
		//		fmt.Print("1:")
		//		fmt.Println(r_opt)
		//		fmt.Print("2:")
		//		fmt.Println(r_opt.Hdr)
		//		fmt.Print("3:")
		//		fmt.Println(r_opt.Option)
		//		fmt.Println("^^^^^^^^^^^^^^")

	} else {
		et = nil
		//		fmt.Println("Not retuen edns0subnet msg")

	}

	for _, a := range r.Answer {
		if aa, ok := a.(*dns.A); ok {
			//			fmt.Println(dns.Field(aa, 1))
			a_rr = append(a_rr, dns.Field(aa, 1))
		}
	}

	return a_rr, et, nil
}

func PackEdns0Subnet(ip string, sourcenetmask uint8, sourcescope uint8) *dns.EDNS0_SUBNET {
	edns0 := new(dns.EDNS0_SUBNET)
	edns0.Code = dns.EDNS0SUBNET
	edns0.SourceScope = sourcescope
	edns0.Address = net.ParseIP(ip).To4()
	edns0.SourceNetmask = sourcenetmask
	edns0.Family = 1
	return edns0
}

func UnpackEdns0Subnet(opt *dns.OPT) {

}

func Test(domain string, edns0ip string, dc *DomainConfig) {
	fmt.Print("DNS Server: ")
	fmt.Println(dc.AuthoritativeServers[0])
	fmt.Print("Request domain: ")
	fmt.Println(domain)

	edns0 := PackEdns0Subnet(edns0ip, 32, 0)
	fmt.Println("-------------------------------------")
	fmt.Print("Request edns0 subnet pack: ")
	fmt.Println(edns0)
	fmt.Println("-------------------------------------")
	m, et, e := QueryA(dns.Fqdn(domain), dc, edns0)
	if e != nil {
		fmt.Println(e)
	}
	fmt.Println("-------------------------------------")
	fmt.Print("Return DNS A Record: ")
	fmt.Println(m)
	fmt.Println("Return ends0subnet :")
	fmt.Println(et)

}

//func doQuery(domain string, aNameServers []string) []string {
//	return "hello,world"
//}

//func main() {
//	var ns_server, query_domain, subnet_ip string
//	flag.StringVar(&ns_server, "ns", "8.8.8.8", "The name server which will be queried ")
//	flag.StringVar(&query_domain, "q", "api.weibo.cn", "The domain name which will be queried")
//	flag.StringVar(&subnet_ip, "ip", "8.8.8.8", "The edns0_subnet address which will be used")

//	flag.Parse()

//	fmt.Println(">>>ns == " + ns_server)
//	fmt.Println(">>> q == " + query_domain)
//	fmt.Println(">>>ip == " + subnet_ip)

//	cf := new(dns.ClientConfig)
//	cf.Servers = append(cf.Servers, ns_server)
//	cf.Port = "53"

//	Test(ns_server, query_domain, subnet_ip, cf)

//}
