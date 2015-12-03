package domain

import (
	"MyError"
	"errors"
	"github.com/miekg/dns"
	"github.com/yasushi-saito/rbtree"
	"net"
	"strconv"
	"strings"
	"sync"
)

type Domain struct {
	DomainName   string
	Ns           []string
	Ttl          uint32
	RegionRRTree *rbtree.Tree
}

type DomainConfig struct {
	DomainName           string   `domainName`
	AuthoritativeServers []string `asServers`
	Port                 string
	Ttl                  string
}

var once sync.Once

// type DomainTree rbtree.Tree

var DomainDB *rbtree.Tree

func NewDomain(d string, ns []string, t uint32) (*Domain, error) {
	if ns == nil {
		ns = []string{""}
	}

	if (len(d) > 0) && (len(ns) > 0) {
		return &Domain{DomainName: d,
			Ns:  ns,
			Ttl: t,
		}, nil
	} else {
		return nil, errors.New("param error ! " + d + ":" + strings.Join(ns, ",") + ":" + strconv.Itoa(int(t)))
	}
}

func InitDomainDB() *rbtree.Tree {
	once.Do(func() {
		DomainDB = rbtree.NewTree(func(a, b rbtree.Item) int {
			return strings.Compare(a.(*Domain).DomainName, b.(*Domain).DomainName)
		})

	})
	return DomainDB
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

func init() {
	domainDB := InitDomainDB()
	bd_ns := []string{"ns1.baidu.com", "ns2.baidu.com"}
	dt, _ := NewDomain("www.baidu.com", bd_ns, 86400)
	domainDB.Insert(dt)
	sina_ns := []string{"ns1.sina.com", "ns2.sina.com"}
	dt, _ = NewDomain("api.weibo.cn", sina_ns, 86400)
	domainDB.Insert(dt)
	dt, _ = NewDomain("api.weibo.cn", sina_ns, 86400)
	domainDB.Insert(dt)
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

func (d *DomainConfig) SetTtl(t string) error {
	if ti, e := strconv.Atoi(t); e == nil {
		if ti > 0 && ti < 65535 {
			d.Ttl = t
		} else {
			return errors.New(t + "is not permited")
		}
	}
	return nil
}

func (d *DomainConfig) SetAS(as []string) error {
	if len(d.AuthoritativeServers) != 0 {
		d.AuthoritativeServers = nil
	}
	for _, s := range as {
		if ip := net.ParseIP(s); ip != nil {
			d.AuthoritativeServers = append(d.AuthoritativeServers, s)
		} else {
			return errors.New(s + " is not standard ip string")
		}
	}
	return nil
}

func CompareDomain(a, b *DomainConfig) int {
	return strings.Compare(a.DomainName, b.DomainName)
}
