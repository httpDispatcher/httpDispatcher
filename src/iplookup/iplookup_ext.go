package iplookup

import "utils"

func GetIpinfoStartEnd(i Ipinfo) (uint32, string, uint32, string) {
	m := NewIpitem()
	defer DeleteIpitem(m)
	x := Il_bin2human(i, m, Id_code)
	if x != nil {
		x1, x2 := x.GetStart(), x.GetEnd()
		return utils.Ip4ToInt32(utils.StrToIP(x1)),
			x1,
			utils.Ip4ToInt32(utils.StrToIP(x2)),
			x2
	} else {
		return uint32(0), "", uint32(0), ""
	}
}

//
//func (IF Ipinfo) GetInt32End()   {
//
//}
