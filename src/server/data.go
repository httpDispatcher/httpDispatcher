package server

type DNS_RR struct {
	priority string `priority`
	ip       string `ip`
	ttl      string `ttl`
}

type DNS_RR_Z struct {
	y string `priority`
	p string `ip`
	t string `ttl`
}

type RDATA struct {
	domain    string   `domain`
	device_ip string   `device_ip`
	device_sp string   `device_sp`
	code      string   `code`
	dns       []DNS_RR `dns_rr`
}

type RDATA_Z struct {
	m string     `domain`
	i string     `device_ip`
	s string     `device_sp`
	c string     `code`
	d []DNS_RR_Z `dns_rr_z`
}

func NewDnsRR(y, p, t string) *DNS_RR {
	return &DNS_RR{
		priority: y,
		ip:       p,
		ttl:      t,
	}
}

func NewDnsRRZ(y, p, t string) *DNS_RR_Z {
	return &DNS_RR_Z{
		y: y,
		p: p,
		t: t,
	}
}

func NewRdata(m, i, s, c string, dns []DNS_RR) *RDATA {
	return &RDATA{
		domain:    m,
		device_ip: i,
		device_sp: s,
		code:      c,
		dns:       dns,
	}
}
func (r *RDATA) AddDNSRR(d DNS_RR) error {
	r.dns = append(r.dns, d)
	return nil
}

func NewRdataZ(m, i, s, c string, dns []DNS_RR_Z) *RDATA_Z {
	return &RDATA_Z{
		m: m,
		i: i,
		s: s,
		c: c,
		d: dns,
	}
}

func (r *RDATA_Z) AddDNSRR_Z(d DNS_RR_Z) error {
	r.d = append(r.d, d)
	return nil
}
