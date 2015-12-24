package domain

import (
	"net"
	"query"
	"reflect"
	"server"
	"testing"
	"utils"

	"github.com/miekg/dns"
)

func TestNewDomainDB(t *testing.T) {
	if DomainRRCache == nil {
		t.Log("DomainDB is nil")
	}
	t.Log(reflect.TypeOf(DomainRRCache))
	t.Log(reflect.ValueOf(DomainRRCache))
	t.Log(reflect.TypeOf(DomainSOACache))
	t.Log(reflect.ValueOf(DomainSOACache))
}

func TestQueryDomainSOAandNS(t *testing.T) {
	ds_array := []string{
		"www.baidu.com",
		"www.a.shifen.com",
		"a.shifen.com",
		"www2.sinaimg.cn",
		"weboimg.gslb.sinaedge.com",
		"weibo.cn",
		"ww3.sinaimg.cn",
		"api.weibo.cn",
		"img.alicdn.com",
		"danuoyi.alicdn.com",
		"img.alicdn.com.danuoyi.alicdn.com.",
	}
	for _, k := range ds_array {
		t.Log(k)
		soa, ns, e := query.QuerySOA(k)
		if e != nil {
			t.Log(e)
			t.Log(soa)
			t.Log(ns)
			t.Fatal()
		} else {
			t.Log(soa)
			t.Log(ns)
			soa_t := &DomainSOANode{
				SOAKey: soa.Hdr.Name,
				NS:     ns,
				SOA:    soa,
			}
			soa_n, e := DomainSOACache.GetDomainSOANodeFromCache(soa_t)
			if e == nil {
				t.Log("Got in DomainSOADB for " + soa.Hdr.Name)
				t.Log(soa_n)
			} else {
				b, e := DomainSOACache.StoreDomainSOANodeToCache(soa_t)
				if b == true && e == nil {
					t.Log("DomainSOADB stored ", soa_t)
				} else {
					t.Log(b)
					t.Log(e)
					t.Fatal()
				}
			}
		}
		t.Log("+++++++++++++++++++end search+++++++++++")
		t.Log("=========================================")

	}
}

func TestNewRegion(t *testing.T) {
	d_arr := []string{
		//		"www.baidu.com",
		"www.a.shifen.com",
		"api.weibo.cn",
		"ww2.sinaimg.cn",
		"weibo.cn",
		"www.qq.com",
		"www.yahoo.com",
		//		"www.google.com",
	}
	for _, d := range d_arr {

		a_rr, ipnet, e := GeneralDNSBackendQuery(d, server.GetClientIP())
		t.Log(a_rr, ipnet, e)
		ip, mask := utils.IpNetToInt32(ipnet)
		if a_rr == nil {
			t.Log("a_rr is Nil ", a_rr)
			continue
		}
		r1, e := NewRegion(a_rr, ip, mask)
		if e != nil {
			t.Log(e)
			t.Fail()
		} else {
			t.Log(r1)
			t.Log(r1.UpdateTime.UnixNano())

		}
		t.Log("--------------------------")

	}
}

func TestInitRegionTree(t *testing.T) {
	d_arr := []string{
		"www.baidu.com",
		"www.a.shifen.com",
		"api.weibo.cn",
		"ww2.sinaimg.cn",
		"weibo.cn",
		"www.qq.com",
		"www.yahoo.com",
		"www.google.com",
	}
	for _, d := range d_arr {
		soa, ns, e := query.QuerySOA(d)

		if e != nil {
			t.Log(e)
			t.Fatal()
		} else {
			t.Log(ns)
			t.Log(soa)
			dn, e := NewDomainNode(d, soa.Hdr.Name, soa.Expire)
			if e != nil {
				t.Log(e)

				//				t.Fatal()
			} else {
				dn.InitRegionTree()
				t.Log(dn)
			}
		}
	}
}

func TestA(t *testing.T) {
	d := "www.a.shifen.com"
	drr, e := DomainRRCache.GetDomainNodeFromCacheWithName(d)
	if e == nil {
		// no error,RR is in RRDB,need check type,if CNAME,need refetch A recode for than CNAME rr
		if drr != nil {
			x := utils.Ip4ToInt32(net.ParseIP("124.207.129.171"))
			rr, e := drr.DomainRegionTree.GetRegionFromCacheWithAddr(x, 32)
			if e == nil {
				//got rr
				t.Log(rr)
				//need return
				if rr.RrType == dns.TypeA {
					t.Log("return rr")
				} else if rr.RrType == dns.TypeCNAME {
					t.Log("return cname ,need query A")
					// QueryA
				}

			} else {
				// err not nil
				// need query A / CNAME with NS that in DomainSOATree
				t.Log("error not nil,need query A / CNAME with NS that in DomainSOATree")

			}
		} else {
			t.Log("unknown error")
			// Unknown error
		}

	} else {
		//No domainNode,need requery all (SOA/NS/CNAME/A/Edns)
		t.Log(e)
		var soa *dns.SOA
		var ns []*dns.NS
		soanode, e := DomainSOACache.GetDomainSOANodeFromCacheWithDomainName(d)
		if e == nil {
			//got soa record from DB,and use ns server to query a
			t.Log(soanode)
			ns = soanode.NS

		} else {
			// QuerySOA
			soa, ns, e = query.QuerySOA(d)
			if e == nil {
				//store soa & ns into SOADB
				DomainSOACache.StoreDomainSOANodeToCache(&DomainSOANode{
					SOAKey: d,
					NS:     ns,
					SOA:    soa,
				})
			} else {
				//QuerySOA failed,need retry, or exit
				t.Fatal()
			}
		}
		if ns != nil {
			//Query A
			rr, i, edns, e := query.QueryA(d, true, ns[0].Ns, "53")
			if e != nil {
				t.Log(e)
			} else {
				t.Log(rr)
				t.Log(i)
				t.Log(edns)
				if i != nil && edns != nil {
					//Edns is not null
					t.Log("edns is not nill,need parse + " + d)
					t.Log(i)
					t.Log(edns)
				}
				if rr != nil {
					//RR is array
					for _, r := range rr {
						rh := r.Header().Rrtype
						switch rh {
						case dns.TypeA:
							//Got A record ,return
							t.Log(r.(*dns.A))
						case dns.TypeCNAME:
							//Query CNAME's Target
							if rc, ok := r.(*dns.CNAME); ok {
								rc_soa, ns, e := query.QuerySOA(rc.Target)
								t.Log(rc_soa)
								t.Log(ns)
								if e == nil && ns != nil {
									rr, i, edns, e := query.QueryA(rc.Target, true, ns[0].Ns, "53")
									t.Log(rr)
									t.Log(i)
									t.Log(edns)
									t.Log(e)
								} else {
									// Need retry
									t.Log(e)
								}
							}
						default:
							t.Log(rr)
						}
					}
				}
			}
		} else {
			t.Fatal()
		}
	}
}
