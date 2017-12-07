package query

import (
	"os"

	"utils"
)

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
//Third: Use query
// .QueryA with the one name server in DomainSOANode.NS.
//Notice: you should store all the infoformation when it is not in the trees(
// Both DomainSOACache and DomainRRCache )

func init() {
	errCache := InitCache()

	if errCache == nil {
		utils.ServerLogger.Critical(utils.GetDebugLine(), "InitDomainRRCache OK")
		utils.ServerLogger.Critical(utils.GetDebugLine(), "InitDomainSOACache OK")
	} else {
		//fmt.Println(utils.GetDebugLine(), "InitDomainRRCache() or InitDomainSOACache() failed")
		//fmt.Println(utils.GetDebugLine(), "Plase contact chunshengster@gmail.com to get more help ")
		utils.ServerLogger.Critical("InitDomainRRCache() or InitDomainSOACache() failed")
		utils.ServerLogger.Info("Plase contact chunshengster@gmail.com to get more help")
		os.Exit(2)
	}

}
