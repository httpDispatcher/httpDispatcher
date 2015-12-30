package query

import (
	"github.com/miekg/dns"
	"testing"
)

func TestInitMySQL(t *testing.T) {
	db := InitMySQL()
	if db != false {
		t.Log("InitMySQL OK")
	} else {
		t.Fail()
	}
}

func TestGetDomainIDFromMySQL(t *testing.T) {
	d_a := []string{"www.sina.com.cn", "www.baidu.com", "www.a.shifen.com", "api.weibo.cn", "weibo.cn", "sinaedge.com"}

	for _, d := range d_a {
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

func TestGetDomainRegionWithIPFromMySQL(t *testing.T) {
	d_a := []string{"www.sina.com.cn", "www.baidu.com", "www.a.shifen.com", "api.weibo.cn", "weibo.cn", "sinaedge.com"}
	ipuint32 := uint32(10245)
	for _, d := range d_a {
		id, e := RRMySQL.GetDomainRegionWithIPFromMySQL(d, ipuint32)
		if e == nil {
			t.Log(id.idRegionID, id.Region)
		} else {
			t.Log(e)
		}
	}
}

func TestGetRRFromMySQL(t *testing.T) {
	r_a := []uint32{1, 2, 3, 4, 5}
	for _, id := range r_a {
		x, e := RRMySQL.GetRRFromMySQL(id)
		if e == nil {
			for _, xx := range x {
				t.Log(xx, xx.RR)
				if xx.RR.RrType == dns.TypeA {
					t.Log("A RR")
				} else if xx.RR.RrType == dns.TypeCNAME {
					t.Log("CNAME RR")
				}
			}
		} else {
			t.Log(e)
			t.Fail()
		}
	}
}
