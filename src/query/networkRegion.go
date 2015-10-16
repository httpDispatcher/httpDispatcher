package query

import "github.com/yasushi-saito/rbtree"

type NetworkRegion struct {
	Records *DomainRR
	IpStart uint64
	IpEnd   uint64
	NetMask uint8
}

type ReginonTree *rbtree.Tree

func NewNetworkRegion(rr *DomainRR, s, e uint64, n uint8) *NetworkRegion {
	return &NetworkRegion{Records: rr, IpStart: s, IpEnd: e, NetMask: n}

}
