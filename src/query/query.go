package query

import (
	"MyError"
	"fmt"
	"net"
	"strings"
	"time"
	"utils"

	"github.com/miekg/dns"
)

const (
	DOMAIN_MAX_LABEL    = 16
	NS_SERVER_PORT      = "53"
	DEFAULT_RESOLV_FILE = "/etc/resolv.conf"
	UDP                 = "udp"
	TCP                 = "tcp"
)

// type RR , dns record
type RR struct {
	Record []string
	Rrtype uint16
	Class  uint16
	Ttl    uint32
}

// type Query , dns Query type and Query result
type Query struct {
	QueryType     uint16
	NS            string
	IsEdns0Subnet bool
	Msg           *dns.Msg
}

// func NewQuery , build a Query Instance for Querying
func NewQuery(t uint16, ns string, edns0Subnet bool) *Query {
	return &Query{
		QueryType:     t,
		NS:            ns,
		IsEdns0Subnet: edns0Subnet,
		Msg:           new(dns.Msg),
	}
}

func Check_DomainName(d string) (int, bool) {
	return dns.IsDomainName(d)
}

// General Query for dns upstream query
// param: t string ["tcp"|"udp]
// 		  queryType uint16 dns.QueryType
func DoQuery(
	domainName,
	domainResolverIP,
	domainResolverPort string,
	queryType uint16,
	queryOpt *dns.OPT, t string) (*dns.Msg, *MyError.MyError) {

	//	fmt.Println("++++++++begin ds dp+++++++++")
	//	fmt.Println(domainResolverIP + "......" + domainResolverPort + "  .......")
	//	fmt.Println("++++++++end ds dp+++++++++")
	c := &dns.Client{
		DialTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		ReadTimeout:  3 * time.Second,
		Net:          t,
	}

	m := &dns.Msg{}
	m.AuthenticatedData = true
	m.RecursionDesired = true
	m.SetQuestion(dns.Fqdn(domainName), queryType)

	if queryOpt != nil {
		m.Extra = append(m.Extra, queryOpt)
	}
	r := &dns.Msg{}
	var ee error
	for l := 0; l < 3; l++ {
		r, _, ee = c.Exchange(m, domainResolverIP+":"+domainResolverPort)
		if ee != nil {
			//		fmt.Println("errrororororororororo:")
			fmt.Println(utils.GetDebugLine(), ee.Error())
			//		os.Exit(1)
			if l >= 2 {
				return nil, MyError.NewError(MyError.ERROR_UNKNOWN, ee.Error())
			}
		}
	}
	return r, nil
}

// Filter CNAME record in dns.Msg.Answer message.
// 	if r(*dns.Msg.Answer) includes (*dns.CNAME) , than return the CNAME record array.
func ParseCNAME(c []dns.RR, d string) ([]*dns.CNAME, bool) {
	fmt.Println(utils.GetDebugLine(), "ParseCNAME line 99: ", c)
	var cname_a []*dns.CNAME
	for _, a := range c {
		if cname, ok := a.(*dns.CNAME); ok {
			fmt.Println(utils.GetDebugLine(), "ParseCNAME: line 103 : ", cname.Hdr.Name, d)
			if cname.Hdr.Name == dns.Fqdn(d) {
				cname_a = append(cname_a, cname)
			} else {
				fmt.Println(utils.GetDebugLine(), "ParseCNAME: line 106: ", cname)
			}
		}
	}
	if len(cname_a) > 0 {
		return cname_a, true
	}
	return nil, false
}

// Parse dns.Msg.Answer in dns response msg that use TypeNS as request type.
func ParseNS(ns []dns.RR) (bool, []*dns.NS) {
	var ns_rr []*dns.NS
	for _, n_s := range ns {
		if x, ok := n_s.(*dns.NS); ok {
			ns_rr = append(ns_rr, x)
		}
	}
	if cap(ns_rr) > 0 {
		return true, ns_rr
	}
	return false, nil
}

func ParseA(a []dns.RR, d string) ([]*dns.A, bool) {
	var a_rr []*dns.A
	for _, aa := range a {
		if x, ok := aa.(*dns.A); ok {
			if x.Hdr.Name == dns.Fqdn(d) {
				a_rr = append(a_rr, x)
			} else {
				fmt.Println(utils.GetDebugLine(), "ParseA: line 135: ", x)
			}
		}
	}
	if cap(a_rr) > 0 {
		return a_rr, true
	}
	return nil, false
}

func GenerateParentDomain(d string) (string, *MyError.MyError) {
	x := dns.SplitDomainName(d)
	if cap(x) > 1 {
		//		fmt.Println(x)
		return strings.Join(x[1:], "."), nil
	} else {
		return d, MyError.NewError(MyError.ERROR_NORESULT, d+" has no subdomain")
	}
	return d, MyError.NewError(MyError.ERROR_UNKNOWN, d+" unknown error")
}

