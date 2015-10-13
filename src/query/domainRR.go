package query

import (
	"strings"

	"github.com/yasushi-saito/rbtree"
	//	"sync"
)

type DomainRR struct {
	rr     []string
	domain *DomainConfig
}

var DomainRRTree *rbtree.Tree

//var once sync.Once

func NewDomainRR(d *DomainConfig, r []string) *DomainRR {
	return &DomainRR{domain: d, rr: r}
}

func (dr *DomainRR) SetRR(r []string, is_append bool) {
	if is_append {
		dr.rr = append(dr.rr, r...)
	} else {
		dr.rr = r
	}
}

func (dr *DomainRR) GetRR() []string {
	return dr.rr
}

func (dr *DomainRR) GetDomain() *DomainConfig {
	return dr.domain
}

func InitDomainRRTree() *rbtree.Tree {
	//	if DomainRRTree == nil {
	//		DomainRRTree = rbtree.NewTree(func(a, b rbtree.Item) int {
	//			return strings.Compare(a.(DomainRR).domain.DomainName, b.(DomainRR).domain.DomainName)
	//		})
	//	}

	once.Do(func() {
		DomainRRTree = rbtree.NewTree(func(a, b rbtree.Item) int {
			return strings.Compare(a.(DomainRR).domain.DomainName, b.(DomainRR).domain.DomainName)
		})
	})
	return DomainRRTree
}
