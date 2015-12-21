package domain

import (
	"MyError"
	"fmt"
	"net"
	"os"
	"query"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/miekg/bitradix"
	"github.com/miekg/dns"
	"github.com/petar/GoLLRB/llrb"
)

const radix_bit = 5

type MuLLRB struct {
	llrb.LLRB
	sync.Mutex
}

type MubitRadix struct {
	Radix32 *bitradix.Radix32
	sync.Mutex
}

//For domain name and Region RR
type DomainRRTree MuLLRB

//For domain SOA and NS record
type DomainSOATree MuLLRB

//For domain Region and A/CNAME record
type RegionTree MubitRadix

//TODO: redundant data types, need to be redesign
type Domain struct {
	DomainName string
	SOAKey     string // Use this key to search DomainSOANode in DomainSOATree,to find out the NS record
	TTL        uint32
}

type DomainNode struct {
	Domain
	DomainRegionTree *RegionTree
}

//TODO: redundant data types, need to be redesign
// dns.RR && RrType && TTL
type Region struct {
	NetWorkAddr uint32
	IpStart     uint32
	IpEnd       uint32
	RR          []dns.RR
	RrType      uint16
	TTL         uint32
	UpdateTime  time.Time
}

type DomainSOANode struct {
	SOAKey string // store SOA record first field,not the full domain name,but only the "dig -t SOA domain" resoponse
	NS     []*dns.NS
	SOA    *dns.SOA
}

type DomainConfig struct {
	DomainName           string
	AuthoritativeServers []string
	Port                 string
	Ttl                  string
}

var once sync.Once

//DomainRRCache for Domain A/CNAME record
var DomainRRCache *DomainRRTree

//DomainSOACache for Domain soa/cname record
var DomainSOACache *DomainSOATree

//if you want to search a A/CNAME record for a domain 'domainname',you should:
//First: search a DomainNode in domainRRCache and get domain region tree
// with DomainNode.RegionTree, and than get the record in the region tree
//Second: if there is no DomainNode in DomainRRCache, you should get DomainSOANode in
// DomainSOACache and get NS with DomainSOANode.NS. Notice that ,DomainSOANode.NS is
// a slice of *dns.NS.
//Third: Use query.QueryA with the one name server in DomainSOANode.NS.
//Notice: you should store all the infoformation when it is not in the trees(
// Both DomainSOACache and DomainRRCache )

func init() {
	errCache := InitCache()

	if errCache == nil {
		fmt.Println("InitDomainRRCache OK")
		fmt.Println("InitDomainSOACache OK")
	} else {
		fmt.Println("InitDomainRRCache() or InitDomainSOACache() failed")
		os.Exit(2)
	}

}

func InitCache() *MyError.MyError {
	once.Do(func() {
		DomainRRCache = &DomainRRTree{}
		DomainSOACache = &DomainSOATree{}
	})
	return nil
}

func (a *DomainNode) Less(b llrb.Item) bool {
	if x, ok := b.(*DomainNode); ok {
		return a.DomainName < x.DomainName
	} else if y, ok := b.(*Domain); ok {
		return a.DomainName < y.DomainName
	}
	panic(MyError.NewError(MyError.ERROR_PARAM, "Param error of b: "+reflect.ValueOf(b).String()))
}

func (a *Domain) Less(b llrb.Item) bool {
	if x, ok := b.(*DomainNode); ok {
		return a.DomainName < x.DomainName
	} else if y, ok := b.(*Domain); ok {
		return a.DomainName < y.DomainName
	}
	panic(MyError.NewError(MyError.ERROR_PARAM, "Param error of b: "+reflect.ValueOf(b).String()))
}

