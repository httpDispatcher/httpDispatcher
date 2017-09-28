package query

//go test -v query -run=none -benchmem -benchtime 10s -count 4 -cpuprofile querybench.cpu.out -memprofile querybench.mem.out -bench "."
import (
	"MyError"
	"config"
	"reflect"
	"testing"
	"utils"

	"github.com/miekg/dns"
)

var cnamemap = map[string]string{
	"ww2.sinaimg.cn":             "weiboimg.gslb.sinaedge.com.",
	"www.baidu.com":              "www.a.shifen.com.",
	"weiboimg.gslb.sinaedge.com": "weiboimg.grid.sinaedge.com.",
}
var ttlmap = map[string]uint32{
	"ww2.sinaimg.cn":             uint32(600),
	"www.baidu.com":              uint32(1200),
	"weiboimg.gslb.sinaedge.com": uint32(600),
}

func TestMain(m *testing.M) {
	config.ParseConf("../conf/httpdispatcher.toml")
	if config.RC.MySQLEnabled {
		RC_MySQLConf = config.RC.MySQLConf
		InitMySQL(RC_MySQLConf)
	}
	utils.InitLogger()

	m.Run()
	//	os.Exit(1)
}

func TestPackEdns0SubnetOPT(t *testing.T) {
	var sm = uint8(32)
	var sc = uint8(0)
	var ip = "124.207.129.171"
	o := PackEdns0SubnetOPT(ip, sm, sc)
	t.Log(o)
	if o.Hdr.Name != "." {
		t.Log(o.Hdr.Name)
		t.Fail()
	}
	if o.Hdr.Rrtype != dns.TypeOPT {
		t.Log(o.Hdr.Name)
		t.Fail()
	}
	if o.Option[0].(*dns.EDNS0_SUBNET).SourceNetmask != sm {
		t.Log(o.Option[0].(*dns.EDNS0_SUBNET).SourceNetmask)
		t.Fail()
	}
	if o.Option[0].(*dns.EDNS0_SUBNET).SourceScope != sc {
		t.Log(o.Option[0].(*dns.EDNS0_SUBNET).SourceScope)
		t.Fail()
	}
}

func testQueryCNAME(t *testing.T, d string) {
	t.Log(d)
	cname_a, edns_h, edns, e := QueryCNAME(d, "202.106.0.20", []string{"114.114.114.114"}, "53")
	if e != nil {
		t.Log(e)
		t.Fail()
	}
	if cap(cname_a) < 1 {
		t.Log("... cap(cname_a) < 1")
		t.Fail()
	}
	t.Log(cname_a)
	for _, c := range cname_a {
		t.Log(reflect.TypeOf(c))
		t.Log(c.Hdr.Name)
		if c.Hdr.Name != dns.Fqdn(d) {
			t.Fail()
		}
		t.Log(c.Hdr.Rrtype)
		if c.Hdr.Rrtype != dns.TypeCNAME {
			t.Fail()
		}
		t.Log(c.Hdr.Ttl)
		if c.Hdr.Ttl > ttlmap[d] {
			t.Log(ttlmap[d])
			t.Fail()
		}
		t.Log(d)
		t.Log(c.Target)
		t.Log(dns.Fqdn(cnamemap[d]))
		if dns.Fqdn(c.Target) != dns.Fqdn(cnamemap[d]) {
			t.Fail()
		}
		t.Log("......................")
	}
	if edns_h != nil {
		t.Log("edns_h")
		t.Log(edns_h.Name)
		t.Log(edns_h.Rrtype)
	}
	if edns != nil {
		t.Log("edns")
		t.Log(edns)
	}
	t.Log("++++++++++++++++++++++++++++++")
}

func TestQueryCNAME(t *testing.T) {
	testQueryCNAME(t, "ww2.sinaimg.cn")
	testQueryCNAME(t, "www.baidu.com")
}

