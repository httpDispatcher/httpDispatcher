package server

import (
	//	"encoding/json"
	"errors"
	"fmt"
	//"log"
	"net"
	"os"
	"net/http"
	//	"path"
	"config"
	"domain"
	"utils"

	"github.com/miekg/dns"
	"strings"
	"time"
)

type myHandler struct {
}

type HttpDnsClient struct {
	ClientAddr string
	AuthToken  string
	Identifier string
	Writer     http.ResponseWriter
	Request    *http.Request
}

func (s *myHandler) ServerHttp(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(r.RemoteAddr))
        utils.ServerLogger.Debug( "Request header: %s  RequestURI: %s  URI: %s", r.Header, r.RequestURI, r.URL)
	//fmt.Println(r.Header)
	//fmt.Println(r.RequestURI)
	//fmt.Println(r.URL)
}

func RegionTraverServe(w http.ResponseWriter, r *http.Request) {
	url_path := r.URL.Path
	query_string := r.URL.Query().Get("d")

        utils.QueryLogger.Info("query_domain: ", query_string, " url_path: ", url_path)
	//fmt.Println(query_string)
	//fmt.Println(url_path)
	w.Write([]byte(r.RemoteAddr))
	w.Write([]byte("\n"))

	fmt.Fprintln(w, r.Header)
	fmt.Fprintln(w, r.RequestURI)
	//fmt.Println(w, r.URL)
	t, e := domain.DomainRRCache.GetDomainNodeFromCacheWithName(query_string)
	if e == nil {
		t.DomainRegionTree.TraverseRegionTree()
	} else {
		w.Write([]byte(e.Error()))
                utils.ServerLogger.Error("query_domain: %s  url_path: %s is error: %s", query_string, url_path, e.Error())
	}

}

func HttpQueryServe(w http.ResponseWriter, r *http.Request) {
	url_path := r.URL.Path
	query_domain := r.URL.Query().Get("d")
	srcIP := r.URL.Query().Get("ip")
	//fmt.Println("src ip: ", srcIP)

	//fmt.Println("query_domain: ", query_domain)
	//fmt.Println("url_path: ", url_path)

	if srcIP == "" {
		hp := strings.Split(r.RemoteAddr, ":")
		srcIP = hp[0]
	}
        //utils.ServerLogger.Info(true, "src ip: ", string(srcIP), " query_domain: ", query_domain, " url_path: ", url_path)
        utils.QueryLogger.Info("src ip: %s query_domain: %s url_path: %s", string(srcIP), query_domain, url_path)
	if x := net.ParseIP(srcIP); x == nil {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(srcIP))
		w.Write([]byte("src ip : " + srcIP + " is not correct\n"))
                utils.ServerLogger.Warning("src ip : %s is not correct", srcIP)
		return
	}

	if config.InWhiteList(query_domain) {
		ok, re, e := GetARecord(query_domain, srcIP)
		if ok {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			for _, ree := range re {
				if a, ok := ree.(*dns.A); ok {
					fmt.Fprintln(w, a.A.String())
		                        utils.ServerLogger.Debug( "query result: %s ", a.A.String())
				} else {
					fmt.Fprintln(w, ree.String())
		                        utils.ServerLogger.Debug( "query result: %s ", ree.String())
				}
			}
		} else if e != nil {
			w.Write([]byte(e.Msg))
		        utils.ServerLogger.Error("query domain: %s src_ip: %s  %s", query_domain, srcIP, e.Error())
		} else {
			w.Write([]byte("unkown error!\n"))
		        utils.ServerLogger.Error("query domain: %s src_ip: %s fail unkown error!", query_domain, srcIP)
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Query for domain: " + query_domain + " is not permited\n"))
		utils.ServerLogger.Info("Query for domain: %s is not permited", query_domain)
		//		runtime.Goexit()          
	}
}

func ParseDomain(d string) (int, bool) {
	return dns.IsDomainName(d)
}

//func ServeDomain(d string) bool {
//	if _, err := ParseDomain(d); err != true {
//		utils.Logger.Println("Param error: " + d + " is not a valid domain name ")
//		return false
//	}
//	//	[]dns.RR, *dns.EDNS0_SUBNET, error
//	rr, subnet, err := query.QueryA(d, true)
//	fmt.Println(rr)
//	fmt.Println(subnet)
//	fmt.Println(err)
//	return true
//}

func NewServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/q", HttpQueryServe)
	mux.HandleFunc("/t", RegionTraverServe)
	server := &http.Server{
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		Handler:      mux,
	}
	listener, err := net.Listen("tcp", config.RC.Bind)
	defer listener.Close()
	if nil != err {
		//log.Fatalln(err)
                utils.ServerLogger.Critical("Create listener error: %s", err.Error())
                os.Exit(1)
	}
	if err := server.Serve(listener); nil != err {
                utils.ServerLogger.Critical("Call server error: %s", err.Error())
                os.Exit(1)
		//log.Fatalln(err)
	}

}

func GetInterfaceAddr() ([]net.Addr, error) {
	return net.InterfaceAddrs()
}

func checkServeAddr(addr string) error {
	if ip := net.ParseIP(addr); ip != nil {
		return nil
	}
	return errors.New("addr error")
}