// 1,Trust d.DomainName is really a DomainName, so, have not use dns.IsDomainName for checking
// Check if d is already in the DomainRRTree,if so,make sure update d.DomainRegionTree = dt.DomainRegionTree
func (DT *DomainRRTree) StoreDomainNodeToCache(d *DomainNode) (bool, *MyError.MyError) {
	if dt, err := DT.GetDomainNodeFromCacheWithName(d.DomainName); dt != nil && err == nil {
		fmt.Println("DomainRRCache already has DomainNode of d " + d.DomainName)
		d.DomainRegionTree = dt.DomainRegionTree

	} else if err.ErrorNo != MyError.ERROR_NOTFOUND || err.ErrorNo != MyError.ERROR_TYPE {
		// for not found and type error, we should replace the node
		fmt.Println(err)
		return false, err
	}

	DT.Mutex.Lock()
	DT.LLRB.ReplaceOrInsert(d)
	DT.Mutex.Unlock()
	return true, nil
}

func (DT *DomainRRTree) GetDomainNodeFromCacheWithName(d string) (*DomainNode, *MyError.MyError) {
	if _, ok := dns.IsDomainName(d); ok {
		dn := &Domain{
			DomainName: dns.Fqdn(d),
		}
		return DT.GetDomainNodeFromCache(dn)
	}
	return nil, MyError.NewError(MyError.ERROR_PARAM, "Eorror param: "+reflect.ValueOf(d).String())
}

func (DT *DomainRRTree) GetDomainNodeFromCache(d *Domain) (*DomainNode, *MyError.MyError) {
	dr := DT.LLRB.Get(d)
	if dr != nil {
		if drr, ok := dr.(*DomainNode); ok {
			return drr, nil
		} else {
			return nil, MyError.NewError(MyError.ERROR_TYPE, "Got error result because of the type of return value is "+reflect.TypeOf(dr).String())
		}
	} else {
		return nil, MyError.NewError(MyError.ERROR_NOTFOUND, "Not found from DomainRRTree for param: "+reflect.ValueOf(d.DomainName).String())
	}
	return nil, MyError.NewError(MyError.ERROR_UNKNOWN, "SearchDomainNode got param: "+reflect.ValueOf(d).String())
}

func (DT *DomainRRTree) UpdateDomainNode(d *DomainNode) (bool, *MyError.MyError) {
	if _, ok := query.Check_DomainName(d.DomainName); ok {
		if dt, err := DT.GetDomainNodeFromCache(&d.Domain); dt != nil && err == nil {
			d.DomainRegionTree = dt.DomainRegionTree
			DT.Mutex.Lock()
			r := DT.LLRB.ReplaceOrInsert(d)
			DT.Mutex.Unlock()
			if r != nil {
				return true, nil

			} else {
				//Exception:see source code of "LLRB.ReplaceOrInsert"
				return true, MyError.NewError(MyError.ERROR_UNKNOWN, "Update error, but inserted")
			}
		} else {
			return false, MyError.NewError(MyError.ERROR_NOTFOUND, "DomainRRTree does not has "+reflect.ValueOf(d).String()+" or it has "+reflect.ValueOf(dt).String())
		}
	} else {
		return false, MyError.NewError(MyError.ERROR_PARAM, " Param d "+reflect.ValueOf(d).String()+" is not valid Domain instance")
	}
	return false, MyError.NewError(MyError.ERROR_UNKNOWN, "UpdateDomainNode return unknown error")
}

//Use interface{} as param ,  may refact other func as this
//TODO: this func has not been completed,don't use it
func (DT *DomainRRTree) DelDomainNode(d *Domain) (bool, *MyError.MyError) {
	DT.Mutex.Lock()
	r := DT.LLRB.Delete(d)
	DT.Mutex.Unlock()
	fmt.Println("Delete " + d.DomainName + " from DomainRRCache " + reflect.ValueOf(r).String())
	return true, nil
}

func InitDomainSOANode(d string,
	soa *dns.SOA,
	ns_a []*dns.NS) *DomainSOANode {
	return &DomainSOANode{
		SOAKey: d,
		NS:     ns_a,
		SOA:    soa,
	}
}

func (DS *DomainSOANode) Less(b llrb.Item) bool {
	if x, ok := b.(*DomainSOANode); ok {
		return DS.SOAKey < x.SOAKey
	}
	panic(MyError.NewError(MyError.ERROR_PARAM, "Param b "+reflect.ValueOf(b).String()+" is not valid DomainSOANode or String"))
}

