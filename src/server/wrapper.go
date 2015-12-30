package server

import (
	"MyError"
	"domain"
	"fmt"
	"net"
	"os"
	"query"
	"reflect"
	"time"
	"utils"

	"github.com/miekg/dns"
)

func GetClientIP() string {
	return "124.207.129.171"
}

func GetSOARecord(d string) (*domain.DomainSOANode, *MyError.MyError) {

	var soa *domain.DomainSOANode

	dn, e := domain.DomainRRCache.GetDomainNodeFromCacheWithName(d)
	if e == nil && dn != nil {
		dsoa_key := dn.SOAKey
		soa, e = domain.DomainSOACache.GetDomainSOANodeFromCacheWithDomainName(dsoa_key)
		fmt.Println(utils.GetDebugLine(), "GetSOARecord: line 28 : GOOOOOOOOOOOOT!!!", soa)
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
		go domain.DomainSOACache.StoreDomainSOANodeToCache(soa)
		rrnode, _ := domain.NewDomainNode(d, soa.SOAKey, soa_t.Expire)
		go domain.DomainRRCache.StoreDomainNodeToCache(rrnode)
		return soa, nil
	}
	//	}
	// QuerySOA fail
	return nil, MyError.NewError(MyError.ERROR_UNKNOWN, "Finally GetSOARecord failed")
}

