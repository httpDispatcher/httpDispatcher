package query

import (
	"reflect"
	"testing"

	"github.com/miekg/dns"
)

var cnamemap = map[string]string{
	"ww2.sinaimg.cn":             "weiboimg.gslb.sinaedge.com.",
	"www.baidu.com":              "www.a.shifen.com.",
	"weiboimg.gslb.sinaedge.com": "weiboimg.grid.sinaedge.com.",
}
var ttlmap = map[string]uint32{
	"ww2.sinaimg.cn":             uint32(60),
	"www.baidu.com":              uint32(1200),
	"weiboimg.gslb.sinaedge.com": uint32(600),
}

//func Test_preQuery(t *testing.T) {
//	d := "ww2.sinaimg.cn."
//	ds,dp ,o, e := preQuery(d, false)
//	if e != nil {
//		t.Fail()
//	}
//	if ds != "ns1.sina.com.cn." {
//		t.Fail()
//	}
//	if dp != "53" {
//		t.Fail()
//	}
//	if o != nil {
//		t.Fail()
//	}
//
//	d = "www.baidu.com"
//	ds, dp, o, e = preQuery(d, true)
//	if e != nil {
//		t.Fail()
//	}
//	if ds != "ns2.baidu.com." {
//		t.Fail()
//	}
//	if dp != "53" {
//		t.Fail()
//	}
//	//	t.Log(e)
//	//	t.Log(ds)
//	//	t.Log(dp)
//	//	t.Log(o)
//	if o.Hdr.Name != "." {
//		t.Log(o.Hdr.Name)
//		t.Log("a")
//		t.Fail()
//	}
//	if o.Option[0].(*dns.EDNS0_SUBNET).Code != dns.EDNS0SUBNET {
//		t.Log("b")
//		t.Fail()
//	}
//
//	d = "www.a.shifen.com"
//	ds, dp, o, e = preQuery(d, true)
//	if e != nil {
//		t.Log(e.Msg)
//		t.Fail()
//	}
//	if o.Hdr.Rrtype != dns.TypeOPT {
//		t.Log(o)
//		t.Fail()
//	}
//	if cap(o.Option) != 1 {
//		t.Log(o.Option)
//		t.Fail()
//	}
//	x := o.Option[0].(*dns.EDNS0_SUBNET)
//	if x == nil {
//		t.Log(o.Option)
//		t.Fail()
//	} else {
//		//		fmt.Println(x.Address)
//		//		fmt.Println(x.Code)
//		//		fmt.Println(x.SourceNetmask)
//		//		fmt.Println(x.SourceScope)
//	}
//}

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
	cname_a, edns_h, edns, e := QueryCNAME(d, true, "124.124.124.124", "53")
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
		t.Log(edns_h)
		t.Log(edns_h.(dns.RR_Header).Rrtype)
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
	testQueryCNAME(t, "weiboimg.gslb.sinaedge.com")
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
	testQueryA("img.alicdn.com", t)
	testQueryNS("sinaimg.cn", t)
	testQueryNS("sinaedge.com", t)
	testQueryNS("danuoyi.alicdn.com.", t)
}

func testLoopforqueryns(d string, t *testing.T) {
	ns_a, e := LoopForQueryNS(d)
	t.Log(cap(ns_a))
	t.Log((e == nil) && (cap(ns_a) > 0))
	if (e == nil) && (cap(ns_a) > 0) {
		for _, ns := range ns_a {
			t.Log(reflect.TypeOf(ns))
			t.Log(ns.Hdr.Name)
			t.Log(ns.Hdr.Rrtype)
			t.Log(ns.Hdr.Class)
			t.Log(ns.Hdr.Rdlength)
			t.Log(ns.Hdr.Ttl)
			t.Log(ns.Ns)
		}
	} else {
		t.Log(ns_a)
		t.Log(e)
		t.Fail()
	}
}

func TestLoopForQueryNS(t *testing.T) {
	testLoopforqueryns("weiboimg.gslb.sinaedge.com", t)
	testLoopforqueryns("ww2.sinaimg.cn", t)
	testLoopforqueryns("api.weibo.cn", t)
	//	testLoopforqueryns("weiboimg.",t)
	testLoopforqueryns("img.alicdn.com.danuoyi.alicdn.com.", t)
	testLoopforqueryns("www.a.shifen.com", t)
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
	a_a, edns_h, edns, e := QueryA(d, true, "124.124.124.124", "53")

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
	if edns_header, ok := edns_h.(dns.RR_Header); ok {
		t.Log(edns_header.Name)
		t.Log(edns_header.Rrtype)
		t.Log(edns_header.Ttl)
		t.Log(edns_header.Rdlength)
		t.Log(edns_header.Class)
	} else {
		t.Log("edns_h is not dns.RR_Header type")
		t.Log(reflect.TypeOf(edns_h))
		//		t.Fail()
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
	for k, _ := range dsmap {
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