func testQueryNS(d string, t *testing.T) {
	n, e := QueryNS(d)
	if (cap(n) > 0) && (e == nil) {
		for _, x := range n {
			t.Log(x.Hdr.Name)
			t.Log(x.Hdr.Rrtype)
			t.Log(x.Hdr.Class)
			t.Log(x.Hdr.Rdlength)
			t.Log(x.Hdr.Ttl)
			t.Log(x.Ns)
		}
	} else {
		t.Log(d)
		t.Log(cap(n))
		t.Log(e)
		t.Fail()
	}
}

func TestQueryNS(t *testing.T) {
	testQueryNS("a.shifen.com", t)
	testQueryNS("sinaimg.cn", t)
	testQueryNS("sinaedge.com", t)
	testQueryNS("danuoyi.alicdn.com.", t)
}

func TestQueryA(t *testing.T) {
	testQueryA("www.a.shifen.com", t)
	testQueryA("www.baidu.com", t)
	testQueryA("ww2.sinaimg.cn", t)
	testQueryA("img.alicdn.com.danuoyi.alicdn.com", t)
	testQueryA("img.alicdn.com", t)
}

func testQueryA(d string, t *testing.T) {
	t.Log(d)
	a_a, edns_h, edns, e := QueryA(d, "201.106.0.20", []string{"114.114.114.114"}, "53")

	if e != nil {
		t.Log(e)
		t.Fail()
	} else {
		t.Log(a_a)

	}

	if cap(a_a) <= 0 {
		t.Log(reflect.TypeOf(a_a))
		t.Fail()
	} else {
		for _, v := range a_a {
			t.Log(reflect.TypeOf(v))
			t.Log(v)
			switch v.Header().Rrtype {
			case dns.TypeA:
				if vv, ok := v.(*dns.A); ok {
					t.Log(vv.Hdr.Name)
					t.Log(vv.Hdr.Ttl)
					t.Log(vv.Hdr.Rrtype)
					t.Log(vv.Hdr.Class)
					t.Log(vv.Hdr.Rdlength)
					t.Log(vv.A)

				} else {
					t.Log(vv)
					t.Fail()
				}
				t.Log("------------------")
			case dns.TypeCNAME:
				if cc, ok := v.(*dns.CNAME); ok {
					t.Log(cc.Hdr.Name)
					t.Log(cc.Hdr.Rrtype)
					t.Log(cc.Hdr.Class)
					t.Log(cc.Hdr.Ttl)
					t.Log(cc.Hdr.Rdlength)
					t.Log(cc.Target)

				} else {
					t.Log(cc)
					t.Fail()
				}
				t.Log("*********************")
			default:
				t.Log(v.String())
				t.Fail()
			}
			t.Log("======================")
		}
	}
	t.Log(reflect.TypeOf(edns_h))
	if edns_h != nil {
		t.Log(edns_h.Name)
		t.Log(edns_h.Rrtype)
		t.Log(edns_h.Ttl)
		t.Log(edns_h.Rdlength)
		t.Log(edns_h.Class)
	}
	t.Log(edns)
	t.Log("++++++++++++++++++++++++++++++++++++++++++")

}

func TestGenerateParentDomain(t *testing.T) {
	dsarray := []string{
		"www.yahoo.com",
		"www.google.com",
		"weboimg.gslb.sinaedge.com",
		"img.alicdn.com.danuoyi.alicdn.com.",
	}
	for _, ds := range dsarray {
		r, e := GenerateParentDomain(ds)
		t.Log(r)
		t.Log(e)
	}
}

