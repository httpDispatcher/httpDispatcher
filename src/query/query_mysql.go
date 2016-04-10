package query

import (
	"MyError"
	"database/sql"
	"utils"

	"strconv"

	"config"

	_ "github.com/go-sql-driver/mysql"
	"github.com/miekg/dns"
)

const (
	RRTable     = "RRTable"
	DomainTable = "DomainTable"
	RegionTable = "RegionTable"
)

type RR_MySQL struct {
	DB *sql.DB
}

type MySQLRegion struct {
	IdRegion uint32
	Region   *RegionNew
}

type MySQLRR struct {
	idRR []uint32
	//	Domain  *domain.Domain
	RR *RRNew
}

var RRMySQL *RR_MySQL
var RC_MySQLConf *config.MySQLConf

//todo: "InitMySQL(config.RC.MySQLConf)" need to be refact. use RC in logic func is not so good!

//func init() {
//	fmt.Println(config.RC)
//	if config.RC != nil {
//		if config.RC.MySQLEnabled {
//			RC_MySQLConf = config.RC.MySQLConf
//			InitMySQL(RC_MySQLConf)
//		}
//	}else {
//		panic("config.RC is nil")
//	}
//}

func InitMySQL(mcf *config.MySQLConf) bool {
	for x := 0; x < 3; x++ {
		db, err := sql.Open("mysql", mcf.MySQLUser+":"+mcf.MySQLPass+
			"@tcp("+mcf.MySQLHost+":"+strconv.Itoa(int(mcf.MySQLPort))+")/"+mcf.MySQLDB)
		if err == nil {
			RRMySQL = &RR_MySQL{DB: db}
			return true
		} else {
			utils.QueryLogger.Error("Connect MySQL faile with conf: %v, error: %v", mcf, err.Error())
		}
	}
	return false
}

func (D *RR_MySQL) GetDomainIDFromMySQL(d string) (int, *MyError.MyError) {
	if e := D.DB.Ping(); e != nil {
		if ok := InitMySQL(RC_MySQLConf); ok != true {
			return 0, MyError.NewError(MyError.ERROR_UNKNOWN, "Connect MySQL Error")
		}
	}
	sql_string := "Select idDomainName From " + DomainTable + " Where DomainName=?"
	var idDomainName int
	e := D.DB.QueryRow(sql_string, dns.Fqdn(d)).Scan(&idDomainName)
	switch {
	case e == sql.ErrNoRows:
		return 0, MyError.NewError(MyError.ERROR_NOTFOUND, "Not found record for DomainName:"+d)
	case e != nil:
		return -1, MyError.NewError(MyError.ERROR_UNKNOWN, e.Error())
	default:
		//		if id,e := strconv.Atoi(idDomainName); e== nil{
		return idDomainName, nil
		//		}
	}
	return -1, MyError.NewError(MyError.ERROR_UNKNOWN, "Unknown Error!")
}

func (D *RR_MySQL) GetRegionWithIPFromMySQL(ip uint32) (*MySQLRegion, *MyError.MyError) {
	if e := D.DB.Ping(); e != nil {
		if ok := InitMySQL(RC_MySQLConf); ok != true {
			return nil, MyError.NewError(MyError.ERROR_UNKNOWN, "Connect MySQL Error")
		}
	}
	sqlstring := "Select idRegion, StartIP, EndIP, NetAddr, NetMask From " + RegionTable + " Where ? >= StartIP and ? <= EndIP"
	var idRegion, StartIP, EndIP, NetAddr, NetMask uint32
	ee := D.DB.QueryRow(sqlstring, ip, ip).Scan(&idRegion, &StartIP, &EndIP, &NetAddr, &NetMask)
	switch {
	case ee == sql.ErrNoRows:
		utils.QueryLogger.Error(ee.Error())
		return nil, MyError.NewError(MyError.ERROR_NOTFOUND, "Not found for Region for IP: "+strconv.Itoa(int(ip)))
	case ee != nil:
		utils.QueryLogger.Error(ee.Error())

		return nil, MyError.NewError(MyError.ERROR_UNKNOWN, ee.Error())
	default:
		utils.QueryLogger.Debug("GetRegionWithIPFromMySQL: ",
			" idRegion: ", idRegion, " StartIP: ", StartIP, " EndIP: ", EndIP, " NetAddr: ",
			NetAddr, " NetMask: ", NetMask, " srcIP: ", utils.Int32ToIP4(ip).String())
		return &MySQLRegion{
			IdRegion: idRegion,
			Region: &RegionNew{
				StarIP:  StartIP,
				EndIP:   EndIP,
				NetAddr: NetAddr,
				NetMask: NetMask},
		}, nil
	}
	return nil, MyError.NewError(MyError.ERROR_UNKNOWN, "Unknown error!")
}

