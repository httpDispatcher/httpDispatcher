package query

import (
	"errors"
	"github.com/miekg/dns"
	"github.com/yasushi-saito/rbtree"
	"net"
	"strconv"
	"strings"
)

type Domain struct {
	DomainName           string   `domainName`
	AuthoritativeServers []string `asServers`
	Port                 string
	Ttl                  string
}

var DomainTree *rbtree.Tree

func NewDomain(D string, AS []string, P, T string) *Domain {
	return &Domain{DomainName: D,
		AuthoritativeServers: AS,
		Port:                 P,
		Ttl:                  T,
	}
}

func (dc *Domain) SetDomain(d string) (bool, error) {
	if _, ok := dns.IsDomainName(d); ok {
		dc.DomainName = dns.Fqdn(d)
		return true, nil
	} else {
		return false, errors.New(d + " is not domain name")
	}
}

func (dc *Domain) SetTtl(t string) error {
	if ti, e := strconv.Atoi(t); e == nil {
		if ti > 0 && ti < 1024 {
			dc.Ttl = t
		} else {
			return errors.New(t + "is not permited")
		}
	}
	return nil
}

func (dc *Domain) SetAS(as []string) error {
	if len(dc.AuthoritativeServers) != 0 {
		dc.AuthoritativeServers = nil
	}
	for _, s := range as {
		if ip := net.ParseIP(s); ip != nil {
			dc.AuthoritativeServers = append(dc.AuthoritativeServers, s)
		} else {
			return errors.New(s + " is not standard ip string")
		}
	}
	return nil
}

func CompareDomain(a, b *Domain) int {
	return strings.Compare(a.DomainName, b.DomainName)
}

func GetDomainTree() *rbtree.Tree {
	if DomainTree == nil {
		DomainTree = rbtree.NewTree(func(a, b rbtree.Item) int {
			return strings.Compare(a.(Domain).DomainName, b.(Domain).DomainName)
		})
	}
	return DomainTree

}
