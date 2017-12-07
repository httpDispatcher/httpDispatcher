package query

import (
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"

	"MyError"
	"utils"
)

const (
	DOMAIN_MAX_LABEL    = 16
	NS_SERVER_PORT      = "53"
	DEFAULT_RESOLV_FILE = "/etc/resolv.conf"
	UDP                 = "udp"
	TCP                 = "tcp"
	DEFAULT_SOURCEMASK  = 32
	DEFAULT_SOURCESCOPE = 0
)

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
		Msg:           DnsMsgPool.Get().(*dns.Msg),
	}
}

var ClientPool = &sync.Pool{
	New: func() interface{} {
		return &dns.Client{
			DialTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
			ReadTimeout:  9 * time.Second,
		}
	},
}

var DnsMsgPool = &sync.Pool{
	New: func() interface{} {
		return &dns.Msg{}
	},
}

func RenewDnsClient(c *dns.Client) {
	c.Net = UDP
}

func RenewDnsMsg(m *dns.Msg) {
	m.Extra = nil
	m.Answer = nil
	m.AuthenticatedData = false
	m.CheckingDisabled = false
	m.Question = nil
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

func doQuery(c dns.Client, m dns.Msg, ds, dp string, queryType uint16, close chan struct{}) *dns.Msg {
	//	r := &dns.Msg{}
	//	var ee error
	//fmt.Println(utils.GetDebugLine(), " doQuery: ", " m.Question: ", m.Question,
	//	" ds: ", ds, " dp: ", dp, " queryType ", queryType)
	utils.ServerLogger.Debug(" doQuery: m.Question: %v ds: %s dp: %s queryType: %v", m.Question, ds, dp, queryType)
	select {
	case <-close:
		return nil
	default:
		for l := 0; l < 3; l++ {
			r, _, ee := c.Exchange(&m, ds+":"+dp)
			if (ee != nil) || (r == nil) || (r.Answer == nil) {
				utils.ServerLogger.Error(" doQuery: retry: %s times error: %s", strconv.Itoa(l), ee.Error())
				//if (queryType == dns.TypeA) || (queryType == dns.TypeCNAME) {
				if strings.Contains(ee.Error(), "connection refused") {
					if c.Net == TCP {
						c.Net = UDP
					}
				} else if (ee == dns.ErrTruncated) && queryType == dns.TypeA {
					utils.ServerLogger.Error(" doQuery: response truncated: %v", r)
					//					m.SetEdns0(4096,false)
					//					m.SetQuestion(dns.Fqdn(domainName),dns.TypeCNAME)
					c.Net = TCP
				} else {
					if c.Net == TCP {
						c.Net = UDP
					} else {
						c.Net = TCP
					}
				}
				//}
			} else {
				return r
			}
		}
	}

	return nil
}

// General Query for dns upstream query
// param: t string ["tcp"|"udp]
// 		  queryType uint16 dns.QueryType
func DoQuery(
	domainName string,
	domainResolverIP []string,
	domainResolverPort string,
	queryType uint16,
	queryOpt *dns.OPT, t string) (*dns.Msg, *MyError.MyError) {

	c := ClientPool.Get().(*dns.Client)
	defer func(c *dns.Client) {
		go func() {
			RenewDnsClient(c)
			ClientPool.Put(c)
		}()
	}(c)
	c.Net = t

	//	m := &dns.Msg{}
	m := DnsMsgPool.Get().(*dns.Msg)
	defer func(m *dns.Msg) {
		go func() {
			RenewDnsMsg(m)
			DnsMsgPool.Put(m)
		}()
	}(m)
	m.AuthenticatedData = true
	m.RecursionDesired = true
	//	m.Truncated= false
	m.SetQuestion(dns.Fqdn(domainName), queryType)

	if queryOpt != nil {
		m.Extra = append(m.Extra, queryOpt)
	}
	var x = make(chan *dns.Msg)
	var closesig = make(chan struct{})
	for _, ds := range domainResolverIP {
		go func(c *dns.Client, m *dns.Msg, ds, dp string, queryType uint16, closesig chan struct{}) {
			select {
			case x <- doQuery(*c, *m, ds, domainResolverPort, queryType, closesig):
			default:
			}
		}(c, m, ds, domainResolverPort, queryType, closesig)
	}

	if r := <-x; r != nil {
		close(closesig)
		return r, nil
	} else {
		return nil, MyError.NewError(MyError.ERROR_UNKNOWN, "Query failed "+domainName)
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
	if _, ok := dns.IsDomainName(d); !ok {
		//		return "", "", nil, MyError.NewError(MyError.ERROR_PARAM, d+" is not a domain name!")
	}

	var o *dns.OPT
	if len(srcIP) > 0 {
		o = PackEdns0SubnetOPT(srcIP, DEFAULT_SOURCEMASK, DEFAULT_SOURCESCOPE)
	} else {
		o = nil
	}
	return o, nil
}

func QuerySOA(d string) (*dns.SOA, []*dns.NS, *MyError.MyError) {
	//fmt.Println(utils.GetDebugLine(), " QuerySOA: ", d)
	utils.ServerLogger.Debug(" QuerySOA domain: %s ", d)
	if _, ok := dns.IsDomainName(d); !ok {
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
		r, e := DoQuery(d, cf.Servers, cf.Port, dns.TypeSOA, nil, UDP)
		//		fmt.Println(r)
		if e != nil {
			utils.QueryLogger.Error("QeurySOA got error : "+e.Error()+
				". Param: %s , %v, %s, %v ", d, cf.Servers, cf.Port, dns.TypeSOA)
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
					utils.ServerLogger.Error("ERROR_NOTVALID: %s", e.Error())
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
					//fmt.Println(utils.GetDebugLine(), e)
					utils.ServerLogger.Error("ERROR_NORESULT: %s", e.Error())
					//fmt.Println("+++++++++++++++++++++++++++++++++++")
					continue
				default:
					//					c++
					utils.ServerLogger.Error("ERROR_DEFAULT: %s", e.Error())
					//fmt.Println(utils.GetDebugLine(), ".....................")
					//fmt.Println(utils.GetDebugLine(), e)
					continue
					//					return nil, nil, e
				}
			} else {
				if cap(ns_a) < 1 {
					//fmt.Println(utils.GetDebugLine(), "QuerySOA: line 223: cap(ns_a)<1, need QueryNS ", soa.Hdr.Name)
					utils.ServerLogger.Debug("QuerySOA: cap(ns_a)<1, need QueryNS: %s", soa.Hdr.Name)
					ns_a, e = QueryNS(soa.Hdr.Name)
					if e != nil {
						//TODO: do some log
					}
				}
				//				fmt.Println("============xxxxxx================")
				//fmt.Println(utils.GetDebugLine(), "QuerySOA: soa record ", soa, " ns_a: ", ns_a)
				utils.ServerLogger.Debug("QuerySOA: soa record %v ns_a: %v", soa, ns_a)
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
					//fmt.Print(utils.GetDebugLine(), "ParseSOA:  ", vv)
					soa = vv
					utils.ServerLogger.Debug("ParseSOA: %v", vv)
				}
			case dns.TypeNS:
				if vv, ok := v.(*dns.NS); ok {
					ns_a = append(ns_a, vv)
				}
			default:
				//fmt.Println(utils.GetDebugLine(), " PasreSOA: error unexpect: ", v)
				utils.ServerLogger.Error("ParseSOA: error unexpect %v", v)
			}
		} else {
			//fmt.Print(utils.GetDebugLine(), "ParseSOA 258 ")
			//fmt.Println(utils.GetDebugLine(), vh.Name+" not match "+d)
			utils.ServerLogger.Debug("%s not match %s", vh.Name, d)
			return nil, nil, MyError.NewError(MyError.ERROR_NOTVALID, d+" has no SOA record,try parent")
		}

	}
	if soa != nil {
		return soa, ns_a, nil
	} else {
		return nil, nil, MyError.NewError(MyError.ERROR_NORESULT, "No SOA record for domain "+d)
	}
}

