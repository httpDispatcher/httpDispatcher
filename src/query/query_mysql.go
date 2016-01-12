package query

import (
	"MyError"
	"database/sql"
	"fmt"
	"utils"

	"domain"
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
	Region   *domain.RegionNew
}

type MySQLRR struct {
	idRR uint32
	//	Domain  *domain.Domain
	RR *domain.RRNew
}

var RRMySQL *RR_MySQL

func init() {
	if config.RC.MySQLEnabled {
		InitMySQL(config.RC.MySQLConf)
	}
}

func InitMySQL(mcf *config.MySQLConf) bool {
	for x := 0; x < 3; x++ {
		db, err := sql.Open("mysql", mcf.MySQLUser+":"+mcf.MySQLPass+
			"@tcp("+mcf.MySQLHost+":"+strconv.Itoa(int(mcf.MySQLPort))+")/"+mcf.MySQLDB)
		if err == nil {
			RRMySQL = &RR_MySQL{DB: db}
			return true
		} else {
			fmt.Println(utils.GetDebugLine(), err)
		}
	}
	return false
}

func (D *RR_MySQL) GetDomainIDFromMySQL(d string) (int, *MyError.MyError) {
	if e := D.DB.Ping(); e != nil {
		if ok := InitMySQL(config.RC.MySQLConf); ok != true {
			return -1, MyError.NewError(MyError.ERROR_UNKNOWN, "Connect MySQL Error")
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
		if ok := InitMySQL(config.RC.MySQLConf); ok != true {
			return nil, MyError.NewError(MyError.ERROR_UNKNOWN, "Connect MySQL Error")
		}
	}
	sqlstring := "Select idRegion, StartIP, EndIP, NetAddr, NetMask From " + RegionTable + " Where ? >= StartIP and ? <= EndIP"
	var idRegion, StartIP, EndIP, NetAddr, NetMask uint32
	ee := D.DB.QueryRow(sqlstring, ip, ip).Scan(&idRegion, &StartIP, &EndIP, &NetAddr, &NetMask)
	switch {
	case ee == sql.ErrNoRows:
		fmt.Println(utils.GetDebugLine(), ee)
		return nil, MyError.NewError(MyError.ERROR_NOTFOUND, "Not found for Region for IP: "+strconv.Itoa(int(ip)))
	case ee != nil:
		fmt.Println(utils.GetDebugLine(), ee)

		return nil, MyError.NewError(MyError.ERROR_UNKNOWN, ee.Error())
	default:
		fmt.Println(utils.GetDebugLine(), "GetRegionWithIPFromMySQL: ",
			" idRegion: ", idRegion, " StartIP: ", StartIP, " EndIP: ", EndIP, " NetAddr: ",
			NetAddr, " NetMask: ", NetMask)
		return &MySQLRegion{
			IdRegion: idRegion,
			Region: &domain.RegionNew{
				StarIP:  StartIP,
				EndIP:   EndIP,
				NetAddr: NetAddr,
				NetMask: NetMask},
		}, nil
	}
	return nil, MyError.NewError(MyError.ERROR_UNKNOWN, "Unknown error!")
}

func (D *RR_MySQL) GetRRFromMySQL(domainId, regionId uint32) ([]*MySQLRR, *MyError.MyError) {
	if e := D.DB.Ping(); e != nil {
		if ok := InitMySQL(config.RC.MySQLConf); ok != true {
			return nil, MyError.NewError(MyError.ERROR_UNKNOWN, "Connect MySQL Error")
		}
	}
	sqlstring := "Select idRRTable, Rrtype, Class, Ttl, Target From " +
		RRTable +
		" where idDomainName = ? and idRegion = ?"
	rows, e := D.DB.Query(sqlstring, domainId, regionId)
	if e == nil {
		var MyRR []*MySQLRR
		for rows.Next() {
			var u, v uint32
			var w, x uint16
			var z string
			e := rows.Scan(&u, &w, &x, &v, &z)
			if e == nil {
				fmt.Println(utils.GetDebugLine(), u, w, x, v, z)
				MyRR = append(MyRR, &MySQLRR{
					idRR: u,
					RR: &domain.RRNew{
						RrType: w,
						Class:  x,
						Target: z,
					},
				})
			} else {
				fmt.Println(utils.GetDebugLine(), e, rows.Err())
				return nil, MyError.NewError(MyError.ERROR_NOTVALID, e.Error()+rows.Err().Error())
			}
		}
		if len(MyRR) < 1 {
			return nil, MyError.NewError(MyError.ERROR_NORESULT, "No Result for domainId : "+strconv.Itoa(int(domainId))+" RegionID: "+strconv.Itoa(int(regionId)))
		}
		fmt.Println(utils.GetDebugLine(), MyRR)
		return MyRR, nil
	}
	fmt.Println(utils.GetDebugLine(), e)
	return nil, MyError.NewError(MyError.ERROR_UNKNOWN, e.Error())

}
