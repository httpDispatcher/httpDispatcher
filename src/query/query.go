package query

import (
	"container/list"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"
	"utils"

	"net"

	"github.com/miekg/dns"
)

const (
	DOMAIN_MAX_LABEL = 16
)

const (
	NS_SERVER_PORT      = "53"
	DEFAULT_RESOLV_FILE = "/etc/resolv.conf"
)

func GeneralQuery(
	domainName,
	domainResolverIP,
	domainResolverPort string,
	queryType uint16,
	queryOpt *dns.OPT) (*dns.Msg, error) {
	if _, ok := dns.IsDomainName(domainName); !ok {
		return nil, errors.New("Param is not a domain name : " + domainName)
	}

	ds, dp, e := GetDomainConfig(domainName)
	if e != nil {
		return nil, e
	}
	fmt.Println("++++++++begin ds dp+++++++++")
	fmt.Println(ds + "......" + dp + "  .......")
	fmt.Println("++++++++end ds dp+++++++++")

	c := new(dns.Client)
	c.DialTimeout = 60 * time.Second
	c.WriteTimeout = 60 * time.Second
	c.ReadTimeout = 60 * time.Second
	c.Net = "udp"

	m := new(dns.Msg)
	m.AuthenticatedData = true
	m.RecursionDesired = true
	m.SetQuestion(dns.Fqdn(domainName), queryType)
	if queryOpt != nil {
		m.Extra = append(m.Extra, queryOpt)
	}
	r, _, e := c.Exchange(m, ds+":"+dp)

	if e != nil {
		fmt.Println(e.Error())
		os.Exit(1)
		return nil, e
	}
	return r, nil
}

func ParseCNAME(r *dns.Msg) (bool, *dns.CNAME, uint32) {
	for _, a := range r.Answer {
		if cname, ok := a.(*dns.CNAME); ok {
			fmt.Println("CNAME Found!: " + cname.Hdr.Name + " <--> " + dns.Field(cname, 1))
			ttl := cname.Hdr.Ttl
			fmt.Println(cname.Target)
			return true, cname, ttl
		}
	}
	return false, nil, 0
}

func ParseNS(ns []dns.RR) ([]string, uint32, error) {
	var ns_rr []string
	var ttl uint32 = 0
	if len(ns) > 0 {
		for _, n_s := range ns {
			if x, ok := n_s.(*dns.NS); ok {
				ns_rr = append(ns_rr, x.Ns)
				ttl = x.Hdr.Ttl
			}
		}
	}

	return ns_rr, ttl, nil
}

func LoopForQueryNS(d string) ([]string, error) {
	if _, ok := dns.IsDomainName(d); !ok {
		return nil, errors.New(d + " is not a domain name")
	}
	r := dns.SplitDomainName(d)

	if cap(r) > DOMAIN_MAX_LABEL {
		return nil, errors.New(d + " is too long")
	}
	var ns_arr []string = nil
	for (cap(r) > 1) && (cap(ns_arr) < 1) {
		ra, e := QueryNS(strings.Join(r, "."))
		if e != nil {
			continue
		}
		arr, _, _ := ParseNS(ra)
		ns_arr = append(ns_arr, arr...)
		if cap(ns_arr) < 1 {
			r = r[1:]
		}
	}
	return ns_arr, nil
}

func QueryNS(d string) ([]dns.RR, error) {

	//@TODO: randow cf.Servers for load banlance
	ds, dp, e := GetDomainConfig(d)

	if e == nil {
		fmt.Println(ds + dp)
	} else {
	}

	r, e := GeneralQuery(d, ds, dp, dns.TypeNS, nil)
	if e != nil {
		return nil, e
	}
	//	if len(r.Ns) > 0 {
	//		r_arr = append(r_arr, r.Ns...)
	//	}
	//	if len(r.Answer) > 0 {
	//		r_arr = append(r_arr, r.Answer...)
	//	}
	return r.Answer, nil
}