// Loop For Query the domain name d's NS servers
// if d has no NS server, Loop for the parent domain's name server
func LoopForQueryNS(d string) ([]*dns.NS, *MyError.MyError) {
	if _, ok := Check_DomainName(d); !ok {
		return nil, MyError.NewError(MyError.ERROR_PARAM, d+" is not a domain name")
	}
	r := dns.SplitDomainName(d)
	if cap(r) > DOMAIN_MAX_LABEL {
		return nil, MyError.NewError(MyError.ERROR_PARAM, d+" is too long")
	}
	for cap(r) > 1 {
		ra, e := QueryNS(strings.Join(r, "."))
		if e != nil {
			//TODO: Log errors
			//			fmt.Println(e.Error())
			if e.ErrorNo == MyError.ERROR_NORESULT {
				r = r[1:] //TODO: this is not safe !
			}
			continue
		} else if cap(ra) >= 1 {
			return ra, nil
		} else {
			r = r[1:]
		}
	}
	return nil, MyError.NewError(MyError.ERROR_NORESULT, "Loop find Ns for Domain name "+d+" No Result")
}

func QuerySOA(d string) (*dns.SOA, []*dns.NS, *MyError.MyError) {
	if _, ok := Check_DomainName(d); !ok {
		return nil, nil, MyError.NewError(MyError.ERROR_PARAM, d+" is not a domain name")
	}
	cf, el := dns.ClientConfigFromFile("/etc/resolv.conf")
	if el != nil {
		return nil, nil, MyError.NewError(MyError.ERROR_UNKNOWN, "Get dns config from file /etc/resolv.conf failed")
	}

	var soa *dns.SOA
	var ns_a []*dns.NS
	for c := 0; (soa == nil) && (c < 3); c++ {

		soa, ns_a = nil, nil
		r, e := DoQuery(d, cf.Servers[0], cf.Port, dns.TypeSOA, nil, UDP)
		//		fmt.Println(r)
		if e != nil {
			//			c++
			continue
		} else {
			var rr []dns.RR
			if r.Answer != nil {
				rr = append(rr, r.Answer...)
			}
			if r.Ns != nil {
				rr = append(rr, r.Ns...)
			}
			soa, ns_a, e = ParseSOA(d, rr)
			if e != nil {
				switch e.ErrorNo {
				case MyError.ERROR_SUBDOMAIN, MyError.ERROR_NOTVALID:
					var ee *MyError.MyError
					d, ee = GenerateParentDomain(d)
					if ee != nil {
						if ee.ErrorNo == MyError.ERROR_NORESULT {
							//							fmt.Println(ee)
							//							continue
						}
						return nil, nil, MyError.NewError(MyError.ERROR_NORESULT,
							d+" has no SOA record "+" because of "+ee.Error())
					}
					continue
				case MyError.ERROR_NORESULT:
					//					c++
					fmt.Println(utils.GetDebugLine(), e)
					fmt.Println("+++++++++++++++++++++++++++++++++++")
					continue
				default:
					//					c++
					fmt.Println(utils.GetDebugLine(), ".....................")
					fmt.Println(utils.GetDebugLine(), e)
					continue
					//					return nil, nil, e
				}
			} else {
				if cap(ns_a) < 1 {
					fmt.Println(utils.GetDebugLine(), "QuerySOA: line 223: cap(ns_a)<1, need QueryNS ", soa.Hdr.Name)
					ns_a, e = QueryNS(soa.Hdr.Name)
					if e != nil {
						//TODO: do some log
					}
				}
				//				fmt.Println("============xxxxxx================")
				fmt.Println(utils.GetDebugLine(), "QuerySOA: line 230 ", soa, "\n Also 230:", ns_a)
				return soa, ns_a, nil
			}
		}
	}
	return nil, nil, MyError.NewError(MyError.ERROR_UNKNOWN, d+" QuerySOA faild with unknow error")
}

func ParseSOA(d string, r []dns.RR) (*dns.SOA, []*dns.NS, *MyError.MyError) {
	var soa *dns.SOA
	var ns_a []*dns.NS
	for _, v := range r {
		vh := v.Header()
		if vh.Name == dns.Fqdn(d) || dns.IsSubDomain(vh.Name, dns.Fqdn(d)) {
			switch vh.Rrtype {
			case dns.TypeSOA:
				if vv, ok := v.(*dns.SOA); ok {
					fmt.Print(utils.GetDebugLine(), "ParseSOA:line 245 ")
					fmt.Println(utils.GetDebugLine(), vv)
					soa = vv
				}
			case dns.TypeNS:
				if vv, ok := v.(*dns.NS); ok {
					ns_a = append(ns_a, vv)
				}
			default:
				fmt.Print(utils.GetDebugLine(), "PasreSOA: line 254 ")
				fmt.Println(utils.GetDebugLine(), v)
			}
		} else {
			fmt.Print(utils.GetDebugLine(), "ParseSOA 258 ")
			fmt.Println(utils.GetDebugLine(), vh.Name+" not match "+d)
			return nil, nil, MyError.NewError(MyError.ERROR_NOTVALID, d+" has no SOA record,try parent")
		}

	}
	if soa != nil {
		return soa, ns_a, nil
	} else {
		return nil, nil, MyError.NewError(MyError.ERROR_NORESULT, "No SOA record for domain "+d)
	}
}

