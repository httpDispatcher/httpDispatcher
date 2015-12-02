package main

import (
	//	"fmt"
	//	"query"
	"server"
	"utils"
)

// param: lower case + Upper Case ,No _ spliter
// Struct unit: Upper Case
// Func: golang style

func main() {
	//	utils.InitUitls()

	//	ns, _ := query.QueryNS("baidu.com")
	//	ns_servers, ttl, err := query.ParseNS(ns)
	//	fmt.Println(ns_servers)
	//	fmt.Println(ttl)
	//	fmt.Println(err)
	//	fmt.Println("----------------------------------")
	//	ns, _ = query.QueryNS("weibo.cn")
	//	ns_servers, ttl, err = query.ParseNS(ns)
	//	fmt.Println(ns_servers)
	//	fmt.Println(ttl)
	//	fmt.Println("----------------------------------")
	//	ns, _ = query.QueryNS("sinaedge.com.")
	//	ns_servers, ttl, err = query.ParseNS(ns)
	//	fmt.Println(ns_servers)
	//	fmt.Println(ttl)
	//	fmt.Println("----------------------------------")
	//	ns, _ = query.QueryNS("grid.sinaedge.com.")
	//	ns_servers, ttl, err = query.ParseNS(ns)
	//	fmt.Println(ns_servers)
	//	fmt.Println(ttl)
	//	fmt.Println("----------------------------------")
	//	a, edns0, _ := query.QueryA("api.weibo.cn.", true)
	//	a_rr, ttl, _ := query.ParseA(a)
	//	fmt.Println(a_rr)
	//	fmt.Println(ttl)
	//	fmt.Println(edns0)
	//
	//
	//	query.InitDomainDB()
	//	bd_ns := []string{"ns1.baidu.com", "ns2.baidu.com"}
	//	dt, _ := query.NewDomain("www.baidu.com", bd_ns, 86400)
	////	sina_ns := []string{"ns1.sina.com", "ns2.sina.com"}
	////	ds, _ := query.NewDomain("api.weibo.cn", sina_ns, 86400)
	////	//	dx, _ := query.NewDomain("www.baidu.com", nil, 0)
	////	query.DomainDB.Insert(dt)
	////
	////	query.DomainDB.Insert(ds)
	//	//	if x := query.DomainDB.Get(dt); x == nil {
	//
	//	fmt.Println(query.DomainDB.Len())
	//	x := query.DomainDB.Get(dt)
	//	//		if x != nil {
	//	fmt.Println(x.(*query.Domain).Ns)
	//	fmt.Println(x.(*query.Domain).Ttl)
	//	dq, _ := query.NewDomain("api.weibo.cn", nil, 0)
	//	y := query.DomainDB.Get(dq)
	//	fmt.Println(y.(*query.Domain).Ns)
	//	fmt.Println(y.(*query.Domain).Ttl)
	//
	//	//		}
	//	//	} else {
	//	//		fmt.Println(query.DomainDB.Len())
	//	//	}
	//
	//	query.QueryCNAME("ww2.sinaimg.cn.", true)
	//	fmt.Println("----------------------------------")
	//	query.QueryCNAME("weiboimg.gslb.sinaedge.com.", true)
	//	fmt.Println("----------------------------------")
	//	query.QueryCNAME("weiboimg.grid.sinaedge.com.", true)
	//	fmt.Println("----------------------------------")
	//	query.QueryCNAME("www.baidu.com", true)
	//
	//	d := "weiboimg.gslb.sinaedge.com."
	//	r, _ := query.LoopForQueryNS(d)
	//	fmt.Println(r)
	//
	//	d = "api.weibo.cn"
	//	r, _ = query.LoopForQueryNS(d)
	//	fmt.Println(r)
	//
	//	d = "a.b.c.d.e.f.a.a.a.d.d.d.d.d.d.d.d.d.d.d.d.d.com"
	//	r, e := query.LoopForQueryNS(d)
	//	fmt.Println(r)
	//	fmt.Println(e)
	ServerAddr := "127.0.0.1"
	ServerPort := int32(8080)
	utils.InitUitls()
	//	utils.Logger.Println(ServerAddr + ":" + ServerPort)
	server.NewServer(ServerAddr, ServerPort)

}