func (ST *DomainSOATree) StoreDomainSOANodeToCache(dsn *DomainSOANode) (bool, *MyError.MyError) {
	if dt, err := ST.GetDomainSOANodeFromCache(dsn); dt != nil && err == nil {
		fmt.Println("DomainSOACache already has DomainSOANode of dsn " + dsn.SOAKey)
	} else if err.ErrorNo != MyError.ERROR_NOTFOUND || err.ErrorNo != MyError.ERROR_TYPE {
		// for not found and type error, we should replace the node
		fmt.Println(err)
		return false, err
	}

	ST.Mutex.Lock()
	ST.LLRB.ReplaceOrInsert(dsn)
	ST.Mutex.Unlock()
	fmt.Println("Store " + dsn.SOAKey + " into DomainSOACache Done!")
	return true, nil
}

func (ST *DomainSOATree) GetDomainSOANodeFromCache(dsn *DomainSOANode) (*DomainSOANode, *MyError.MyError) {
	if dt := ST.LLRB.Get(dsn); dt != nil {
		if dsn_r, ok := dt.(*DomainSOANode); ok {
			return dsn_r, nil
		} else {
			return nil, MyError.NewError(MyError.ERROR_TYPE, "ERROR_TYPE")
		}
	} else {
		return nil, MyError.NewError(MyError.ERROR_NOTFOUND, "Not found soa record via domainname "+dsn.SOAKey)
	}
	return nil, MyError.NewError(MyError.ERROR_UNKNOWN, "Unknown Error!")
}

func (ST *DomainSOATree) GetDomainSOANodeFromCacheWithDomainName(d string) (*DomainSOANode, *MyError.MyError) {
	ds := &DomainSOANode{
		SOAKey: dns.Fqdn(d),
	}
	return ST.GetDomainSOANodeFromCache(ds)
}

//func (ST *DomainSOATree) UpdateDomainSOANode(ds *DomainSOANode) *MyError.MyError {
//
//	ST.LLRB.ReplaceOrInsert(ds)
//	return nil
//}

//todo:have not completed
func (ST *DomainSOATree) DelDomainSOANode(ds *DomainSOANode) *MyError.MyError {
	ST.Mutex.Lock()
	ST.LLRB.Delete(ds)
	ST.Mutex.Unlock()
	return nil
}

func (a *DomainNode) InitRegionTree() (bool, *MyError.MyError) {
	if a.DomainRegionTree == nil {
		a.DomainRegionTree = initDomainRegionTree()
	}
	return true, nil
}

func initDomainRegionTree() *RegionTree {
	//	tbitRadix := bitradix.New32()
	return &RegionTree{
		Radix32: bitradix.New32(),
	}
}

func (RT *RegionTree) GetRegionFromCache(r *Region) (*Region, *MyError.MyError) {
	return RT.GetRegionFromCacheWithAddr(r.NetWorkAddr)
}

func (RT *RegionTree) GetRegionFromCacheWithAddr(addr uint32) (*Region, *MyError.MyError) {
	if r := RT.Radix32.Find(addr, radix_bit); r != nil {
		fmt.Println(r.Value)
		if rr, ok := r.Value.(*Region); ok {
			return rr, nil
		} else {
			return nil, MyError.NewError(MyError.ERROR_NOTVALID, "Found result but not valid,need check !")
		}
	} else {
		return nil, MyError.NewError(MyError.ERROR_NOTFOUND, "Not found result")
	}
	return nil, MyError.NewError(MyError.ERROR_NOTFOUND, "Not found search region "+string(addr))
}

//Todo: need check wheather this region.NetWorkAddr is in Cache Radix tree,but the IpStart and IpEnd are not same as r
// thus, you need to split the region
func CheckRegionFromCache(r *Region) bool {
	if len(r.RR) < 1 {
		return false
	}

	return true
}

