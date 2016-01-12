package server

import (
	//	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	//	"path"
	"config"
	"domain"

	"github.com/miekg/dns"
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
	fmt.Println(r.Header)
	fmt.Println(r.RequestURI)
	fmt.Println(r.URL)
}

func RegionTraverServe(w http.ResponseWriter, r *http.Request) {
	url_path := r.URL.Path
	query_string := r.URL.Query().Get("d")

	fmt.Println(query_string)
	fmt.Println(url_path)
	w.Write([]byte(r.RemoteAddr))
	w.Write([]byte("\n"))

	fmt.Println(r.Header)
	fmt.Println(r.RequestURI)
	fmt.Println(r.URL)
	t, e := domain.DomainRRCache.GetDomainNodeFromCacheWithName(query_string)
	if e == nil {
		t.DomainRegionTree.TraverseRegionTree()
	} else {
		w.Write([]byte(e.Error()))
	}

}

func HttpQueryServe(w http.ResponseWriter, r *http.Request) {
	url_path := r.URL.Path
	query_domain := r.URL.Query().Get("d")
	clientip := r.URL.Query().Get("ip")
	fmt.Println("client ip: ", clientip)

	fmt.Println("query_domain: ", query_domain)
	fmt.Println("url_path: ", url_path)

	if clientip == "" {
		clientip = r.RemoteAddr
	}
	if x := net.ParseIP(clientip); x == nil {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(clientip))
		w.Write([]byte("client ip : " + clientip + " is not correct\n"))
		return
	}

	if config.InWhiteList(query_domain) {
		ok, re, e := GetARecord(query_domain, clientip)
		if ok {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			for _, ree := range re {
				if a, ok := ree.(*dns.A); ok {
					fmt.Fprintln(w, a.A.String())
				} else {
					fmt.Fprintln(w, ree.String())
				}
			}
		} else if e != nil {
			w.Write([]byte(e.Msg))
		} else {
			w.Write([]byte("unkown error!\n"))
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Query for domain: " + query_domain + " is not permited\n"))
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
	//	s := &http.Server{
	//		Addr:           ":8080",
	//		Handler:        TmpServe,
	//		ReadTimeout:    10 * time.Second,
	//		WriteTimeout:   10 * time.Second,
	//		MaxHeaderBytes: 1 << 20,
	//	}
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
		log.Fatalln(err)
	}
	if err := server.Serve(listener); nil != err {
		log.Fatalln(err)
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
