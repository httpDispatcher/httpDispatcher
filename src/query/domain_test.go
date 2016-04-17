package query

import (
	"net"
	"reflect"
	"testing"
	"utils"

	"MyError"

	"github.com/miekg/bitradix"
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

var ds_array = []string{
	"www.baidu.com",
	"www.a.shifen.com",
	"a.shifen.com",
	//"www2.sinaimg.cn",
	//"weboimg.gslb.sinaedge.com",
	//"weibo.cn",
	//"ww3.sinaimg.cn",
	//"api.weibo.cn",
	"img.alicdn.com",
	"danuoyi.alicdn.com",
	"img.alicdn.com.danuoyi.alicdn.com.",
}

var ip_arr = []string{
	"202.106.0.20",
	"112.112.112.112",
	"11.12.13.14",
	"210.35.249.1",
	"198.6.3.5",
	"124.207.129.171",
}

func TestQueryDomainSOAandNS(t *testing.T) {
	for _, k := range ds_array {
		t.Log(k)
		soa, ns, e := QuerySOA(k)
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

func buildDomainSOACache(b *testing.B, ds_arr []string) {

	for _, d := range ds_arr {
		soa, ns_arr, e := QuerySOA(d)
		if e != nil {
			b.Log(soa, ns_arr, e)
			b.Fatal(e)
		} else {
			dn, e := NewDomainNode(d, soa.Hdr.Name, soa.Hdr.Ttl)
			if e != nil {
				b.Log(dn, e)
				b.Fatal()
			}
			ok, e := DomainRRCache.StoreDomainNodeToCache(dn)
			if ok != true || e != nil {
				b.Log(dn, ok, e)
				b.Fatal()
			}
			soa_node := NewDomainSOANode(soa, ns_arr)
			ok, e = DomainSOACache.StoreDomainSOANodeToCache(soa_node)
			if ok != true || e != nil {
				b.Log(ok)
				b.Fatal(e)
			}
		}
	}
}

func BenchmarkGetDomainSOANodeFromCache(b *testing.B) {
	buildDomainSOACache(b, ds_array)
	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			DomainSOACache.GetDomainSOANodeFromCache(&DomainSOANode{
				SOAKey: "alicdn.com",
			})
		}
	})
	b.StopTimer()
	b.ReportAllocs()

}

func initRegionTree(d string) (*DomainNode, *MyError.MyError) {
	soa, ns_arr, e := QuerySOA(d)
	if e != nil {
		//		fmt.Println(ns_arr)
		return nil, e
	}
	soa_node := NewDomainSOANode(soa, ns_arr)
	ok, e := DomainSOACache.StoreDomainSOANodeToCache(soa_node)
	if ok == true && e == nil {
		dn, e := NewDomainNode(d, soa.Hdr.Name, soa.Expire)
		ok, e := DomainRRCache.StoreDomainNodeToCache(dn)
		if ok == true && e == nil {
			dn.InitRegionTree()
			return dn, nil
		}
	}
	return nil, e

}

func TestInitRegionTree(t *testing.T) {

	for _, d := range ds_array {
		dn, e := initRegionTree(d)
		if e != nil {
			t.Fatal(e)
		} else {
			t.Log(dn)
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
			soa, ns, e = QuerySOA(d)
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
			var ns_a []string
			for _, n := range ns {
				ns_a = append(ns_a, n.Ns)
			}
			rr, i, edns, e := QueryA(d, "202.106.0.20", ns_a, "53")
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
								rc_soa, ns, e := QuerySOA(rc.Target)
								t.Log(rc_soa)
								t.Log(ns)
								if e == nil && ns != nil {
									var ns_a []string
									for _, n := range ns {
										ns_a = append(ns_a, n.Ns)
									}
									rr, i, edns, e := QueryA(rc.Target, "202.106.0.20", ns_a, "53")
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

func TestStoreDomainNodeToCache(t *testing.T) {

}

func TestMubitRadix(t *testing.T) {
	cidrNet := []string{
		"10.0.0.2/8",
		"10.20.0.0/14",
		"10.21.0.0/16",
		"192.168.0.0/16",
		"192.168.2.0/24",
		"8.0.0.0/9",
		"8.8.8.0/24",
		"0.0.0.0/0",
		//		"128.0.0.0/1",
	}
	ip2Find := []string{
		"10.20.1.2",
		"10.22.1.2",
		"10.19.0.1",
		"10.21.0.1",
		"192.168.2.3",
		"10.22.0.5",
		"202.106.0.20",
		"172.16.3.133",
		"8.8.8.8",
		"8.8.7.1",
	}
	RadixTree := NewDomainRegionTree()
	for _, x := range cidrNet {
		i, n, e := net.ParseCIDR(x)
		if e != nil {
			t.Log(e.Error())
			t.Fail()
			continue
		}
		a, m := utils.IpNetToInt32(n)
		RadixTree.AddRegionToCache(&Region{
			NetworkAddr: a,
			NetworkMask: m,
			RR: []dns.RR{
				dns.RR(&dns.A{
					A: i,
					Hdr: dns.RR_Header{
						Rrtype: 1,
						Class:  1,
						Ttl:    60,
					},
				}),
			},
		})
	}
	// For default route
	RadixTree.AddRegionToCache(&Region{
		NetworkAddr: DefaultRadixNetaddr,
		NetworkMask: DefaultRadixNetMask,
		RR: []dns.RR{
			dns.RR(&dns.A{
				A: utils.Int32ToIP4(DefaultRadixNetaddr),
				Hdr: dns.RR_Header{
					Rrtype: 1,
					Class:  1,
					Ttl:    60,
				},
			}),
		},
	})
	for _, i := range ip2Find {

		ii := utils.Ip4ToInt32(utils.StrToIP(i))
		r, e := RadixTree.GetRegionFromCacheWithAddr(ii, DefaultRadixSearchMask)
		if e != nil {
			t.Log(e)
			t.Log(i)
			t.Fail()
		} else {
			t.Log(r)
		}

	}

	RadixTree.Radix32.Do(func(r1 *bitradix.Radix32, i int) {
		t.Log(r1.Key(), r1.Value, r1.Bits())
	})
}
