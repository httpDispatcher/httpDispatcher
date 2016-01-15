package server

import (
	"config"
	"domain"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"utils"

	"strings"
	"time"

	"github.com/miekg/dns"
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
	utils.ServerLogger.Debug("Request header: %s  RequestURI: %s  URI: %s", r.Header, r.RequestURI, r.URL)
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
	if _, ok := dns.IsDomainName(query_domain); !ok {
		fmt.Fprintln(w, "Error domain name: ", query_domain)
		utils.ServerLogger.Info("error domain name : %s ", query_domain)
		return
	}

	if srcIP == "" {
		hp := strings.Split(r.RemoteAddr, ":")
		srcIP = hp[0]
	}
	utils.QueryLogger.Info("src ip: %s query_domain: %s url_path: %s", string(srcIP), query_domain, url_path)
	if x := net.ParseIP(srcIP); x == nil {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, srcIP)
		fmt.Fprintln(w, "src ip : "+srcIP+" is not correct")
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
					utils.ServerLogger.Debug("query result: %s ", a.A.String())
				} else {
					fmt.Fprintln(w, ree.String())
					utils.ServerLogger.Debug("query result: %s ", ree.String())
				}
			}
		} else if e != nil {
			fmt.Fprintln(w, e.Error())
			utils.ServerLogger.Error("query domain: %s src_ip: %s  %s", query_domain, srcIP, e.Error())
		} else {
			fmt.Fprintln(w, "unkown error!\n")
			utils.ServerLogger.Error("query domain: %s src_ip: %s fail unkown error!", query_domain, srcIP)
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintln(w, "Query for domain: "+query_domain+" is not permited\n")
		utils.ServerLogger.Info("Query for domain: %s is not permited", query_domain)
		return
	}
}

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
		utils.ServerLogger.Critical("Create listener error: %s", err.Error())
		os.Exit(1)
	}
	if err := server.Serve(listener); nil != err {
		utils.ServerLogger.Critical("Call server error: %s", err.Error())
		os.Exit(1)
	}

}

func checkServeAddr(addr string) error {
	if ip := net.ParseIP(addr); ip != nil {
		return nil
	}
	return errors.New("addr error")
}
