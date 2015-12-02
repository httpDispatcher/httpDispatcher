package server

import (
	//	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	//	"path"
	"github.com/miekg/dns"
	"query"
	"utils"
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

func TmpServe(w http.ResponseWriter, r *http.Request) {
	url_path := r.URL.Path
	query_string := r.URL.Query().Get("d")

	fmt.Println(query_string)
	fmt.Println(url_path)
	w.Write([]byte(r.RemoteAddr))
	fmt.Println(r.Header)
	fmt.Println(r.RequestURI)
	fmt.Println(r.URL)
	//	b := r.Body
}

//func GetClientAddr(r *http.Request)

func ParseDomain(d string) (int, bool) {
	return dns.IsDomainName(d)
}

func ServeDomain(d string) bool {
	if _, err := ParseDomain(d); err != true {
		utils.Logger.Println("Param error: " + d + " is not a valid domain name ")
		return false
	}
	//	[]dns.RR, *dns.EDNS0_SUBNET, error
	rr, subnet, err := query.QueryA(d, true)
	fmt.Println(rr)
	fmt.Println(subnet)
	fmt.Println(err)
	return true
}

func NewServer(addr string, port int32) {
	if err := checkServeAddr(addr); err != nil {
		fmt.Println(err)
		panic("Server addr error ! :" + addr)
	}
	//	s := &http.Server{
	//		Addr:           ":8080",
	//		Handler:        TmpServe,
	//		ReadTimeout:    10 * time.Second,
	//		WriteTimeout:   10 * time.Second,
	//		MaxHeaderBytes: 1 << 20,
	//	}
	http.HandleFunc("/", TmpServe)
	log.Fatal(http.ListenAndServe(":8080", nil))
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