// Preparation for Query A and CNAME / NS record.
// param d 		: domain name to query
// 		isEdns0 : either use edns0_subnet or not
// return :
//		*MyError.Myerror
//		domain name server ip
//		domain name server port
//		*dns.OPT (for edns0_subnet)
func preQuery(d, srcIP string) (*dns.OPT, *MyError.MyError) {
	if _, ok := Check_DomainName(d); !ok {
		//		return "", "", nil, MyError.NewError(MyError.ERROR_PARAM, d+" is not a domain name!")
	}
	//	ds, dp, e := GetDomainResolver(d)
	//	if e != nil {
	//		return "", "", nil, e
	//	}

	var o *dns.OPT
	if len(srcIP) > 0 {
		//TODO:modify GetClientIP func to use Server package
		//Done!

		o = PackEdns0SubnetOPT(srcIP, utils.DEFAULT_SOURCEMASK, utils.DEFAULT_SOURCESCOPE)
	} else {
		o = nil
	}
	return o, nil
}

//
func QueryNS(d string) ([]*dns.NS, *MyError.MyError) {
	//	ds, dp, _, e := preQuery(d, false)
	cf, _ := dns.ClientConfigFromFile("/etc/resolv.conf")
	e := &MyError.MyError{}
	r := &dns.Msg{}

	//	for c := 0; (c < 3) && cap(r.Answer) < 1; c++ {
	r, e = DoQuery(d, cf.Servers[0], cf.Port, dns.TypeNS, nil, UDP)
	if (e == nil) && (cap(r.Answer) > 0) {
		b, ns_a := ParseNS(r.Answer)
		if b != false {
			return ns_a, nil
		} else {
			return nil, MyError.NewError(MyError.ERROR_NORESULT, "ParseNS() has no result returned")
		}

		//		}
	}
	return nil, e
}

// Query
func QueryCNAME(d, srcIP, ds, dp string) ([]*dns.CNAME, *dns.RR_Header, *dns.EDNS0_SUBNET, *MyError.MyError) {
	o, e := preQuery(d, srcIP)
	r, e := DoQuery(d, ds, dp, dns.TypeCNAME, o, UDP)
	if e != nil {
		return nil, nil, nil, e
	}
	fmt.Println(utils.GetDebugLine(), r)
	fmt.Println(utils.GetDebugLine(), e)
	cname_a, ok := ParseCNAME(r.Answer, d)
	if ok != true {
		return nil, nil, nil, MyError.NewError(MyError.ERROR_NORESULT, "No CNAME record returned")
	}

	var edns_header *dns.RR_Header
	var edns *dns.EDNS0_SUBNET
	if len(srcIP) > 0 {
		if x := r.IsEdns0(); x != nil {
			edns_header, edns = parseEdns0subnet(x)
		}
	}
	return cname_a, edns_header, edns, nil
}

func parseEdns0subnet(edns_opt *dns.OPT) (*dns.RR_Header, *dns.EDNS0_SUBNET) {

	if edns_opt == nil {
		return nil, nil
	}
	edns_header, edns := UnpackEdns0Subnet(edns_opt)
	//	if cap(edns) == 0 {
	//		//TODO: if edns_array is empty,that mean:
	//		//	1. server does not support edns0_subnet
	//		//  2. this domain does not configure as edns0_subnet support
	//		//  3. ADD function default edns0_subnet_fill() ??
	//		edns = nil
	//	}
	return edns_header, edns
}

func QueryA(d string, srcIp, ds, dp string) ([]dns.RR, *dns.RR_Header, *dns.EDNS0_SUBNET, *MyError.MyError) {
	o, e := preQuery(d, srcIp)
	r, e := DoQuery(d, ds, dp, dns.TypeA, o, UDP)
	if e != nil || r == nil {
		//		fmt.Println(r)
		return nil, nil, nil, e
	}
	var edns_header *dns.RR_Header
	var edns *dns.EDNS0_SUBNET
	if len(srcIp) > 0 {
		if x := r.IsEdns0(); x != nil {
			edns_header, edns = parseEdns0subnet(x)
		}
	}
	return r.Answer, edns_header, edns, nil
}

