package iplookup

import (
	"reflect"
	"testing"
	"utils"
)

const DBPATH = "../../data/ip.db.db"

var IPS = []string{"106.185.48.28", "202.106.0.20", "192.168.1.1", "124.207.129.171"}

func TestOpenIPDB(t *testing.T) {
	ok := Il_open(DBPATH)
	if ok >= 0 {
		t.Log(ok)
		defer Il_close(ok)
	} else {
		t.Log(reflect.TypeOf(ok))
		t.Log(ok)
		t.Fail()
	}
}

func TestStringToIP(t *testing.T) {
	for _, i := range IPS {
		t.Log(i)
		ip := NewIp(i)
		t.Log(ip.GetIn())
	}
}

func TestIlSearch(t *testing.T) {

	ipinfo := NewIpinfo()
	defer DeleteIpinfo(ipinfo)
	for _, ip := range IPS {

		if ok := Il_open(DBPATH); ok > 0 {
			defer Il_close(ok)
			nip := NewIp(ip)
			defer DeleteIp(nip)
			t.Log(nip.GetIn())
			n := Il_search(nip, ipinfo, ok)
			if n > 0 {
				t.Log(n)
				//				t.Log(ipinfo.GetStart(), ipinfo.GetEnd())
				//				t.Log(ipinfo.GetStart().GetIn(), ipinfo.GetEnd().GetIn())
				//				t.Log(reflect.ValueOf(ipinfo.GetStart().GetIn()), reflect.ValueOf(ipinfo.GetEnd().GetIn()))

				//				GetIpValid(ipinfo.GetStart())
				n1, n2, n3, n4 := GetIpinfoStartEnd(ipinfo)
				t.Log(n1, n2, n3, n4)
				m := NewIpitem()
				x := Il_bin2human(ipinfo, m, Text)
				t.Log(m, x, x.GetStart(), x.GetEnd())

				t.Log(utils.Ip4ToInt32(utils.StrToIP(x.GetStart())), utils.Ip4ToInt32(utils.StrToIP(x.GetEnd())))

			} else {
				t.Log(n)
				t.Log(ipinfo.GetStart(), ipinfo.GetEnd())
				if ip != "192.168.1.1" {
					t.Fail()
				}
			}
		}
	}
}

func TestGetIPinfoWithString(t *testing.T) {
	for _, ip := range IPS {
		ipinfo, e := GetIPinfoWithString(ip)
		if e == nil && ipinfo != nil {
			t.Log(ipinfo.GetStart(), ipinfo.GetEnd())
			a, b, c, d := GetIpinfoStartEnd(ipinfo)
			t.Log(ip, a, b, c, d)
		} else {
			t.Log(ip, ipinfo, e)
		}
	}
}

func TestGetIpinfoStartEndWithIPString(t *testing.T) {
	for _, ip := range IPS {
		x, y := GetIpinfoStartEndWithIPString(ip)
		t.Log(ip, x, y)
	}
}
