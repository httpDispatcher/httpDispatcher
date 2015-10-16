package main

import (
	"fmt"
	"query"
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

	query.QueryCNAME("ww2.sinaimg.cn.", true)
	fmt.Println("----------------------------------")
	query.QueryCNAME("weiboimg.gslb.sinaedge.com.", true)
	fmt.Println("----------------------------------")
	query.QueryCNAME("weiboimg.grid.sinaedge.com.", true)
	fmt.Println("----------------------------------")
	query.QueryCNAME("www.baidu.com", true)

	d := "weiboimg.gslb.sinaedge.com."
	r, _ := query.LoopForQueryNS(d)
	fmt.Println(r)

	d = "api.weibo.cn"
	r, _ = query.LoopForQueryNS(d)
	fmt.Println(r)

	d = "a.b.c.d.e.f.a.a.a.d.d.d.d.d.d.d.d.d.d.d.d.d.com"
	r, e := query.LoopForQueryNS(d)
	fmt.Println(r)
	fmt.Println(e)
}