//
func QueryNS(d string) ([]*dns.NS, *MyError.MyError) {
	//	ds, dp, _, e := preQuery(d, false)
	cf, _ := dns.ClientConfigFromFile("/etc/resolv.conf")
	e := &MyError.MyError{}
	r := &dns.Msg{}

	//	for c := 0; (c < 3) && cap(r.Answer) < 1; c++ {
	r, e = DoQuery(d, cf.Servers, cf.Port, dns.TypeNS, nil, "udp")
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

func QueryA(d, srcIp string, ds []string, dp string) ([]dns.RR, *dns.RR_Header, *dns.EDNS0_SUBNET, *MyError.MyError) {
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

func ParseA(a []dns.RR, d string) ([]*dns.A, bool) {
	var a_rr []*dns.A
	for _, aa := range a {
		if x, ok := aa.(*dns.A); ok {
			if x.Hdr.Name == dns.Fqdn(d) {
				a_rr = append(a_rr, x)
			} else {
				//fmt.Println(utils.GetDebugLine(), "ParseA: line 135: ", x)
				utils.ServerLogger.Debug("ParseA: %s", x)
			}
		}
	}
	if cap(a_rr) > 0 {
		return a_rr, true
	}
	return nil, false
}

// Query
func QueryCNAME(d, srcIP string, ds []string, dp string) ([]*dns.CNAME, *dns.RR_Header, *dns.EDNS0_SUBNET, *MyError.MyError) {
	o, e := preQuery(d, srcIP)
	r, e := DoQuery(d, ds, dp, dns.TypeCNAME, o, UDP)
	if e != nil {
		return nil, nil, nil, e
	}
	//fmt.Println(utils.GetDebugLine(), r)
	//fmt.Println(utils.GetDebugLine(), e)
	cname_a, ok := ParseCNAME(r.Answer, d)
	if ok != true {
		return nil, nil, nil, MyError.NewError(MyError.ERROR_NORESULT, "No CNAME record returned")
	}

	utils.ServerLogger.Debug("QueryCNAME domain: %s srcip: %s result: %v", d, srcIP, cname_a)

	var edns_header *dns.RR_Header
	var edns *dns.EDNS0_SUBNET
	if len(srcIP) > 0 {
		if x := r.IsEdns0(); x != nil {
			edns_header, edns = parseEdns0subnet(x)
		}
	}
	return cname_a, edns_header, edns, nil
}

// Filter CNAME record in dns.Msg.Answer message.
// 	if r(*dns.Msg.Answer) includes (*dns.CNAME) , than return the CNAME record array.
func ParseCNAME(c []dns.RR, d string) ([]*dns.CNAME, bool) {
	//fmt.Println(utils.GetDebugLine(), "ParseCNAME line 99: ", c)
	utils.ServerLogger.Debug("ParseCNAME: %s", c)
	var cname_a []*dns.CNAME
	for _, a := range c {
		if cname, ok := a.(*dns.CNAME); ok {
			//fmt.Println(utils.GetDebugLine(), "ParseCNAME: line 103 : ", cname.Hdr.Name, d)
			utils.ServerLogger.Debug("ParseCNAME: %s  %s", cname.Hdr.Name, d)
			if cname.Hdr.Name == dns.Fqdn(d) {
				cname_a = append(cname_a, cname)
			} else {
				utils.ServerLogger.Debug("ParseCNAME: %s", cname)
			}
		}
	}
	if len(cname_a) > 0 {
		return cname_a, true
	}
	return nil, false
}