func (D *RR_MySQL) GetRRFromMySQL(domainId, regionId uint32) (*MySQLRR, *MyError.MyError) {
	if e := D.DB.Ping(); e != nil {
		if ok := InitMySQL(RC_MySQLConf); ok != true {
			return nil, MyError.NewError(MyError.ERROR_UNKNOWN, "Connect MySQL Error")
		}
	}
	sqlstring := "Select idRRTable, Rrtype, Class, Ttl, Target From " +
		RRTable +
		" where idDomainName = ? and idRegion = ?"
	rows, e := D.DB.Query(sqlstring, domainId, regionId)
	if e == nil {
		var MyRR *MySQLRR
		var rtype_tmp uint16
		var isHybird bool = false //if both hava dns.TypeA and dns.TypeCNAME, isHybird is true ,else false
		var uu []uint32
		var u, v uint32 // u is for idRRTable, v is for Ttl
		var w, x uint16 // w is for Rrtype(5 for dns.TypdCNAME and 1 for dns.TypeA),
		var zz []string // for Target(s)
		var z string
		var rows_count int
		for rows.Next() {
			rows_count++
			e := rows.Scan(&u, &w, &x, &v, &z)
			if e == nil {
				utils.QueryLogger.Debug(" got row : %v,%v,%v,%v,%v ", u, w, x, v, z)
				if (rtype_tmp != uint16(0)) && (rtype_tmp != w) {
					isHybird = true
					rtype_tmp = w
					utils.QueryLogger.Debug(" got row: %v,%v,%v,%v,%v,%v,%v", rtype_tmp, w, u, x, v, z, isHybird)
					//rtype is not same as previous one
				} else {
					uu = append(uu, u)
					zz = append(zz, z)
				}
			} else {
				utils.QueryLogger.Error(" rows.Scan error %s ,%v", e.Error(), rows.Err())
				return nil, MyError.NewError(MyError.ERROR_NOTVALID, e.Error()+rows.Err().Error())
			}
		}
		if rows_count > 0 {
			utils.QueryLogger.Debug("get uu: %v, zz: %v", uu, zz)
			if isHybird {
				utils.QueryLogger.Infof("Both TypeA and TypeCNAME for domain: " +
					strconv.Itoa(int(domainId)) + " and Regionid :" + strconv.Itoa(int(regionId)) +
					", that's not good !")
				MyRR = nil
			} else {
				MyRR = &MySQLRR{
					idRR: uu,
					RR: &RRNew{
						RrType: w,
						Class:  x,
						Target: zz,
					},
				}
				return MyRR, nil
			}
		} else {
			return nil, MyError.NewError(MyError.ERROR_NORESULT, "No Result for domainId : "+strconv.Itoa(int(domainId))+" RegionID: "+strconv.Itoa(int(regionId)))
		}
	}
	utils.QueryLogger.Error("MySQL query error: %v", e)
	return nil, MyError.NewError(MyError.ERROR_UNKNOWN, e.Error())

}