func PackEdns0SubnetOPT(ip string, sourceNetmask, sourceScope uint8) *dns.OPT {
	edns0subnet := &dns.EDNS0_SUBNET{
		Code:          dns.EDNS0SUBNET,
		SourceScope:   sourceScope,
		SourceNetmask: sourceNetmask,
		Address:       net.ParseIP(ip).To4(),
		Family:        1,
	}
	o := &dns.OPT{
		Hdr: dns.RR_Header{
			Name:   ".",
			Rrtype: dns.TypeOPT,
		},
	}
	o.Option = append(o.Option, edns0subnet)
	return o
}

// func UnpackEdns0Subnet() Unpack Edns0Subnet OPT to []*dns.EDNS0_SUBNET
// if is not dns.EDNS0_SUBNET type ,return nil .
//
///*
//	fmt.Println("-------isEdns0--------")
//	if x := r.IsEdns0(); x != nil {
//		re := UnpackEdns0Subnet(x)
//		fmt.Println(x.Hdr.Name)
//		fmt.Println(x.Hdr.Class)
//		fmt.Println(x.Hdr.Rdlength)
//		fmt.Println(x.Hdr.Rrtype)
//		fmt.Println(x.Hdr.Ttl)
//		//		fmt.Println(x.Hdr.Header())
//		//		fmt.Println(x.Hdr.String())
//		fmt.Println("xxxxxxxxxx")
//		for _, v := range re {
//			fmt.Println(v.Address)
//			fmt.Println(v.SourceNetmask)
//			fmt.Println(v.SourceScope)
//			fmt.Println(v.Code)
//			fmt.Println(v.Family)
//			if on := v.Option(); on == dns.EDNS0SUBNET || on == dns.EDNS0SUBNETDRAFT {
//				fmt.Println("sure of ends0subnet")
//			} else {
//				fmt.Println("not sure")
//			}
//		}
//		fmt.Println(x.Version())
//	} else {
//		fmt.Println("no edns0")
//	}
// */
//TODO: edns_h and edns (*dns.EDNS0_SUBNET) can be combined into a struct
func UnpackEdns0Subnet(opt *dns.OPT) (*dns.RR_Header, *dns.EDNS0_SUBNET) {
	var re *dns.EDNS0_SUBNET = nil
	if cap(opt.Option) > 0 {
		for _, v := range opt.Option {
			if vo := v.Option(); vo == dns.EDNS0SUBNET || vo == dns.EDNS0SUBNETDRAFT {
				if oo, ok := v.(*dns.EDNS0_SUBNET); ok {
					re = oo
				}
			}
		}
		//TODO: consider Do not return opt.Hdr ??
		return &opt.Hdr, re
	}
	return nil, nil
}

//@TODO: randow cf.Servers for load banlance
func GetDomainResolver(domainName string) (string, string, *MyError.MyError) {
	if _, ok := dns.IsDomainName(domainName); !ok {
		return "", "", MyError.NewError(MyError.ERROR_PARAM, "Parma is not a domain name : "+domainName)
	}
	domainResolverIP, domainResolverPort, e := GetDomainConfigFromDomainTree(domainName)
	if (e != nil) && (e.ErrorNo == MyError.ERROR_NORESULT) {
		//TODO: 后台获取 domainName 的 Resolver 信息,并存储
		//QueryNS()
		domainResolverIP = "114.114.114.114"
		domainResolverPort = "53"

	} else {
		//		return "", "", e
	}
	return domainResolverIP, domainResolverPort, nil
}

func GetDomainConfigFromDomainTree(domain string) (string, string, *MyError.MyError) {
	var ds, dp string
	dp = "53"
	domain = dns.Fqdn(domain)
	switch domain {
	case "www.baidu.com.":
		ds = "ns2.baidu.com."
	case "www.a.shifen.com.":
		ds = "ns1.a.shifen.com."
	case "a.shifen.com.":
		ds = "ns1.a.shifen.com."
	case "ww2.sinaimg.cn.":
		ds = "ns1.sina.com.cn."
	case "weiboimg.gslb.sinaedge.com.":
		ds = "ns2.sinaedge.com."
	case "weiboimg.grid.sinaedge.com.":
		ds = "ns1.sinaedge.com."
	case "api.weibo.cn.":
		ds = "ns1.sina.com.cn."
	case "img.alicdn.com.":
		ds = "ns8.alibabaonline.com."
	case "img.alicdn.com.danuoyi.alicdn.com.":
		ds = "danuoyinewns1.gds.alicdn.com."
	default:
		if _, ok := dns.IsDomainName(domain); ok {
			ds = "8.8.8.8"
		}
	}
	return ds, dp, nil
}
