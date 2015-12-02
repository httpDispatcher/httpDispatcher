package query

import "sync"

type RegionNode struct {
	IpStart uint64
	IpEnd   uint64
	Rr      []RR
}

type ReginonTree struct {
	Region RegionNode
	mu     sync.Mutex
}
