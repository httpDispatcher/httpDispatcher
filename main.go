package main

// param: lower case
// Struct unit: Upper Case
// Func: golang style

import (
	"fmt"
	"os"
	//	"flag"
	//	"fmt"
	//	"io"
	//	"log"
	//	"os"
	"log/syslog"
	"query"
)

func main() {
	loger, e := syslog.New(syslog.LOG_DEBUG, "httpDispacher")
	if e != nil {
		fmt.Println("Syslog Instance is error, exiting....")
		os.Exit(1)
	}
	loger.Debug("Starting httpDispacher...")

	qdc := new(query.Domain)
	qdc.AuthoritativeServers = append(qdc.AuthoritativeServers, "ns1.a.shifen.com.")
	qdc.Port = query.NS_SERVER_PORT
	//	query.Test("www.a.shifen.com.", "106.186.24.134", qdc)

	qdc.AuthoritativeServers = nil
	qdc.AuthoritativeServers = append(qdc.AuthoritativeServers, "ns1.baidu.com")
	//	qdc.SetAS("ns2.baidu.com")
	ns, e := query.QueryNS("www.baidu.com", qdc)
	fmt.Println(ns)

}