func QueryCNAME(d string, isEdns0 bool) {
	if _, is := dns.IsDomainName(d); !is {
		//		return nil, errors.New("param error " + d + " is not a  domain name")
		//		fmt.Println(d + " is not a domain name")
	}
	var o *dns.OPT = nil
	if isEdns0 {
		ip := utils.GetClientIP()
		o = PackEdns0SubnetOPT(ip, utils.DEFAULT_SOURCEMASK, utils.DEFAULT_SOURCESCOPE)
	} else {
		o = nil
	}
	ds, dp, _ := GetDomainConfig(d)
	//	fmt.Println("dp :" + dp)
	r, e := GeneralQuery(d, ds, dp, dns.TypeCNAME, o)
	if e != nil {
		fmt.Println(e.Error())
		os.Exit(1)
	}
	if r.Rcode == 0 && r.Truncated == false {
		//		for _, r_t := range r.Answer {
		fmt.Println(r.Answer)
		//		}
	} else {
		fmt.Print("r.Rcode:")
		fmt.Println(r.Rcode)
		fmt.Print("r.Truncated:")
		fmt.Println(r.Truncated)
		fmt.Print("r :")
		fmt.Println(r)
		os.Exit(1)
	}
}

func UnpackCNAMEAnswer(a dns.RR, l *list.List) *list.List {

	if ac, ok := a.(*dns.CNAME); ok {
		fmt.Println("CNAME:")
		fmt.Println(ac)
		//		l.PushBack(a)
	}
	return l
}

func UnpackAAnswer(a_rr []string, a dns.RR) ([]string, uint32) {
	var ttl uint32 = 0
	if aa, ok := a.(*dns.A); ok {
		a_rr = append(a_rr, aa.A.To4().String())
		if ttl == 0 {
			ttl = aa.Hdr.Ttl
		}
	}
	return a_rr, ttl
}

func ParseA(answer []dns.RR) ([]string, uint32, error) {
	var a_rr []string
	var ttl uint32 = 0
	fmt.Print("dns.IsRRset(answer):")
	fmt.Println(dns.IsRRset(answer))
	for _, a := range answer {
		fmt.Println(reflect.TypeOf(a))
		switch a.Header().Rrtype {
		case dns.TypeCNAME:
			{
				fmt.Print("dns.TypeCNAME :")
				fmt.Println(a.String())
			}
		case dns.TypeA:
			{
				a_rr, ttl = UnpackAAnswer(a_rr, a)
			}
		}
	}
	return a_rr, ttl, nil

}

func QueryA(domain string, isEdns0 bool) ([]dns.RR, *dns.EDNS0_SUBNET, error) {
	//	var a_rr []string
	var o *dns.OPT = nil
	if isEdns0 {
		ip := utils.GetClientIP()
		o = PackEdns0SubnetOPT(ip, utils.DEFAULT_SOURCEMASK, utils.DEFAULT_SOURCESCOPE)
	} else {
		o = nil
	}
	ds, dp, e := GetDomainConfig(domain)
	r, e := GeneralQuery(domain, ds, dp, dns.TypeA, o)

	if e != nil {
		utils.Logger.Println(r)
		return nil, nil, e
	}
	//	fmt.Println(r)
	et := new(dns.EDNS0_SUBNET)

	if isEdns0 {
		if r_opt := r.IsEdns0(); r_opt != nil {
			for _, ot1 := range r_opt.Option {
				if ot1.Option() == dns.EDNS0SUBNET || ot1.Option() == dns.EDNS0SUBNETDRAFT {
					et = ot1.(*dns.EDNS0_SUBNET)
				} else {
					et = nil
				}
			}
		} else {
			et = nil
		}
	}
	fmt.Print("r.Answer:")
	fmt.Println(r.Answer)
	//	for _, a := range r.Answer {
	//		fmt.Println(reflect.TypeOf(a))
	//		if aa, ok := a.(*dns.A); ok {
	//			a_rr = append(a_rr, dns.Field(aa, 1))
	//		}
	//	}
	//	fmt.Println(a_rr)

	//	fmt.Println(r.Answer)
	return r.Answer, et, nil
}

func PackEdns0Subnet(ip string, sourceNetmask uint8, sourceScope uint8) *dns.EDNS0_SUBNET {
	edns0 := new(dns.EDNS0_SUBNET)
	edns0.Code = dns.EDNS0SUBNET
	edns0.SourceScope = sourceScope
	edns0.Address = net.ParseIP(ip).To4()
	edns0.SourceNetmask = sourceNetmask
	edns0.Family = 1
	return edns0
}

func PackEdns0SubnetOPT(ip string, sourceNetmask, sourceScope uint8) *dns.OPT {

	edns0subnet := PackEdns0Subnet(ip, sourceNetmask, sourceScope)
	o := new(dns.OPT)
	o.Hdr.Name = "."
	o.Hdr.Rrtype = dns.TypeOPT
	o.Option = append(o.Option, edns0subnet)
	fmt.Print("o.string(): ")
	fmt.Println(o.String())
	return o

}

func UnpackEdns0Subnet(opt *dns.OPT) {

}
