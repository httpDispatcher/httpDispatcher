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

	"config"
	"strconv"

	"github.com/miekg/dns"
	"iplookup"
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
		fmt.Println(utils.GetDebugLine(), " GOOOOOOOOOOOOT!!!", soa)
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
	var bigloopflag bool = false // big loop flag
	var c = 0                    //big loop count

	fmt.Println(utils.GetDebugLine(), "^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^")

	//Can't loop for CNAME chain than bigger than 10
	for dst := d; (bigloopflag == false) && (c < 10); c++ {
		fmt.Println(utils.GetDebugLine(), "GetARecord : ", dst)

		ok, dn, RR, e := GetAFromCache(dst, srcIP)
		fmt.Println(utils.GetDebugLine(), ok, dn, RR, e)
		if ok {
			// All is right and especilly RR is A record
			return ok, RR, nil
		} else {
			if (dn != nil) && (RR != nil) {
				dst = RR[0].(*dns.CNAME).Target
				continue
			} else if (dn != nil) && (RR == nil) {
				// hava domain node ,but region node is nil,need queryA

			} //else if dn == nil {
			//				// have no daomain node,need query SQA
			//			}
			//			goto GetFromSOA
		}

		fmt.Println(utils.GetDebugLine(), "++++++++++++++++++++++++++++++++++++++++++++++")
		if config.IsLocalMysqlBackend(dst) {
			fmt.Println(utils.GetDebugLine(), "**********************************************")
			//need pass dn to GetAFromMySQLBackend, to fill th dn.RegionTree node
			ok, RR, rtype, ee := GetAFromMySQLBackend(dst, GetClientIP())
			fmt.Println(utils.GetDebugLine(), ok, RR, ee)
			if !ok {
				fmt.Println(utils.GetDebugLine(), "GetAFromMySQL error : ", ee)
			} else if rtype == dns.TypeA {
				fmt.Println(utils.GetDebugLine(), "ReGet dst :", dst, RR)
				return true, RR, nil
			} else if rtype == dns.TypeCNAME {
				dst = RR[0].(*dns.CNAME).Target
				continue
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
		fmt.Println(utils.GetDebugLine(), "init Cache dn:", cacheflag, cc, (cacheflag == false) && (cc < 5))
		for cacheflag = false; (cacheflag == false) && (cc < 5); {
			// wait for goroutine 'StoreDomainNodeToCache' in GetSOARecord to be finished
			dn, e = domain.DomainRRCache.GetDomainNodeFromCacheWithName(dst)
			if e != nil {
				// here ,may be nil
				// error! need return
				fmt.Println(utils.GetDebugLine(), "GetARecord : error222,need waite", e)
				time.Sleep(1 * time.Second)
				if e.ErrorNo == MyError.ERROR_NOTFOUND {
					cc++
				} else {
					fmt.Println(utils.GetDebugLine(), e)
					os.Exit(3)
				}
			} else {
				cacheflag = true
				fmt.Println(utils.GetDebugLine(), "GetARecord: ", dn)
			}
		}
		if e != nil || len(soa.NS) <= 0 {
			//GetSOA failed , need log and return
			return false, nil, MyError.NewError(MyError.ERROR_UNKNOWN,
				"GetARecord func GetSOARecord failed: "+d)
		}
		dn.InitRegionTree()
		Regiontree = dn.DomainRegionTree

		fmt.Println(utils.GetDebugLine(), dst, srcIP, soa.NS)
		var ns_a []string
		//todo: is that soa.NS may nil ?
		for _, x := range soa.NS {
			ns_a = append(ns_a, x.Ns)
		}
		ok, rr_i, rtype, ee := GetAFromDNSBackend(dst, srcIP, ns_a, Regiontree)
		if ok && rtype == dns.TypeA {
			return true, rr_i, nil
		} else if ok && rtype == dns.TypeCNAME {
			dst = rr_i[0].(*dns.CNAME).Target
			continue
		} else if !ok && rr_i == nil && ee != nil && ee.ErrorNo == MyError.ERROR_NORESULT {
			continue
		} else {
			return false, nil, MyError.NewError(MyError.ERROR_UNKNOWN, "Unknown error")
		}
	}
	fmt.Println(utils.GetDebugLine(), "GetARecord: ", Regiontree)
	return false, nil, MyError.NewError(MyError.ERROR_UNKNOWN, "Unknown error")
}

func GetAFromCache(dst, srcIP string) (bool, *domain.DomainNode, []dns.RR, *MyError.MyError) {
	dn, e := domain.DomainRRCache.GetDomainNodeFromCacheWithName(dst)
	fmt.Println(utils.GetDebugLine(), "GetAFromCache:", dn, e)
	if e == nil && dn != nil && dn.DomainRegionTree != nil {
		//Get DomainNode succ,
		r, e := dn.DomainRegionTree.GetRegionFromCacheWithAddr(
			utils.Ip4ToInt32(net.ParseIP(srcIP)), domain.DefaultRedaxMask)
		fmt.Println(utils.GetDebugLine(), "GetAFromCache : ", r, e)
		if e == nil {
			fmt.Println(utils.GetDebugLine(), "GetAFromCache: Gooooot: ", r)
			if r.RrType == dns.TypeA {
				fmt.Println(utils.GetDebugLine(), "GetAFromCache: Goooot A", r.RR)
				return true, dn, r.RR, nil
			} else if r.RrType == dns.TypeCNAME {
				fmt.Println(utils.GetDebugLine(), "GetAFromCache : Goooot CNAME", r.RR)
				if len(r.RR) > 0 {
					//					dst = r.RR[0].(*dns.CNAME).Target
					return false, dn, r.RR, MyError.NewError(MyError.ERROR_NOTVALID,
						"Get CNAME From,Requery A for "+r.RR[0].(*dns.CNAME).Target)
				} else {
					fmt.Println(utils.GetDebugLine(), "Error Got RegionFromCacheWithAddr", r.RR)
					return false, dn, nil, MyError.NewError(MyError.ERROR_NORESULT, "Error Got RegionFromCacheWithAddr ")
				}
				//				continue
			}
		} else if e.ErrorNo == MyError.ERROR_NOTFOUND {
			fmt.Println(utils.GetDebugLine(), "Not found r, need query dns")
			return false, dn, nil, MyError.NewError(MyError.ERROR_NOTFOUND,
				"Not found R in cache, dst :"+dst+" srcIP "+srcIP)
		}
		// return
	} else if e == nil && dn != nil && dn.DomainRegionTree == nil {
		// Get domainNode in cache tree,but no RR in region tree,need query with NS
		// if RegionTree is nil, init RegionTree First
		ok, e := dn.InitRegionTree()
		//
		fmt.Println("RegionTree is nil ,Init it: "+reflect.ValueOf(ok).String(), e)
		return false, dn, nil, MyError.NewError(MyError.ERROR_NORESULT,
			"Get domainNode in cache tree,but no RR in region tree,need query with NS, dst : "+dst+" srcIP "+srcIP)
	} else {
		// e != nil
		// RegionTree is not nil
		fmt.Print(utils.GetDebugLine(), "GetAFromCache dst: "+dst+" srcIP: "+srcIP)
		fmt.Println(dn, e)
		if e.ErrorNo != MyError.ERROR_NOTFOUND {
			fmt.Println("Found unexpected error, need return !")
			os.Exit(2)
		} else {
			return false, nil, nil, e
		}

	}
	return false, nil, nil, MyError.NewError(MyError.ERROR_UNKNOWN, "Unknown error!")
}

func GetAFromMySQLBackend(dst, srcIP string) (bool, []dns.RR, uint16, *MyError.MyError) {
	domainId, e := query.RRMySQL.GetDomainIDFromMySQL(dst)
	if e != nil {
		//todo:
		fmt.Println(utils.GetDebugLine(), "Error, GetDomainIDFromMySQL:", e)
		return false, nil, uint16(0), e
	}
	region, ee := query.RRMySQL.GetRegionWithIPFromMySQL(utils.Ip4ToInt32(utils.StrToIP(srcIP)))
	if ee != nil {
		fmt.Println(utils.GetDebugLine(), "Error GetRegionWithIPFromMySQL:", ee)
		return false, nil, uint16(0), MyError.NewError(ee.ErrorNo, "GetRegionWithIPFromMySQL return "+e.Error())
	}
	RR, eee := query.RRMySQL.GetRRFromMySQL(uint32(domainId), region.IdRegion)
	if eee != nil {
		fmt.Println(utils.GetDebugLine(), "Error GetRRFromMySQL with DomainID:", domainId,
			"RegionID:", region.IdRegion, eee)
		fmt.Println(utils.GetDebugLine(), "Try to GetRRFromMySQL with Default Region")
		RR, eee = query.RRMySQL.GetRRFromMySQL(uint32(domainId), uint32(0))
		if eee != nil {
			fmt.Println(utils.GetDebugLine(), "Error GetRRFromMySQL with DomainID:", domainId,
				"RegionID:", 0, eee)
			return false, nil, uint16(0), MyError.NewError(eee.ErrorNo, "Error GetRRFromMySQL with DomainID:"+strconv.Itoa(domainId)+eee.Error())
		}
	}
	fmt.Println(utils.GetDebugLine(), "GetRRFromMySQL Succ!:", RR)
	if len(RR) > 0 {
		var R []dns.RR
		var rtype uint16
		var reE *MyError.MyError
		for _, mr := range RR {
			fmt.Println(utils.GetDebugLine(), mr.RR)
			hdr := dns.RR_Header{
				Name:   dst,
				Class:  mr.RR.Class,
				Rrtype: mr.RR.RrType,
				Ttl:    mr.RR.Ttl,
			}
			if mr.RR.RrType == dns.TypeA {
				rh := &dns.A{
					Hdr: hdr,
					A:   utils.StrToIP(mr.RR.Target),
				}
				R = append(R, dns.RR(rh))
				rtype = dns.TypeA
				fmt.Println(utils.GetDebugLine(), "Get A RR from MySQL, requery dst:", dst)
			} else if mr.RR.RrType == dns.TypeCNAME {
				rh := &dns.CNAME{
					Hdr:    hdr,
					Target: mr.RR.Target,
				}
				R = append(R, dns.RR(rh))
				rtype = dns.TypeCNAME
				fmt.Println(utils.GetDebugLine(), "Get CNAME RR from MySQL, requery dst:", dst)
				reE = MyError.NewError(MyError.ERROR_NOTVALID,
					"Got CNAME result for dst : "+dst+" with srcIP : "+srcIP)
			}
		}
		fmt.Println(utils.GetDebugLine(), R)
		if len(R) > 0 {
			return true, R, rtype, reE
		}
	}
	return false, nil, uint16(0), MyError.NewError(MyError.ERROR_UNKNOWN, utils.GetDebugLine()+"Unknown Error ")

}

func GetAFromDNSBackend(
	dst, srcIP string,
	ns_a []string,
	regionTree *domain.RegionTree) (bool, []dns.RR, uint16, *MyError.MyError) {

	var reE *MyError.MyError = nil
	var rtype uint16
	rr, edns_h, edns, e := query.QueryA(dst, srcIP, ns_a, "53")
	if e == nil && rr != nil {
		var rr_i []dns.RR
		if a, ok := query.ParseA(rr, dst); ok {
			//rr is A record
			fmt.Print(utils.GetDebugLine(), "GetAFromDNSBackend : typeA ")
			fmt.Println(utils.GetDebugLine(), a, ok)
			for _, i := range a {
				rr_i = append(rr_i, dns.RR(i))
			}
			//if A ,need parse edns client subnet
			//			return true,rr_i,nil
			rtype = dns.TypeA
		} else if b, ok := query.ParseCNAME(rr, dst); ok {
			//rr is CNAME record
			fmt.Println(utils.GetDebugLine(), "GetAFromDNSBackend: typeCNAME ", b, ok)
			dst = b[0].Target
			for _, i := range b {
				rr_i = append(rr_i, dns.RR(i))
			}
			rtype = dns.TypeCNAME
			reE = MyError.NewError(MyError.ERROR_NOTVALID,
				"Got CNAME result for dst : "+dst+" with srcIP : "+srcIP)
			//if CNAME need parse edns client subnet
		} else {
			//error return and retry
			fmt.Println(utils.GetDebugLine(), "GetAFromDNSBackend: ", rr)
			return false, nil, uint16(0), MyError.NewError(MyError.ERROR_NORESULT,
				"Got error result, need retry for dst : "+dst+" with srcIP : "+srcIP)
		}
		// Parse edns client subnet
		fmt.Println(utils.GetDebugLine(), "GetAFromDNSBackend: ", edns_h, edns)
		//		go func(regionTree *domain.RegionTree) {
		var ipnet *net.IPNet
		if edns != nil {
			ipnet, e = utils.ParseEdnsIPNet(edns.Address, edns.SourceScope, edns.Family)
		}
		fmt.Println(utils.GetDebugLine(), "GetAFromDNSBackend: ", e)

		startIP, endIP := iplookup.GetIpinfoStartEndWithIPString(srcIP)
		cidrmask := utils.GetCIDRMaskWithUint32Range(startIP, endIP)
		fmt.Println(utils.GetDebugLine(), "iplookup.GetIpinfoStartEndWithIPString with srcIP: ",
			srcIP, " StartIP : ", startIP, "==", utils.Int32ToIP4(startIP).String(),
			" EndIP: ", endIP, "==", utils.Int32ToIP4(endIP).String(), " cidrmask : ", cidrmask)
		if ipnet != nil {
			netaddr, mask := utils.IpNetToInt32(ipnet)
			fmt.Println(utils.GetDebugLine(), "Got Edns client subnet from ecs query, netaddr : ", netaddr,
				" mask : ", mask)
			if (netaddr != startIP) || (mask != cidrmask) {
				fmt.Println(utils.GetDebugLine(), "iplookup data dose not match edns query result , netaddr : ",
					netaddr, "<->", startIP, " mask: ", mask, "<->", cidrmask)
			}
			r, _ := domain.NewRegion(rr_i, startIP, cidrmask)
			regionTree.AddRegionToCache(r)
			fmt.Print(utils.GetDebugLine(), "GetAFromDNSBackend: ")
			fmt.Println(regionTree.GetRegionFromCacheWithAddr(startIP, cidrmask))

		} else {
			//todo: get StartIP/EndIP from iplookup module

			//				netaddr, mask := domain.DefaultNetaddr, domain.DefaultMask
			r, _ := domain.NewRegion(rr_i, startIP, cidrmask)
			regionTree.AddRegionToCache(r)
			fmt.Println(utils.GetDebugLine(), "GetAFromDNSBackend: ", r)
			fmt.Println(regionTree.GetRegionFromCacheWithAddr(startIP, cidrmask))
		}
		//		}(regionTree)

		return true, rr_i, rtype, reE
	}
	return false, nil, rtype, MyError.NewError(MyError.ERROR_UNKNOWN, utils.GetDebugLine()+"Unknown error")
}
