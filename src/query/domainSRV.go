package query

import (
	"errors"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/miekg/dns"
	"github.com/yasushi-saito/rbtree"
)

type DomainConfig struct {
	DomainName           string   `domainName`
	AuthoritativeServers []string `asServers`
	Port                 string
	Ttl                  string
}

type Domain struct {
	name       string
	ns         DomainNS
	rrTree     *DomainRRTree
	regionRtee ReginonTree
	ttl        uint32
}

type DomainNS *dns.NS

var DomainTree *rbtree.Tree
var once sync.Once

func NewDomain(D string, AS []string, P, T string) *DomainConfig {
	return &DomainConfig{DomainName: D,
		AuthoritativeServers: AS,
		Port:                 P,
		Ttl:                  T,
	}
}

//@TODO: randow cf.Servers for load banlance
func GetDomainConfig(domainName string) (string, string, error) {
	if _, ok := dns.IsDomainName(domainName); !ok {
		return "", "", errors.New("Parma is not a domain name : " + domainName)
	}
	domainResolverIP, domainResolverPort, e := GetDomainConfigFromDomainTree(domainName)
	if e != nil {
		return "", "", e
	}
	return domainResolverIP, domainResolverPort, nil
}

func GetDomainConfigFromDomainTree(domain string) (string, string, error) {
	var ds, dp string
	dp = "53"
	domain = dns.Fqdn(domain)
	switch domain {
	case "www.baidu.com.":
		ds = "ns2.baidu.com."
	case "www.a.shifen.com.":
		ds = "ns1.a.shifen.com."
	case "ww2.sinaimg.cn.":
		ds = "ns1.sina.com.cn."
		//	case "weiboimg.gslb.sinaedge.com.":
		//		ds = "ns1.sinaedge.com."
		//	case "weiboimg.grid.sinaedge.com.":
		//		ds = "ns1.sinaedge.com."
	case "api.weibo.cn.":
		ds = "ns1.sina.com.cn."

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

func GetDomainTree() *rbtree.Tree {
	//	if DomainTree == nil {
	//		DomainTree = rbtree.NewTree(func(a, b rbtree.Item) int {
	//			return strings.Compare(a.(Domain).DomainName, b.(Domain).DomainName)
	//		})
	//	}
	once.Do(func() {
		DomainTree = rbtree.NewTree(func(a, b rbtree.Item) int {
			return strings.Compare(a.(DomainConfig).DomainName, b.(DomainConfig).DomainName)
		})
	})
	return DomainTree
}
