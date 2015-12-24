package server

import (
	"testing"
	"time"
)

func TestGetSOARecord(t *testing.T) {
	soa, e := GetSOARecord("www.a.shifen.com")
	t.Log(soa)
	t.Log(e)
	time.Sleep(10 * time.Second)
	soa, e = GetSOARecord("www.a.shifen.com")
	t.Log(soa)
	t.Log(e)
}

func TestGetARecord(t *testing.T) {
	d_arr := []string{
		"www.taobao.com",
		"www.baidu.com",
		"www.sina.com.cn",
		"api.weibo.cn",
		"weibo.cn",
		//		"ww2.sinaimg.cn",
	}
	for _, d := range d_arr {
		t.Log("Query for: ", d, " with clientip ", GetClientIP())
		ok, a, e := GetARecord(d, GetClientIP())
		t.Log(ok, a, e)
	}

}
func BenchmarkGetARecord(b *testing.B) {
	d_arr := []string{
		"www.taobao.com",
		"www.baidu.com",
		"www.sina.com.cn",
		"api.weibo.cn",
		"weibo.cn",
		//		"ww2.sinaimg.cn",
	}
	for _, d := range d_arr {
		GetARecord(d, GetClientIP())
	}
	b.StartTimer()
	for i := 0; i < 1000; i++ {
		for _, d := range d_arr {
			GetARecord(d, GetClientIP())
		}
	}
	b.StopTimer()
	b.ReportAllocs()
}
