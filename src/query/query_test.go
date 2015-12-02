package query

import (
	"fmt"
	"github.com/miekg/dns"
	"reflect"
	"testing"
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

func Test_preQuery(t *testing.T) {
	d := "ww2.sinaimg.cn."
	e, ds, dp, o := preQuery(d, false)
	if e != nil {
		t.Fatal()
	}
	if ds != "ns1.sina.com.cn." {
		t.Fatal()
	}
	if dp != "53" {
		t.Fatal()
	}
	if o != nil {
		t.Fatal()
	}

	d = "www.baidu.com"
	e, ds, dp, o = preQuery(d, true)
	if e != nil {
		t.Fatal()
	}
	if ds != "ns2.baidu.com." {
		t.Fatal()
	}
	if dp != "53" {
		t.Fatal()
	}
	//	t.Log(e)
	//	t.Log(ds)
	//	t.Log(dp)
	//	t.Log(o)
	if o.Hdr.Name != "." {
		t.Log(o.Hdr.Name)
		t.Log("a")
		t.Fatal()
	}
	if o.Option[0].(*dns.EDNS0_SUBNET).Code != dns.EDNS0SUBNET {
		t.Log("b")
		t.Fatal()
	}

	d = "www.a.shifen.com"
	e, ds, dp, o = preQuery(d, true)
	if e != nil {
		t.Log(e.Msg)
		t.Fatal()
	}
	if o.Hdr.Rrtype != dns.TypeOPT {
		t.Log(o)
		t.Fatal()
	}
	if cap(o.Option) != 1 {
		t.Log(o.Option)
		t.Fatal()
	}
	x := o.Option[0].(*dns.EDNS0_SUBNET)
	if x == nil {
		t.Log(o.Option)
		t.Fatal()
	} else {
		//		fmt.Println(x.Address)
		//		fmt.Println(x.Code)
		//		fmt.Println(x.SourceNetmask)
		//		fmt.Println(x.SourceScope)
	}
}

func TestPackEdns0SubnetOPT(t *testing.T) {
	var sm = uint8(32)
	var sc = uint8(0)
	var ip = "124.207.129.171"
	o := PackEdns0SubnetOPT(ip, sm, sc)
	fmt.Println(o)
	if o.Hdr.Name != "." {
		t.Log(o.Hdr.Name)
		t.Fatal()
	}
	if o.Hdr.Rrtype != dns.TypeOPT {
		t.Log(o.Hdr.Name)
		t.Fatal()
	}
	if o.Option[0].(*dns.EDNS0_SUBNET).SourceNetmask != sm {
		t.Log(o.Option[0].(*dns.EDNS0_SUBNET).SourceNetmask)
		t.Fatal()
	}
	if o.Option[0].(*dns.EDNS0_SUBNET).SourceScope != sc {
		t.Log(o.Option[0].(*dns.EDNS0_SUBNET).SourceScope)
		t.Fatal()
	}
}

func testQueryCNAME(t *testing.T, d string) {
	e, cname_a, edns_h, edns_a := QueryCNAME(d, true)
	if e != nil {
		t.Log(e)
		t.Fatal()
	}
	if cap(cname_a) < 1 {
		t.Log("... cap(cname_a) < 1")
		t.Fatal()
	}
	t.Log(cname_a)
	for _, c := range cname_a {
		t.Log(reflect.TypeOf(c))
		t.Log(c.Hdr.Name)
		if c.Hdr.Name != dns.Fqdn(d) {
			t.Fatal()
		}
		t.Log("......")
		t.Log(c.Hdr.Rrtype)
		if c.Hdr.Rrtype != dns.TypeCNAME {
			t.Fatal()
		}
		t.Log("......")
		t.Log(c.Hdr.Ttl)
		if c.Hdr.Ttl > ttlmap[d] {
			t.Log(ttlmap[d])
			t.Fatal()
		}
		t.Log("......")
		t.Log(d)
		t.Log(c.Target)
		t.Log(dns.Fqdn(cnamemap[d]))
		if dns.Fqdn(c.Target) != dns.Fqdn(cnamemap[d]) {
			t.Fatal()
		}

	}
	if edns_h != nil {
		t.Log("edns_h")
		t.Log(edns_h)
		t.Log(edns_h.(dns.RR_Header).Rrtype)
	}
	if cap(edns_a) > 0 {
		t.Log("edns_a")
		t.Log(edns_a)
	}
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
		t.Log(e)
		t.Fatal()
	}
}

func TestQueryNS(t *testing.T) {
	testQueryNS("sinaimg.cn", t)
	testQueryNS("sinaedge.com", t)
}

func TestLoopForQueryNS(t *testing.T) {
	ns_a, e := LoopForQueryNS("weiboimg.gslb.sinaedge.com")
	t.Log(ns_a)
	t.Log(e)
}