func TestQuerySOA(t *testing.T) {
	dsmap := map[string]string{
		"baidu.com":                          "xxxx",
		"www.google.com":                     "abcdefg",
		"www.yahoo.com":                      "bbb",
		"yahoo.com":                          "zzzz",
		"weibo.cn":                           "xxxx",
		"www.baidu.com":                      "ns2.baidu.com.",
		"www.a.shifen.com":                   "ns1.a.shifen.com.",
		"a.shifen.com":                       "ns1.a.shifen.com.",
		"www2.sinaimg.cn":                    "ns1.sina.com.cn.",
		"weboimg.gslb.sinaedge.com":          "ns2.sinaedge.com.",
		"api.weibo.cn":                       "ns1.sina.com.cn.",
		"img.alicdn.com":                     "ns8.alibabaonline.com.",
		"alicdn.com":                         "yyyy",
		"img.alicdn.com.danuoyi.alicdn.com.": "danuoyinewns1.gds.alicdn.com.",
		"danuoyi.alicdn.com.":                "xxxxx",
		"fjdsljflsj.jfslj":                   "...",
	}
	for k,_ := range dsmap {
		t.Log("----------------------------------")
		t.Log(k)
		soa, ns_a, e := QuerySOA(k)
		if e == nil {
			t.Log(soa.Hdr.Name)
			t.Log(soa.Hdr.Rrtype)
			t.Log(soa.Hdr.Class)
			t.Log(soa.Ns)
			t.Log(soa.Expire)
			t.Log(soa.Mbox)
			t.Log(soa.Minttl)
			t.Log(soa.Retry)
		} else {
			t.Log(e)
		}
		t.Log(ns_a)
		t.Log("----------------------------------")
	}
}

func TestInitMySQL(t *testing.T) {
	if config.RC.MySQLEnabled {
		db := InitMySQL(RC_MySQLConf)
		if db != false {
			t.Log("InitMySQL OK")
		} else {
			t.Fail()
		}
	}
}

func TestGetDomainIDFromMySQL(t *testing.T) {
	d_a := []string{
		"www.sina.com.cn",
		"www.baidu.com",
		"www.a.shifen.com",
		"api.weibo.cn",
		"weibo.cn",
		"sinaedge.com",
		"ww2.sinaimg.cn",
	}

	if config.RC.MySQLEnabled {
		for _, d := range d_a {

			t.Log(d)
			id, e := RRMySQL.GetDomainIDFromMySQL(d)
			if e != nil {
				t.Log(id)
				t.Log(e)
				//			t.Fail()
			} else {
				t.Log(d)
				t.Log(id)
			}
		}
	}
}

func TestGetRegionWithIPFromMySQL(t *testing.T) {
	if config.RC.MySQLEnabled {
		// d_a := []string{"www.sina.com.cn", "www.baidu.com", "www.a.shifen.com", "api.weibo.cn", "weibo.cn", "sinaedge.com"}
		ipuint32 := uint32(1790519448)
		id, e := RRMySQL.GetRegionWithIPFromMySQL(ipuint32)
		if e == nil {
			t.Log(id.Region)
		} else {
			t.Log(e)
		}
	}
}

func TestGetRRFromMySQL(t *testing.T) {
	if config.RC.MySQLEnabled {
		t.Log("Test..")
		d_a := []uint32{1, 2, 6, 7, 9}
		r_a := []uint32{0, 1, 2, 6, 7, 8}
		for _, d := range d_a {
			for _, id := range r_a {
				x, e := RRMySQL.GetRRFromMySQL(d, id)
				if e == nil {
					t.Log("DomainId: ", d, " RegionId: ", id, "result:", x.idRR, x.RR)
					if x.RR.RrType == dns.TypeA {
						t.Log("A RR")
					} else if x.RR.RrType == dns.TypeCNAME {
						t.Log("CNAME RR")
					}
				} else {
					t.Log(e)
					if e.ErrorNo == MyError.ERROR_NORESULT {
					} else {
						t.Fail()
					}
				}
			}
		}
	}
}

