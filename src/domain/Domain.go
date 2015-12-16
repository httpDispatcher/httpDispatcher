package domain

import (
	"MyError"
	"fmt"
	"os"
	"query"
	"reflect"
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
	bitradix.Radix32
	sync.Mutex
}

//TODO: redundant data types, need to be redesign
type Domain struct {
	DomainName string
	SOA        string
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
	DomainName string
	NS         []*dns.NS
	SOA        *dns.SOA
}

type DomainConfig struct {
	DomainName           string
	AuthoritativeServers []string
	Port                 string
	Ttl                  string
}

//For domain name and Region RR
type DomainRRTree MuLLRB

//For domain SOA and NS record
type DomainSOATree MuLLRB

//For domain Region and A/CNAME record
type RegionTree MubitRadix

var once sync.Once
var DomainRRDB *DomainRRTree
var DomainSOADB *DomainSOATree

func init() {
	errdb := InitDB()

	if errdb == nil {
		fmt.Println("InitDomainRRDB OK")
		fmt.Println("InitDomainSOADB OK")
	} else {
		fmt.Println("InitDomainRRDB() or InitDomainSOADB() failed")
		os.Exit(2)
	}

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

func InitDB() *MyError.MyError {
	once.Do(func() {
		DomainRRDB = &DomainRRTree{}
		DomainSOADB = &DomainSOATree{}
	})
	return nil
}

//func InitDomainSOADB() *MyError.MyError {
//	once.Do(func() {
//		DomainSOADB = &DomainSOATree{}
//	})
//	return nil
//}

// 1,Trust d.DomainName is really a DomainName, so, don't use dns.IsDomainName for checking
// Check if d is already in the DomainRRTree,if so,make sure update d.DomainRegionTree = dt.DomainRegionTree
func (DT *DomainRRTree) StoreDomainNode(d *DomainNode) (bool, *MyError.MyError) {
	if dt, err := DT.SearchDomainNodeWithName(d.DomainName); dt != nil && err == nil {
		fmt.Println("DomainRRTree already has DomainNode of d " + reflect.ValueOf(dt).String())
		d.DomainRegionTree = dt.DomainRegionTree
	}
	DT.Mutex.Lock()
	DT.LLRB.ReplaceOrInsert(d)
	DT.Mutex.Unlock()
	return true, nil
}

func (DT *DomainRRTree) SearchDomainNodeWithName(d string) (*DomainNode, *MyError.MyError) {
	if _, ok := dns.IsDomainName(d); ok {
		dn := &Domain{
			DomainName: dns.Fqdn(d),
		}
		return DT.SearchDomainNode(dn)
	}
	return nil, MyError.NewError(MyError.ERROR_PARAM, "Eorror param: "+reflect.ValueOf(d).String())
}

func (DT *DomainRRTree) SearchDomainNode(d *Domain) (*DomainNode, *MyError.MyError) {
	dr := DT.LLRB.Get(d)
	if dr != nil {
		if drr, ok := dr.(*DomainNode); ok {
			return drr, nil
		} else {
			return nil, MyError.NewError(MyError.ERROR_TYPE, "Got error result because of the type of return value is "+reflect.TypeOf(dr).String())
		}
	} else {
		return nil, MyError.NewError(MyError.ERROR_NORESULT, "Got no result for param: "+reflect.ValueOf(d).String())
	}
	return nil, MyError.NewError(MyError.ERROR_UNKNOWN, "SearchDomainNode got param: "+reflect.ValueOf(d).String())
}

func (DT *DomainRRTree) UpdateDomainNode(d *DomainNode) (bool, *MyError.MyError) {
	if _, ok := query.Check_DomainName(d.DomainName); ok {
		if dt, err := DT.SearchDomainNodeWithName(d.DomainName); dt != nil && err == nil {
			d.DomainRegionTree = dt.DomainRegionTree
			DT.Mutex.Lock()
			DT.LLRB.ReplaceOrInsert(d)
			DT.Mutex.Unlock()
			return true, nil
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
func (DT *DomainRRTree) DelDomainNode(d interface{}) (bool, *MyError.MyError) {
	var e *MyError.MyError
	var dd *DomainNode
	t := reflect.TypeOf(d).Kind()
	fmt.Println(t)
	switch t {
	case reflect.String:
		if ds, ok := d.(string); ok {
			if _, ok := query.Check_DomainName(ds); ok {

				dd, e = NewDomainNode(ds, "", 0)
			}
		} else {
			return false, MyError.NewError(MyError.ERROR_PARAM, "Error in type of param "+reflect.ValueOf(d).String())
		}
	case reflect.TypeOf(&Domain{}).Kind():
	case reflect.TypeOf(&DomainNode{}).Kind():
	default:
		fmt.Println("fdjlsjflsjdlfj")
	}

	if e == nil {
		DT.Mutex.Lock()
		r := DT.LLRB.Delete(dd)
		DT.Mutex.Unlock()
		if r != nil {
			return true, nil
		} else {
			return false, nil
		}
	}
	return false, nil
}

func InitDomainSOANode(d string,
	soa *dns.SOA,
	ns_a []*dns.NS) *DomainSOANode {
	return &DomainSOANode{
		DomainName: d,
		NS:         ns_a,
		SOA:        soa,
	}
}

func (a *DomainSOANode) Less(b llrb.Item) bool {
	if x, ok := b.(*DomainSOANode); ok {
		return a.DomainName < x.DomainName
	}
	panic(MyError.NewError(MyError.ERROR_PARAM, "Param b "+reflect.ValueOf(b).String()+" is not valid DomainSOANode or String"))
}

func (ST *DomainSOATree) StoreDomainSOANode(dsn *DomainSOANode) (bool, *MyError.MyError) {
	ST.Mutex.Lock()
	ST.LLRB.ReplaceOrInsert(dsn)
	ST.Mutex.Unlock()
	return true, nil
}

func (ST *DomainSOATree) GetDomainSOANode(dsn *DomainSOANode) (*DomainSOANode, *MyError.MyError) {
	dt := ST.LLRB.Get(dsn)
	if dsn_r, ok := dt.(*DomainSOANode); ok {
		return dsn_r, nil
	} else {
		return nil, MyError.NewError(MyError.ERROR_TYPE, "ERROR_TYPE")
	}
}

func (ST *DomainSOATree) SearchDomainSOANodeWithDomainName(d string) (*DomainSOANode, *MyError.MyError) {
	ds := &DomainSOANode{
		DomainName: dns.Fqdn(d),
	}
	fmt.Println(ds)
	fmt.Println(ST)
	//	dsn := ST.LLRB.Get(ds)
	//	if dsn != nil{
	//		if dsn_r,ok := dsn.(*DomainSOANode); ok{
	//			return dsn_r, nil
	//		}else{
	//			fmt.Println(dsn)
	//		}
	//		return nil,MyError.NewError(MyError.ERROR_UNKNOWN,"DomainSOATree returned Not DomainSOANode Type OBJ,the value of returned is " + reflect.ValueOf(dsn).String())
	//	}
	return nil, MyError.NewError(MyError.ERROR_NORESULT, "Returned no result")
}

func (ST *DomainSOATree) UpdateDomainSOANode(ds *DomainSOANode) *MyError.MyError {
	ST.LLRB.ReplaceOrInsert(ds)
	return nil
}

func (ST *DomainSOATree) DelDomainSOANode(ds *DomainSOANode) *MyError.MyError {
	ST.LLRB.Delete(ds)
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
	return &RegionTree{}
}

func (RT *RegionTree) SearchRegion(addr uint32) {
	r := RT.Radix32.Find(addr, radix_bit)
	fmt.Println(r)
	RT.Radix32.Do(func(r1 *bitradix.Radix32, i int) {
		fmt.Println(r1.Key(),
			r1.Value,
			r1.Bits(),
			r1.Leaf(), i)
	})
}

func (RT *RegionTree) AddRegion(r *Region) {

}

func (RT *RegionTree) UpdateRegion(d *Domain, r *Region) {

}

func (RT *RegionTree) DelRegion(d *Domain, r *Region) {

}

func NewRegion(r []dns.RR, networkAddr, ipStart, ipEnd uint32) (*Region, *MyError.MyError) {
	if cap(r) < 1 {
		return nil, MyError.NewError(MyError.ERROR_PARAM, "cap of r ([]dns.RR) can not be less then 1 ")
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

func NewDomainNode(d string, soa string, t uint32) (*DomainNode, *MyError.MyError) {
	if _, ok := dns.IsDomainName(d); !ok {
		return nil, MyError.NewError(MyError.ERROR_PARAM, d+" is not valid domain name")
	}
	//	if ns != nil {
	//		for _,n := range ns {
	//			fmt.Println(n)
	//		}
	//	}
	return &DomainNode{
		Domain: Domain{
			DomainName: dns.Fqdn(d),
			SOA:        soa,
			TTL:        t,
		},
		DomainRegionTree: nil,
	}, nil
}
