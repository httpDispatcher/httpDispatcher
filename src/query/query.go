package query

import (
	"MyError"
	"domain"
	"fmt"
	"github.com/miekg/dns"
	"net"
	"strings"
	"time"
	"utils"
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

	fmt.Println("++++++++begin ds dp+++++++++")
	fmt.Println(domainResolverIP + "......" + domainResolverPort + "  .......")
	fmt.Println("++++++++end ds dp+++++++++")

	c := &dns.Client{
		DialTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		ReadTimeout:  60 * time.Second,
		Net:          t,
	}

	m := &dns.Msg{}
	m.AuthenticatedData = true
	m.RecursionDesired = true
	m.SetQuestion(dns.Fqdn(domainName), queryType)

	if queryOpt != nil {
		m.Extra = append(m.Extra, queryOpt)
	}
	r, _, ee := c.Exchange(m, domainResolverIP+":"+domainResolverPort)

	if ee != nil {
		fmt.Println(ee.Error())
		//		os.Exit(1)
		return nil, MyError.NewError(MyError.ERROR_UNKNOWN, ee.Error())
	}
	return r, nil
}

// Filter CNAME record in dns.Msg.Answer message.
// 	if r(*dns.Msg.Answer) includes (*dns.CNAME) , than return the CNAME record array.
func ParseCNAME(c []dns.RR) (bool, []*dns.CNAME) {
	var cname_a []*dns.CNAME
	for _, a := range c {
		if cname, ok := a.(*dns.CNAME); ok {
			//						fmt.Println("CNAME Found!: " + cname.Hdr.Name + " <--> " + cname.Target)
			cname_a = append(cname_a, cname)
		}
	}
	if cap(cname_a) > 0 {
		return true, cname_a
	}
	return false, nil
}

// Parse dns.Msg.Answer in dns response msg that use TypeNS as request type.
func ParseNS(ns []dns.RR) []*dns.NS {
	var ns_rr []*dns.NS
	if cap(ns) > 0 {
		for _, n_s := range ns {
			if x, ok := n_s.(*dns.NS); ok {
				ns_rr = append(ns_rr, x)
			}
		}
	}
	return ns_rr
}

// Loop For Query the domain name d's NS servers
// if d has no NS server, Loop for the parent domain's name server
func LoopForQueryNS(d string) ([]*dns.NS, *MyError.MyError) {
	if _, ok := Check_DomainName(d); !ok {
		return nil, MyError.NewError(MyError.ERROR_PARAM, d+" is not a domain name")
	}
	r := dns.SplitDomainName(d)
	fmt.Println(r)
	if cap(r) > DOMAIN_MAX_LABEL {
		return nil, MyError.NewError(MyError.ERROR_PARAM, d+" is too long")
	}
	for cap(r) > 2 {
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

// Preparation for Query A and CNAME / NS record.
// param d 		: domain name to query
// 		isEdns0 : either use edns0_subnet or not
// return :
//		*MyError.Myerror
//		domain name server ip
//		domain name server port
//		*dns.OPT (for edns0_subnet)
func preQuery(d string, isEdns0 bool) (*MyError.MyError, string, string, *dns.OPT) {
	if _, ok := Check_DomainName(d); !ok {
		return MyError.NewError(MyError.ERROR_PARAM, d+" is not a domain name!"), "", "", nil
	}
	ds, dp, e := domain.GetDomainResolver(d)
	if e != nil {
		return e, "", "", nil
	}

	var o *dns.OPT
	if isEdns0 {
		//TODO:modify GetClientIP func to use Server package
		ip := utils.GetClientIP()
		o = PackEdns0SubnetOPT(ip, utils.DEFAULT_SOURCEMASK, utils.DEFAULT_SOURCESCOPE)
	} else {
		o = nil
	}
	return nil, ds, dp, o
}

//
func QueryNS(d string) ([]*dns.NS, *MyError.MyError) {
	e, ds, dp, _ := preQuery(d, false)
	r, e := DoQuery(d, ds, dp, dns.TypeNS, nil, UDP)
	if (e == nil) && (cap(r.Answer) > 0) {
		ns_a := ParseNS(r.Answer)
		return ns_a, nil
	}
	return nil, e
}

// Query
func QueryCNAME(d string, isEdns0 bool) (*MyError.MyError, []*dns.CNAME, interface{}, []*dns.EDNS0_SUBNET) {
	e, ds, dp, o := preQuery(d, isEdns0)
	r, e := DoQuery(d, ds, dp, dns.TypeCNAME, o, UDP)
	if e != nil {
		return e, nil, nil, nil
	}
	ok, cname_a := ParseCNAME(r.Answer)
	if ok != true {
		return MyError.NewError(MyError.ERROR_NORESULT, "No CNAME record returned"), nil, nil, nil
	}

	var edns_header interface{}
	var edns_array []*dns.EDNS0_SUBNET
	if isEdns0 {
		if x := r.IsEdns0(); x != nil {
			edns_header, edns_array = parseEdns0subnet(x)
		}
	}
	return nil, cname_a, edns_header, edns_array
}

func parseEdns0subnet(edns_opt *dns.OPT) (interface{}, []*dns.EDNS0_SUBNET) {

	if edns_opt == nil {
		return nil, nil
	}
	edns_header, edns_array := UnpackEdns0Subnet(edns_opt)
	if cap(edns_array) == 0 {
		//TODO: if edns_array is empty,that mean:
		//	1. server does not support edns0_subnet
		//  2. this domain does not configure as edns0_subnet support
		//  3. ADD function default edns0_subnet_fill() ??
		edns_array = nil
	}
	return edns_header, edns_array
}

func QueryA(d string, isEdns0 bool) ([]dns.RR, *dns.EDNS0_SUBNET, error) {

	e, ds, dp, o := preQuery(d, isEdns0)

	r, e := DoQuery(d, ds, dp, dns.TypeA, o, UDP)

	if e != nil {
		fmt.Println(r)
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

//func ParseA(answer []dns.RR) ([]string, uint32, error) {
//	var a_rr []string
//	var ttl uint32 = 0
//	fmt.Print("dns.IsRRset(answer):")
//	fmt.Println(dns.IsRRset(answer))
//
//		fmt.Println(reflect.TypeOf(a))
//		switch a.Header().Rrtype {
//		case dns.TypeCNAME:
//			{
//				fmt.Print("dns.TypeCNAME :")
//				fmt.Println(a.String())
//			}
//		case dns.TypeA:
//			{
//
//				a_rr, ttl = UnpackAAnswer(a_rr, a)
//			}
//		}
//	return a_rr, ttl, nil
//
//}

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

// Unpack Edns0Subnet OPT to []*dns.EDNS0_SUBNET
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
func UnpackEdns0Subnet(opt *dns.OPT) (interface{}, []*dns.EDNS0_SUBNET) {
	var re []*dns.EDNS0_SUBNET = nil
	if cap(opt.Option) > 0 {
		for _, v := range opt.Option {
			if vo := v.Option(); vo == dns.EDNS0SUBNET || vo == dns.EDNS0SUBNETDRAFT {
				if oo := v.(*dns.EDNS0_SUBNET); oo != nil {
					re = append(re, oo)
				}
			}
		}
		//TODO: consider Do not return opt.Hdr ??
		return opt.Hdr, re
	}
	return nil, nil
}