//func BenchmarkGetRRfFromMySQL(b *testing.B) {
//	d_a := []uint32{1, 2, 6, 7, 9}
//	r_a := []uint32{0, 1, 2, 6, 7, 8}
//	b.ResetTimer()
//	for i := 0; i < 10; i++ {
//		for _, d := range d_a {
//			for _, id := range r_a {
//				x, e := RRMySQL.GetRRFromMySQL(d, id)
//				if e == nil {
//					b.Log("DomainId: ", d, " RegionId: ", id, "result:", x.idRR, x.RR)
//					//				for _, xx := range x {
//					//					t.Log(xx, xx.RR)
//					//					if xx.RR.RrType == dns.TypeA {
//					//						t.Log("A RR")
//					//					} else if xx.RR.RrType == dns.TypeCNAME {
//					//						t.Log("CNAME RR")
//					//					}
//					//				}
//				} else {
//					b.Log(e)
//					if e.ErrorNo == MyError.ERROR_NORESULT {
//					} else {
//						//						b.Fail()
//					}
//				}
//			}
//		}
//	}
//}

//func BenchmarkQuerySOA(b *testing.B) {
//	dsmap := map[string]string{
//		"baidu.com":                          "xxxx",
//		"www.yahoo.com":                      "bbb",
//		"yahoo.com":                          "zzzz",
//		"weibo.cn":                           "xxxx",
//		"www.baidu.com":                      "ns2.baidu.com.",
//		"www.a.shifen.com":                   "ns1.a.shifen.com.",
//		"a.shifen.com":                       "ns1.a.shifen.com.",
//		"www2.sinaimg.cn":                    "ns1.sina.com.cn.",
//		"weboimg.gslb.sinaedge.com":          "ns2.sinaedge.com.",
//		"api.weibo.cn":                       "ns1.sina.com.cn.",
//		"img.alicdn.com":                     "ns8.alibabaonline.com.",
//		"alicdn.com":                         "yyyy",
//		"img.alicdn.com.danuoyi.alicdn.com.": "danuoyinewns1.gds.alicdn.com.",
//		"danuoyi.alicdn.com.":                "xxxxx",
//		//		"fjdsljflsj.jfslj":                   "...",
//	}
//	for i := 0; i < b.N; i++ {
//		for k, _ := range dsmap {
//			//			b.Log("----------------------------------")
//			//			b.Log(k)
//			_, _, e := QuerySOA(k)
//			if e == nil {
//				//				b.Log(soa, ns_a)
//				//				t.Log(soa.Hdr.Name)
//				//				t.Log(soa.Hdr.Rrtype)
//				//				t.Log(soa.Hdr.Class)
//				//				t.Log(soa.Ns)
//				//				t.Log(soa.Expire)
//				//				t.Log(soa.Mbox)
//				//				t.Log(soa.Minttl)
//				//				t.Log(soa.Retry)
//			} else {
//				b.Log(e)
//			}
//		}
//	}
//	b.ReportAllocs()
//}

//func BenchmarkQueryCNAME(b *testing.B) {
//	for i := 0; i < b.N; i++ {
//		QueryCNAME("api.weibo.cn", "202.106.0.20", []string{"114.114.114.114"}, "53")
//	}
//	b.ReportAllocs()
//}

//func BenchmarkQueryA(b *testing.B) {
//	for i := 0; i < b.N; i++ {
//		QueryA("www.baidu.com", "201.106.0.20", []string{"114.114.114.114"}, "53")
//		//		b.Log(a_a, edns_h, edns, e)
//	}
//	b.ReportAllocs()
//}

//func BenchmarkParseA(b *testing.B) {
//	a, _, _, e := QueryA("www.a.shifen.com",
//		"202.106.0.20",
//		[]string{"ns1.a.shifen.com.",
//			"ns2.a.shifen.com.",
//			"ns3.a.shifen.com.",
//			"ns4.a.shifen.com.",
//			"ns5.a.shifen.com."}, "53")
//	b.ResetTimer()
//	if e != nil {
//		b.Fatal(e)
//	}
//	for n := 0; n < b.N; n++ {
//		_, ok := ParseA(a, "www.a.shifen.com")
//		if !ok {
//			b.Fatal("Parse error")
//		}
//	}
//}
