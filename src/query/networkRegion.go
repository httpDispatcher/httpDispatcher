package query

type NetworkRegion struct {
	D       *DomainConfig
	Records *DomainRR
	IpStart uint64
	IpEnd   uint64
	NetMask uint8
}

func NewNetworkRegion(d *DomainConfig, rr *DomainRR, s, e uint64, n uint8) *NetworkRegion {
	return &NetworkRegion{D: d, Records: rr, IpStart: s, IpEnd: e, NetMask: n}

}
