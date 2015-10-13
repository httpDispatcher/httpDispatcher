package main

// param: lower case + Upper Case ,No _ spliter
// Struct unit: Upper Case
// Func: golang style

import (
	"fmt"
	"query"
)

func main() {
	//	utils.InitUitls()

	ns, _ := query.QueryNS("baidu.com")
	ns_servers, ttl, err := query.ParseNS(ns)
	fmt.Println(ns_servers)
	fmt.Println(ttl)
	fmt.Println(err)
	fmt.Println("----------------------------------")
	ns, _ = query.QueryNS("weibo.cn")
	ns_servers, ttl, err = query.ParseNS(ns)
	fmt.Println(ns_servers)
	fmt.Println(ttl)
	fmt.Println("----------------------------------")
	ns, _ = query.QueryNS("sinaedge.com.")
	ns_servers, ttl, err = query.ParseNS(ns)
	fmt.Println(ns_servers)
	fmt.Println(ttl)
	fmt.Println("----------------------------------")
	ns, _ = query.QueryNS("grid.sinaedge.com.")
	ns_servers, ttl, err = query.ParseNS(ns)
	fmt.Println(ns_servers)
	fmt.Println(ttl)
	fmt.Println("----------------------------------")
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

}
