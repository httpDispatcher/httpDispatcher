package main

import (
	//	"flag"
	//	"fmt"
	//	"io"
	//	"log"
	//	"os"
	"query"
)

func main() {
	//	logger := log.New(os.Stderr, "_httpDispather_", log.Lshortfile)
	qdc := new(query.DomainConfig)
	qdc.AuthoritativeServers = append(qdc.AuthoritativeServers, "8.8.8.8")
	qdc.Port = query.NS_SERVER_PORT
	query.Test("www.google.com", "106.186.24.134", qdc)

}
