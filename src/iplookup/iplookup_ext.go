package iplookup

import (
	"MyError"
	"fmt"
	"utils"
)

//todo:make
var DBFile = "../../data/ip.db.db"

func GetIPinfoWithString(ip string) (Ipinfo, *MyError.MyError) {
	if ipdb := Il_open(DBFile); ipdb > 0 {
		defer Il_close(ipdb)
		nip := NewIp(ip)
		defer DeleteIp(nip)
		ipinfo := NewIpinfo()
		//		defer DeleteIpinfo(ipinfo)  "Can not delete ip info within this func!
		n := Il_search(nip, ipinfo, ipdb)
		if n > 0 {
			return ipinfo, nil
		}
	} else {
		return nil, MyError.NewError(MyError.ERROR_UNKNOWN, "Open ipdb file error :"+DBFile)
	}

	return nil, MyError.NewError(MyError.ERROR_NORESULT, "Can not get ipinfo with ip :"+ip)
}

func GetIpinfoStartEnd(i Ipinfo) (uint32, string, uint32, string) {
	//	defer DeleteIpinfo(i) // Must delete ipinfo !
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

func GetIpinfoStartEndWithIPString(s string) (uint32, uint32) {
	info, e := GetIPinfoWithString(s)
	if e == nil && info != nil {
		w, x, y, z := GetIpinfoStartEnd(info)
		fmt.Println(utils.GetDebugLine(), w, x, y, z)
		return w, y
	} else {
		fmt.Println(utils.GetDebugLine(), e)
		return uint32(0), uint32(0)
	}
}
