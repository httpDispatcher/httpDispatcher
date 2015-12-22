package server

import (
	"MyError"
	"domain"
	"fmt"
	"query"
)

func GetSOARecord(d string) (*domain.DomainSOANode, *MyError.MyError) {

	var soa *domain.DomainSOANode

	dn, e := domain.DomainRRCache.GetDomainNodeFromCacheWithName(d)
	if e == nil && dn != nil {
		dsoa_key := dn.SOAKey
		soa, e = domain.DomainSOACache.GetDomainSOANodeFromCacheWithDomainName(dsoa_key)
		fmt.Println("GOOOOOOOOOOOOT!!!")
		fmt.Println(soa)
		if e == nil && soa != nil {
			return soa, nil
		} else {
			// error == nil bug soa record also == nil
		}
	}
	//else if e != nil && e.ErrorNo == MyError.ERROR_NOTFOUND{
	soa_t, ns, e := query.QuerySOA(d)
	// Need to store DomainSOANode and DomainNOde both
	if e == nil && soa_t != nil && ns != nil {
		soa = &domain.DomainSOANode{
			SOAKey: soa_t.Hdr.Name,
			SOA:    soa_t,
			NS:     ns,
		}
		//TODO: get StoreDomainSOANode return values
		domain.DomainSOACache.StoreDomainSOANodeToCache(soa)
		rrnode, _ := domain.NewDomainNode(d, soa.SOAKey, soa_t.Expire)
		domain.DomainRRCache.StoreDomainNodeToCache(rrnode)
		return soa, nil
	}
	//	}
	// QuerySOA fail
	return nil, MyError.NewError(MyError.ERROR_UNKNOWN, "Finally GetSOARecord failed")
}

func GetARecord(d string, srcIP string) {
	//	//		var ar *dns.A
	//	var dt *domain.RegionTree
	//	//First, find RetionTree
	//	dn, e := domain.DomainRRDB.GetDomainNodeWithName(d)
	//	if e == nil && dn != nil {
	//		dt = dn.DomainRegionTree
	//	} else {
	//		//RegionNode and RegionTree is nil
	//		//Need add this node
	//		soa,e := GetSOARecord(d)
	//		if e==nil && soa != nil{
	//
	//		}else {
	//
	//		}
	//	}

}
