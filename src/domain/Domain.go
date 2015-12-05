package domain

import (
	"MyError"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/miekg/dns"
	"github.com/petar/GoLLRB/llrb"
)

type MuLLRB struct {
	llrb.LLRB
	sync.Mutex
}

//TODO: redundant data types, need to be redesign
type Domain struct {
	DomainName string
	NS         []*dns.NS
	TTL        uint32
}

type DomainNode struct {
	Domain
	RegionRRTree *MuLLRB
}

//TODO: redundant data types, need to be redesign
// dns.RR && RrType && TTL
type Region struct {
	IpStart    uint32
	IpEnd      uint32
	RR         []*dns.RR
	RrType     dns.Type
	TTL        uint32
	UpdateTime time.Time
}

type DomainConfig struct {
	DomainName           string   `domainName`
	AuthoritativeServers []string `asServers`
	Port                 string
	Ttl                  string
}

type DomainTree MuLLRB
type RegionTree MuLLRB

var once sync.Once
var DomainDB *DomainTree

func init() {
	db := InitDomainDB()
	if db == nil {
		fmt.Println("InitDomainDB failed!")
		os.Exit(2)
	} else {
		fmt.Println("InitDomainDB succ!")
	}
}

func (a *DomainNode) Less(b llrb.Item) bool {
	if x, ok := b.(*DomainNode); ok {
		return a.DomainName < x.DomainName
	} else if y, ok := b.(*Domain); ok {
		return a.DomainName < y.DomainName
	}
	return false
}
func (a *Domain) Less(b llrb.Item) bool {
	if x, ok := b.(*DomainNode); ok {
		return a.DomainName < x.DomainName
	} else if y, ok := b.(*Domain); ok {
		return a.DomainName < y.DomainName
	}
	return false
}

func InitDomainDB() *MuLLRB {
	once.Do(func() {
		DomainDB = &MuLLRB{}
		return DomainDB
	})
	return nil
}

func (DT *DomainTree) StoreDomain(d *Domain) {

}

func (DT *DomainTree) SearchDomain(d *Domain) {

}

func (DT *DomainTree) UpdateDomain(d *Domain) {

}

func InitDomainRegionTree(d Domain) *RegionTree {

}

func (RT *RegionTree) defaultRegion() {

}

func (RT *RegionTree) SearRegion(d Domain, s, e uint32) {

}

func (RT *RegionTree) AddRegion(d *Domain, r *Region) {

}

func (RT *RegionTree) UpdateRegion(d *Domain, r *Region) {

}

func (RT *RegionTree) DelRegion(d *Domain, r *Region) {

}

func NewDomain(d string, ns []*dns.NS, t uint32) (*Domain, *MyError.MyError) {
	if _, ok := dns.IsDomainName(d); !ok {
		return MyError.NewError(MyError.ERROR_PARAM, d+" is not valid domain name")
	}
	//	if ns != nil {
	//		for _,n := range ns {
	//			fmt.Println(n)
	//		}
	//	}
	return &Domain{
		DomainName: dns.Fqdn(d),
		NS:         ns,
		TTL:        t,
	}
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