func (RT *RegionTree) AddRegionToCache(r *Region) bool {
	if ok := CheckRegionFromCache(r); !ok {
		//Todo: add split region logic
	}
	RT.Mutex.Lock()
	RT.Radix32.Insert(r.NetWorkAddr, radix_bit, r)
	RT.Mutex.Unlock()
	return true
}

func (RT *RegionTree) UpdateRegionToCache(r *Region) bool {
	if rnode := RT.Radix32.Find(r.NetWorkAddr, radix_bit); rnode != nil {
		RT.Mutex.Lock()
		RT.Radix32.Remove(r.NetWorkAddr, radix_bit)
		RT.Radix32.Insert(r.NetWorkAddr, radix_bit, r)
		RT.Mutex.Unlock()
	} else {
		RT.AddRegionToCache(r)
	}
	return true
}

func (RT *RegionTree) DelRegionFromCache(r *Region) (bool, *MyError.MyError) {
	if rnode, e := RT.GetRegionFromCache(r); rnode != nil && e == nil {
		RT.Mutex.Lock()
		RT.Radix32.Remove(r.NetWorkAddr, radix_bit)
		RT.Mutex.Unlock()
		fmt.Println("Remove Region from RegionCache " + string(r.NetWorkAddr))
		return true, nil
	} else {
		return true, MyError.NewError(MyError.ERROR_NOTFOUND, "Not found Region from RegionCache")
	}

}

func (RT *RegionTree) TraverseRegionTree() {
	RT.Radix32.Do(func(r1 *bitradix.Radix32, i int) {
		fmt.Println(r1.Key(),
			r1.Value,
			r1.Bits(),
			r1.Leaf(), i)
	})
}

func NewRegion(r []dns.RR, networkAddr, ipStart, ipEnd uint32) (*Region, *MyError.MyError) {
	if len(r) < 1 {
		return nil, MyError.NewError(MyError.ERROR_PARAM, "cap of r ([]dns.RR) can not be less than 1 ")
	} else {
		fmt.Println("NewRegion: line 314 ", cap(r))
	}
	// When the default region for a domain, the networkAddr and ipStart/ipEnd will be 0
	//	if (networkAddr == 0) || (ipStart == 0) || (ipEnd == 0) {
	//		return nil, MyError.NewError(MyError.ERROR_PARAM, "networkAddr, ipStart, ipEnd could not be 0")
	//	}
	dr := &Region{
		NetWorkAddr: networkAddr,
		IpStart:     ipStart,
		IpEnd:       ipEnd,
		RR:          r,
		RrType:      r[0].Header().Rrtype,
		TTL:         r[0].Header().Ttl,
		UpdateTime:  time.Now(),
	}
	return dr, nil
}

func NewDomainNode(d string, soakey string, t uint32) (*DomainNode, *MyError.MyError) {
	if _, ok := dns.IsDomainName(d); !ok {
		return nil, MyError.NewError(MyError.ERROR_PARAM, d+" is not valid domain name")
	}
	return &DomainNode{
		Domain: Domain{
			DomainName: dns.Fqdn(d),
			SOAKey:     soakey,
			TTL:        t,
		},
		DomainRegionTree: nil,
	}, nil
}

func GeneralDNSBackendQuery(d string, srcIP string) ([]dns.RR, *net.IPNet, *MyError.MyError) {
	_, ns, e := query.QuerySOA(d)
	if e != nil || cap(ns) < 1 {
		return nil, nil, e
	}
	a_rr, _, edns, e := query.QueryA(d, true, ns[0].Ns, "53")
	if e != nil || cap(a_rr) < 1 {
		return nil, nil, e
	}

	//	var ipnet *net.IPNet
	if edns != nil {
		_, ipnet, ee := net.ParseCIDR(strings.Join([]string{edns.Address.String(), strconv.Itoa(int(edns.SourceScope))}, "/"))
		if ee == nil {
			return a_rr, ipnet, nil
		} else {
			return a_rr, nil, MyError.NewError(MyError.ERROR_NOTVALID, ee.Error())
		}
	}
	return a_rr, nil, e
}
