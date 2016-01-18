package server

import (
	"net"
	"os"
	"reflect"
	"strconv"
	"time"

	"MyError"
	"config"
	"domain"
	"query"
	"utils"

	"github.com/miekg/dns"
)

func GetSOARecord(d string) (*domain.DomainSOANode, *MyError.MyError) {

	var soa *domain.DomainSOANode

	dn, e := domain.DomainRRCache.GetDomainNodeFromCacheWithName(d)
	if e == nil && dn != nil {
		dsoa_key := dn.SOAKey
		soa, e = domain.DomainSOACache.GetDomainSOANodeFromCacheWithDomainName(dsoa_key)
		utils.ServerLogger.Debug("GetDomainSOANodeFromCacheWithDomainName: key: %s soa %v", dsoa_key, soa)
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

	//fmt.Println(utils.GetDebugLine(), "^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^")

	//Can't loop for CNAME chain than bigger than 10
	for dst := d; (bigloopflag == false) && (c < 10); c++ {
		//fmt.Println(utils.GetDebugLine(), "GetARecord : ", dst, " srcIP: ", srcIP)
		utils.ServerLogger.Debug("GetARecord : %s srcIP: %s", dst, srcIP)

		ok, dn, RR, e := GetAFromCache(dst, srcIP)
		//fmt.Println(utils.GetDebugLine(), " GetAFromCache return : ", ok, dn, RR, e)
		utils.ServerLogger.Debug("GetAFromCache return: ", ok, dn, RR, e)
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

		soa, e := GetSOARecord(dst)
		//fmt.Println(utils.GetDebugLine(), "GetARecord: GetSOARecord return : ", soa, " error: ", e)
		utils.ServerLogger.Debug("GetARecord return: ", soa, " error: ", e)
		if e != nil {
			// error! need return
			//fmt.Print(utils.GetDebugLine(), "GetARecord: ")
			utils.ServerLogger.Error("GetARecord error: %s", e.Error())
			//fmt.Println(e)
			//fmt.Println("error111,need return")
		}
		var cacheflag bool = false
		var cc = 0
		//fmt.Println(utils.GetDebugLine(), " Debuginfo : init Cache dn , cacheflag: ", cacheflag,
		//	" cc: ", cc, " (cacheflag == false) && (cc < 5) ", (cacheflag == false) && (cc < 5))
		for cacheflag = false; (cacheflag == false) && (cc < 5); {
			// wait for goroutine 'StoreDomainNodeToCache' in GetSOARecord to be finished
			dn, e = domain.DomainRRCache.GetDomainNodeFromCacheWithName(dst)
			if e != nil {
				// here ,may be nil
				// error! need return
				//fmt.Println(utils.GetDebugLine(),
				//	" GetARecord : have not got cache GetDomainNodeFromCacheWithName, need waite ", e)
				utils.ServerLogger.Error("GetARecord : have not got cache GetDomainNodeFromCacheWithName, need waite %s", e.Error())
				time.Sleep(1 * time.Second)
				if e.ErrorNo == MyError.ERROR_NOTFOUND {
					cc++
				} else {
					utils.ServerLogger.Critical("GetARecord error %s", e.Error())
					//fmt.Println(utils.GetDebugLine(), e)
					os.Exit(3)
				}
			} else {
				cacheflag = true
				//fmt.Println(utils.GetDebugLine(), "GetARecord: ", dn)
				utils.ServerLogger.Debug("GetARecord dn: %v", dn)
			}
		}
		if e != nil || len(soa.NS) <= 0 {
			//GetSOA failed , need log and return
			return false, nil, MyError.NewError(MyError.ERROR_UNKNOWN,
				"GetARecord func GetSOARecord failed: "+d)
		}
		dn.InitRegionTree()
		Regiontree = dn.DomainRegionTree

		//fmt.Println(utils.GetDebugLine(), "++++++++++++++++++++++++++++++++++++++++++++++")
		if config.IsLocalMysqlBackend(dst) {
			//fmt.Println(utils.GetDebugLine(), "**********************************************")
			//need pass dn to GetAFromMySQLBackend, to fill th dn.RegionTree node
			ok, RR, rtype, ee := GetAFromMySQLBackend(dst, srcIP, Regiontree)
			//fmt.Println(utils.GetDebugLine(), " Debug: GetAFromMySQLBackend: return ", ok,
			//	" RR: ", RR, " error: ", ee)
			utils.ServerLogger.Debug("GetAFromMySQLBackend: return ", ok, RR, rtype, ee)
			if !ok {
				//fmt.Println(utils.GetDebugLine(), "Error: GetAFromMySQL error : ", ee)
				utils.ServerLogger.Error("Error: GetAFromMySQL error : ", ee)
			} else if rtype == dns.TypeA {
				//fmt.Println(utils.GetDebugLine(), "Info: Got A record, : ", RR)
				utils.ServerLogger.Debug("Got A record: ", RR)
				return true, RR, nil
			} else if rtype == dns.TypeCNAME {
				//fmt.Println(utils.GetDebugLine(), "Info: Got CNAME record, ReGet dst : ", dst, RR)
				utils.ServerLogger.Debug("Got CNAME record, ReGet dst: ", dst, RR)
				dst = RR[0].(*dns.CNAME).Target
				continue
			}
		}

		//fmt.Println(utils.GetDebugLine(), "Info: Got dst: ", dst, " srcIP: ", srcIP, " soa.NS: ", soa.NS)
		utils.ServerLogger.Debug("Got dst: ", dst, " srcIP: ", srcIP, " soa.NS: ", soa.NS)
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
	//fmt.Println(utils.GetDebugLine(), "GetARecord: ", Regiontree)
	utils.ServerLogger.Debug("GetARecord Regiontree: %s", Regiontree)
	return false, nil, MyError.NewError(MyError.ERROR_UNKNOWN, "Unknown error")
}

func GetAFromCache(dst, srcIP string) (bool, *domain.DomainNode, []dns.RR, *MyError.MyError) {
	dn, e := domain.DomainRRCache.GetDomainNodeFromCacheWithName(dst)
	//fmt.Println(utils.GetDebugLine(), "GetAFromCache:", dn, e)
	if e == nil && dn != nil && dn.DomainRegionTree != nil {
		//Get DomainNode succ,
		r, e := dn.DomainRegionTree.GetRegionFromCacheWithAddr(
			utils.Ip4ToInt32(net.ParseIP(srcIP)), domain.DefaultRedaxSearchMask)
		//fmt.Println(utils.GetDebugLine(), "GetAFromCache : ", r, e)
		if e == nil {
			//fmt.Println(utils.GetDebugLine(), "GetAFromCache: Gooooot: ", r)
			if r.RrType == dns.TypeA {
				//fmt.Println(utils.GetDebugLine(), "GetAFromCache: Goooot A", r.RR)
				utils.ServerLogger.Debug("GetAFromCache: Goooot A ", r.RR)
				return true, dn, r.RR, nil
			} else if r.RrType == dns.TypeCNAME {
				//fmt.Println(utils.GetDebugLine(), "GetAFromCache : Goooot CNAME", r.RR)
				utils.ServerLogger.Debug("GetAFromCache: Goooot CNAME ", r.RR)
				if len(r.RR) > 0 {
					//					dst = r.RR[0].(*dns.CNAME).Target
					return false, dn, r.RR, MyError.NewError(MyError.ERROR_NOTVALID,
						"Get CNAME From,Requery A for "+r.RR[0].(*dns.CNAME).Target)
				} else {
					//fmt.Println(utils.GetDebugLine(), "Error Got RegionFromCacheWithAddr", r.RR)
					return false, dn, nil, MyError.NewError(MyError.ERROR_NORESULT, "Error Got RegionFromCacheWithAddr ")
				}
				//				continue
			}
		} else if e.ErrorNo == MyError.ERROR_NOTFOUND {
			//fmt.Println(utils.GetDebugLine(), "Not found r, need query dns")
			return false, dn, nil, MyError.NewError(MyError.ERROR_NOTFOUND,
				"Not found R in cache, dst :"+dst+" srcIP "+srcIP)
		}
		// return
	} else if e == nil && dn != nil && dn.DomainRegionTree == nil {
		// Get domainNode in cache tree,but no RR in region tree,need query with NS
		// if RegionTree is nil, init RegionTree First
		ok, e := dn.InitRegionTree()
		if e != nil {
			utils.ServerLogger.Error("InitRegionTree fail %s", e.Error())
		}
		//
		//fmt.Println("RegionTree is nil ,Init it: "+reflect.ValueOf(ok).String(), e)
		utils.ServerLogger.Debug("RegionTree is nil ,Init it: %s ", reflect.ValueOf(ok).String())
		return false, dn, nil, MyError.NewError(MyError.ERROR_NORESULT,
			"Get domainNode in cache tree,but no RR in region tree,need query with NS, dst : "+dst+" srcIP "+srcIP)
	} else {
		// e != nil
		// RegionTree is not nil
		//fmt.Print(utils.GetDebugLine(), "GetAFromCache dst: "+dst+" srcIP: "+srcIP)
		//fmt.Println(dn, e)
		if e.ErrorNo != MyError.ERROR_NOTFOUND {
			//fmt.Println("Found unexpected error, need return !")
			utils.ServerLogger.Critical("Found unexpected error, need return !")
			os.Exit(2)
		} else {
			return false, nil, nil, e
		}

	}
	return false, nil, nil, MyError.NewError(MyError.ERROR_UNKNOWN, "Unknown error!")
}

func GetAFromMySQLBackend(dst, srcIP string, regionTree *domain.RegionTree) (bool, []dns.RR, uint16, *MyError.MyError) {
	domainId, e := query.RRMySQL.GetDomainIDFromMySQL(dst)
	if e != nil {
		//todo:
		//fmt.Println(utils.GetDebugLine(), "Error, GetDomainIDFromMySQL:", e)
		return false, nil, uint16(0), e
	}
	region, ee := query.RRMySQL.GetRegionWithIPFromMySQL(utils.Ip4ToInt32(utils.StrToIP(srcIP)))
	if ee != nil {
		//fmt.Println(utils.GetDebugLine(), "Error GetRegionWithIPFromMySQL:", ee)
		return false, nil, uint16(0), MyError.NewError(ee.ErrorNo, "GetRegionWithIPFromMySQL return "+e.Error())
	}
	RR, eee := query.RRMySQL.GetRRFromMySQL(uint32(domainId), region.IdRegion)
	if eee != nil && eee.ErrorNo == MyError.ERROR_NORESULT {
		//fmt.Println(utils.GetDebugLine(), "Error GetRRFromMySQL with DomainID:", domainId,
		//	"RegionID:", region.IdRegion, eee)
		//fmt.Println(utils.GetDebugLine(), "Try to GetRRFromMySQL with Default Region")
		utils.ServerLogger.Debug("Try to GetRRFromMySQL with Default Region")
		RR, eee = query.RRMySQL.GetRRFromMySQL(uint32(domainId), uint32(0))
		if eee != nil {
			//fmt.Println(utils.GetDebugLine(), "Error GetRRFromMySQL with DomainID:", domainId,
			//	"RegionID:", 0, eee)
			return false, nil, uint16(0), MyError.NewError(eee.ErrorNo, "Error GetRRFromMySQL with DomainID:"+strconv.Itoa(domainId)+eee.Error())
		}
	} else if eee != nil {
		utils.ServerLogger.Error(eee.Error())
		return false, nil, uint16(0), eee
	}
	//fmt.Println(utils.GetDebugLine(), "GetRRFromMySQL Succ!:", RR)
	utils.ServerLogger.Debug("GetRRFromMySQL Succ!: ", RR)
	var R []dns.RR
	var rtype uint16
	var reE *MyError.MyError
	hdr := dns.RR_Header{
		Name:   dst,
		Class:  RR.RR.Class,
		Rrtype: RR.RR.RrType,
		Ttl:    RR.RR.Ttl,
	}

	//fmt.Println(utils.GetDebugLine(), mr.RR)
	if RR.RR.RrType == dns.TypeA {
		for _, mr := range RR.RR.Target {
			rh := &dns.A{
				Hdr: hdr,
				A:   utils.StrToIP(mr),
			}
			R = append(R, dns.RR(rh))
		}
		rtype = dns.TypeA
		//	fmt.Println(utils.GetDebugLine(), "Get A RR from MySQL, requery dst:", dst)
	} else if RR.RR.RrType == dns.TypeCNAME {
		for _, mr := range RR.RR.Target {
			rh := &dns.CNAME{
				Hdr:    hdr,
				Target: mr,
			}
			R = append(R, dns.RR(rh))
		}
		rtype = dns.TypeCNAME
		//fmt.Println(utils.GetDebugLine(), "Get CNAME RR from MySQL, requery dst:", dst)
		reE = MyError.NewError(MyError.ERROR_NOTVALID,
			"Got CNAME result for dst : "+dst+" with srcIP : "+srcIP)
	}

	if len(R) > 0 {
		//Add timer for auto refrech the RegionCache
		go func(dst, srcIP string, r dns.RR, regionTree *domain.RegionTree) {
			//fmt.Println(utils.GetDebugLine(), " Refresh record after ", r.Header().Ttl-5,
			//	" Second, dst: ", dst, " srcIP: ", srcIP, "add timer ")
			time.AfterFunc(time.Duration(r.Header().Ttl-5)*time.Second,
				func() { GetAFromMySQLBackend(dst, srcIP, regionTree) })
		}(dst, srcIP, R[0], regionTree)

		go func(regionTree *domain.RegionTree, R []dns.RR, srcIP string) {
			//fmt.Println(utils.GetDebugLine(), "GetAFromMySQLBackend: ", e)

			startIP, endIP := region.Region.StarIP, region.Region.EndIP
			cidrmask := utils.GetCIDRMaskWithUint32Range(startIP, endIP)

			//fmt.Println(utils.GetDebugLine(), " GetRegionWithIPFromMySQL with srcIP: ",
			//	srcIP, " StartIP : ", startIP, "==", utils.Int32ToIP4(startIP).String(),
			//	" EndIP: ", endIP, "==", utils.Int32ToIP4(endIP).String(), " cidrmask : ", cidrmask)
			//				netaddr, mask := domain.DefaultNetaddr, domain.DefaultMask
			r, _ := domain.NewRegion(R, startIP, cidrmask)
			regionTree.AddRegionToCache(r)
			//fmt.Println(utils.GetDebugLine(), "GetAFromMySQLBackend: ", r)
			//				fmt.Println(regionTree.GetRegionFromCacheWithAddr(startIP, cidrmask))
		}(regionTree, R, srcIP)
		return true, R, rtype, reE
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
			//fmt.Println(utils.GetDebugLine(), "GetAFromDNSBackend : typeA record: ", a, " dns.TypeA: ", ok)
			utils.ServerLogger.Debug("GetAFromDNSBackend : typeA record: ", a, " dns.TypeA: ", ok)
			for _, i := range a {
				rr_i = append(rr_i, dns.RR(i))
			}
			//if A ,need parse edns client subnet
			//			return true,rr_i,nil
			rtype = dns.TypeA
		} else if b, ok := query.ParseCNAME(rr, dst); ok {
			//rr is CNAME record
			//fmt.Println(utils.GetDebugLine(), "GetAFromDNSBackend: typeCNAME record: ", b, " dns.TypeCNAME: ", ok)
			utils.ServerLogger.Debug("GetAFromDNSBackend: typeCNAME record: ", b, " dns.TypeCNAME: ", ok)
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
			//fmt.Println(utils.GetDebugLine(), "GetAFromDNSBackend: ", rr)
			utils.ServerLogger.Debug("GetAFromDNSBackend: ", rr)
			return false, nil, uint16(0), MyError.NewError(MyError.ERROR_NORESULT,
				"Got error result, need retry for dst : "+dst+" with srcIP : "+srcIP)
		}

		//Add timer for auto refrech the RegionCache
		go func(dst, srcIP string, ns_a []string, r dns.RR, regionTree *domain.RegionTree) {
			//fmt.Println(utils.GetDebugLine(), " Refresh record after ", r.Header().Ttl-5,
			//	" Second, dst: ", dst, " srcIP: ", srcIP, " ns_a: ", ns_a, "add timer ")
			time.AfterFunc((time.Duration(r.Header().Ttl-5))*time.Second,
				func() { GetAFromDNSBackend(dst, srcIP, ns_a, regionTree) })
		}(dst, srcIP, ns_a, rr_i[0], regionTree)

		// Parse edns client subnet
		//fmt.Println(utils.GetDebugLine(), "GetAFromDNSBackend: ", " edns_h: ", edns_h, " edns: ", edns)
		utils.ServerLogger.Debug("GetAFromDNSBackend: ", " edns_h: ", edns_h, " edns: ", edns)

		go func(regionTree *domain.RegionTree, R []dns.RR, edns *dns.EDNS0_SUBNET, srcIP string) {
			//todo: Need to be combined with the go func within GetAFromMySQLBackend
			var startIP, endIP uint32

			if config.RC.MySQLEnabled {
				region, ee := query.RRMySQL.GetRegionWithIPFromMySQL(utils.Ip4ToInt32(utils.StrToIP(srcIP)))
				if ee != nil {
					//fmt.Println(utils.GetDebugLine(), "Error GetRegionWithIPFromMySQL:", ee)
					utils.ServerLogger.Error("Error GetRegionWithIPFromMySQL: %s", ee.Error())
				} else {
					startIP, endIP = region.Region.StarIP, region.Region.EndIP
					//fmt.Println(utils.GetDebugLine(), region.Region, startIP, endIP)
				}

			} else {
				//				startIP, endIP = iplookup.GetIpinfoStartEndWithIPString(srcIP)
			}
			cidrmask := utils.GetCIDRMaskWithUint32Range(startIP, endIP)
			//fmt.Println(utils.GetDebugLine(), "Search client region info with srcIP: ",
			//	srcIP, " StartIP : ", startIP, "==", utils.Int32ToIP4(startIP).String(),
			//	" EndIP: ", endIP, "==", utils.Int32ToIP4(endIP).String(), " cidrmask : ", cidrmask)
			if edns != nil {
				var ipnet *net.IPNet

				ipnet, e = utils.ParseEdnsIPNet(edns.Address, edns.SourceScope, edns.Family)
				netaddr, mask := utils.IpNetToInt32(ipnet)
				//fmt.Println(utils.GetDebugLine(), "Got Edns client subnet from ecs query, netaddr : ", netaddr,
				//	" mask : ", mask)
				utils.ServerLogger.Debug("Got Edns client subnet from ecs query, netaddr : ", netaddr, " mask : ", mask)
				if (netaddr != startIP) || (mask != cidrmask) {
					//fmt.Println(utils.GetDebugLine(), "iplookup data dose not match edns query result , netaddr : ",
					//	netaddr, "<->", startIP, " mask: ", mask, "<->", cidrmask)
					utils.ServerLogger.Debug("iplookup data dose not match edns query result , netaddr : ", netaddr, "<->", startIP, " mask: ", mask, "<->", cidrmask)
				}
				// if there is no region info in region table of mysql db or no info in ipdb
				if cidrmask <= 0 || startIP <= 0 {
					startIP = netaddr
					cidrmask = mask
				}
				r, _ := domain.NewRegion(R, startIP, cidrmask)
				//todo: modify to go func,so you can cathe the result
				regionTree.AddRegionToCache(r)
				//fmt.Print(utils.GetDebugLine(), "GetAFromDNSBackend: ")
				//				fmt.Println(regionTree.GetRegionFromCacheWithAddr(startIP, cidrmask))

			} else {
				//todo: get StartIP/EndIP from iplookup module

				//				netaddr, mask := domain.DefaultNetaddr, domain.DefaultMask
				r, _ := domain.NewRegion(R, startIP, cidrmask)
				//todo: modify to go func,so you can cathe the result
				regionTree.AddRegionToCache(r)
				//fmt.Println(utils.GetDebugLine(), "GetAFromDNSBackend: AddRegionToCache: ", r)
				//fmt.Println(regionTree.GetRegionFromCacheWithAddr(startIP, cidrmask))
			}
		}(regionTree, rr_i, edns, srcIP)

		return true, rr_i, rtype, reE
	}
	return false, nil, rtype, MyError.NewError(MyError.ERROR_UNKNOWN, utils.GetDebugLine()+"Unknown error")
}