func GetARecord(d string, srcIP string) (bool, []dns.RR, *MyError.MyError) {
	var Regiontree *domain.RegionTree
	var bigloopflag bool = false
	var A []dns.RR
	var c = 0
	for dst := d; (bigloopflag == false) && (c < 15); c++ {
		fmt.Println(utils.GetDebugLine(), "GetARecord : ", dst)

		dn, e := domain.DomainRRCache.GetDomainNodeFromCacheWithName(dst)
		fmt.Println(utils.GetDebugLine(), "GetARecord:", dn, e)
		if e == nil && dn != nil && dn.DomainRegionTree != nil {
			//Get DomainNode succ,
			r, e := dn.DomainRegionTree.GetRegionFromCacheWithAddr(utils.Ip4ToInt32(net.ParseIP(srcIP)), 32)
			fmt.Println(utils.GetDebugLine(), "GetARecord : ", r, e)
			if e == nil {
				fmt.Println(utils.GetDebugLine(), "GetArecord: Gooooot: ", r)
				if r.RrType == dns.TypeA {
					fmt.Println(utils.GetDebugLine(), "GetARecord: Goooot A", r.RR)
					bigloopflag = true
					//					os.Exit(2)
					A = r.RR
					return true, A, nil
					continue
				} else if r.RrType == dns.TypeCNAME {
					fmt.Println(utils.GetDebugLine(), "GetARecord : Goooot CNAME", r.RR)
					dst = r.RR[0].(*dns.CNAME).Target
					continue
				}
			} else if e.ErrorNo == MyError.ERROR_NOTFOUND {
				fmt.Println(utils.GetDebugLine(), "Not found r, need query dns")
			}
			// return
		} else if e == nil && dn != nil && dn.DomainRegionTree == nil {
			// if RegionTree is nil, init RegionTree First
			ok, e := dn.InitRegionTree()
			//
			fmt.Println("RegionTree is nil ,Init it: "+reflect.ValueOf(ok).String(), e)
		} else {
			// e != nil
			// RegionTree is not nil
			fmt.Print(utils.GetDebugLine(), "GetARecord:")
			fmt.Println(dn, e)
			if e.ErrorNo != MyError.ERROR_NOTFOUND {
				fmt.Println("Found unexpected error, need return !")
				os.Exit(2)
			}
		}

		soa, e := GetSOARecord(dst)
		fmt.Println(utils.GetDebugLine(), "GetARecord:", soa, e)
		if e != nil {
			// error! need return
			fmt.Print(utils.GetDebugLine(), "GetARecord: ")
			fmt.Println(e)
			fmt.Println("error111,need return")
		}
		var cacheflag bool = false
		var cc = 0
		for cacheflag = false; (cacheflag != true) && (cc < 3); cc++ {
			// wait for goroutine 'StoreDomainNodeToCache' in GetSOARecord to be finished
			dn, e = domain.DomainRRCache.GetDomainNodeFromCacheWithName(dst)
			if e != nil {
				// here ,may be nil
				// error! need return
				fmt.Println(utils.GetDebugLine(), "GetARecord : error222,need waite", e)
				time.Sleep(1 * time.Second)
			} else {
				cacheflag = true
				fmt.Println(utils.GetDebugLine(), "GetARecord: ", dn)
			}
		}
		if e != nil || len(soa.NS) <= 0 {
			//GetSOA failed , need log and return
			return false, nil, MyError.NewError(MyError.ERROR_UNKNOWN, "GetARecord func GetSOARecord failed: "+d)
		}
		dn.InitRegionTree()
		Regiontree = dn.DomainRegionTree

		//		soa.NS[0].Ns = "8.8.8.8"
		fmt.Println(utils.GetDebugLine(), dst, srcIP, soa.NS)
		var ns_a []string
		for _, x := range soa.NS {
			ns_a = append(ns_a, x.Ns)
		}
		rr, edns_h, edns, e := query.QueryA(dst, srcIP, ns_a, "53")

		if e == nil && rr != nil {
			var rr_i []dns.RR
			if a, ok := query.ParseA(rr, dst); ok {
				//rr is A record
				fmt.Print(utils.GetDebugLine(), "GetARecord : typeA ")
				fmt.Println(utils.GetDebugLine(), a, ok)
				bigloopflag = true
				for _, i := range a {
					rr_i = append(rr_i, dns.RR(i))
				}
				A = rr_i
			} else if b, ok := query.ParseCNAME(rr, dst); ok {
				//rr is CNAME record
				fmt.Println(utils.GetDebugLine(), "GetARecord: typeCNAME ", b, ok)
				dst = b[0].Target
				for _, i := range b {
					rr_i = append(rr_i, dns.RR(i))
				}
			} else {
				fmt.Println(utils.GetDebugLine(), "GetARecord: ", rr)
				continue
			}
			fmt.Println(utils.GetDebugLine(), "GetARecord: ", edns_h, edns)
			var ipnet *net.IPNet
			if edns != nil {
				ipnet, e = utils.ParseEdnsIPNet(edns.Address, edns.SourceScope, edns.Family)
			}
			fmt.Println(utils.GetDebugLine(), "GetARecord: ", e)
			if ipnet != nil {
				netaddr, mask := utils.IpNetToInt32(ipnet)
				r, _ := domain.NewRegion(rr_i, netaddr, mask)
				Regiontree.AddRegionToCache(r)
				fmt.Print(utils.GetDebugLine(), "GetARecord: ")
				fmt.Println(Regiontree.GetRegionFromCacheWithAddr(netaddr, mask))

			} else {
				netaddr, mask := domain.DefaultNetaddr, domain.DefaultMask
				r, _ := domain.NewRegion(rr_i, netaddr, mask)
				Regiontree.AddRegionToCache(r)
				fmt.Println(utils.GetDebugLine(), "GetARecord: ", r)
				fmt.Println(Regiontree.GetRegionFromCacheWithAddr(netaddr, mask))
			}
		} else {
			//QueryA error!
			//			bigloopflag = true
			fmt.Println(utils.GetDebugLine(), e)
			return false, nil, e
		}
	}
	fmt.Println(utils.GetDebugLine(), "GetARecord: ", Regiontree)
	return true, A, nil
}

//func GeneralDNSBackendQuery(d string, srcIP string) ([]dns.RR, *net.IPNet, *MyError.MyError) {
//	_, ns, e := query.QuerySOA(d)
//	if e != nil || cap(ns) < 1 {
//		return nil, nil, e
//	}
//	var ns_a []string
//	for _, x := range ns {
//		ns_a = append(ns_a, x.Ns)
//	}
//	a_rr, _, edns, e := query.QueryA(d, srcIP, ns_a, "53")
//	if e != nil || cap(a_rr) < 1 {
//		return nil, nil, e
//	}
//
//	//	var ipnet *net.IPNet
//	if edns != nil {
//		_, ipnet, ee := net.ParseCIDR(strings.Join([]string{edns.Address.String(), strconv.Itoa(int(edns.SourceScope))}, "/"))
//		if ee == nil {
//			return a_rr, ipnet, nil
//		} else {
//			return a_rr, nil, MyError.NewError(MyError.ERROR_NOTVALID, ee.Error())
//		}
//	}
//	return a_rr, nil, e
//}
