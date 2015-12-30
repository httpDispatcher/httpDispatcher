package query

import (
	"MyError"
	"database/sql"
	"fmt"
	"utils"

	"domain"
	_ "github.com/go-sql-driver/mysql"
	"github.com/miekg/dns"
	"strconv"
)

const (
	MySQL_Host  = "127.0.0.1"
	MySQL_Port  = "3306"
	MySQL_User  = "root"
	MySQL_Pass  = ""
	MySQL_DB    = "httpDispacher"
	RRTable     = "RRTable"
	DomainTable = "DomainTable"
	RegionTable = "RegionTable"
)

type RR_MySQL struct {
	DB *sql.DB
}

type MySQLRegion struct {
	idRegionID uint32
	Region     *domain.RegionNew
}

type MySQLRR struct {
	idRR     uint32
	idRegion uint32
	RR       *domain.RRNew
}

var RRMySQL *RR_MySQL

func init() {
	InitMySQL()
}

func InitMySQL() bool {
	for x := 0; x < 3; x++ {
		db, err := sql.Open("mysql", MySQL_User+":"+MySQL_Pass+"@tcp("+MySQL_Host+":"+MySQL_Port+")/"+MySQL_DB)
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
		if ok := InitMySQL(); ok != true {
			return -1, MyError.NewError(MyError.ERROR_UNKNOWN, "Connect MySQL Error")
		}
	}
	sql_string := "Select idDomainName From " + DomainTable + " Where DomainName = ?"
	var idDomainName int
	e := D.DB.QueryRow(sql_string, dns.Fqdn(d)).Scan(&idDomainName)
	switch {
	case e == sql.ErrNoRows:
		return 0, MyError.NewError(MyError.ERROR_NOTFOUND, "Not found for DomainName:"+d)
	case e != nil:
		return -1, MyError.NewError(MyError.ERROR_UNKNOWN, e.Error())
	default:
		//		if id,e := strconv.Atoi(idDomainName); e== nil{
		return idDomainName, nil
		//		}
	}
	return -1, MyError.NewError(MyError.ERROR_UNKNOWN, "Unknown Error!")
}

func (D *RR_MySQL) GetDomainRegionWithIPFromMySQL(d string, ip uint32) (*MySQLRegion, *MyError.MyError) {
	if id, e := D.GetDomainIDFromMySQL(d); e == nil {
		sqlstring := "Select idRegion, StartIP, EndIP, NetAddr, NetMask From " + RegionTable + " Where DomainNameID=? and ? >= StartIP and ? <= EndIP"
		var idRegion, StartIP, EndIP, NetAddr, NetMask uint32
		ee := D.DB.QueryRow(sqlstring, id, ip, ip).Scan(&idRegion, &StartIP, &EndIP, &NetAddr, &NetMask)
		switch {
		case e == sql.ErrNoRows:
			fmt.Println(utils.GetDebugLine(), ee)

			return nil, MyError.NewError(MyError.ERROR_NOTFOUND, "Not found for Region for Domain: "+d+" IP: "+strconv.Itoa(int(ip)))
		case e != nil:
			fmt.Println(utils.GetDebugLine(), ee)

			return nil, MyError.NewError(MyError.ERROR_UNKNOWN, e.Error())
		default:
			fmt.Println(utils.GetDebugLine(), idRegion, StartIP, EndIP, NetAddr, NetMask)
			return &MySQLRegion{
				idRegionID: idRegion,
				Region: &domain.RegionNew{
					StarIP:  StartIP,
					EndIP:   EndIP,
					NetAddr: NetAddr,
					NetMask: NetMask},
			}, nil
		}
	} else {
		fmt.Println(utils.GetDebugLine(), e)
		return nil, e
	}
	return nil, MyError.NewError(MyError.ERROR_UNKNOWN, "Unknown error!")
}

func (D *RR_MySQL) GetRRFromMySQL(regionId uint32) ([]*MySQLRR, error) {
	if e := D.DB.Ping(); e != nil {
		if ok := InitMySQL(); ok != true {
			return nil, MyError.NewError(MyError.ERROR_UNKNOWN, "Connect MySQL Error")
		}
	}
	sqlstring := "Select idRRTable, RegionID, Rrtype, Class, Ttl,Target From " + RRTable + " where RegionID = ?"
	rows, e := D.DB.Query(sqlstring, regionId)
	if e == nil {
		var MyRR []*MySQLRR
		for rows.Next() {
			var u, v, y uint32
			var w uint16
			var x uint8
			var z string
			e := rows.Scan(&u, &v, &w, &x, &y, &z)
			if e == nil {
				MyRR = append(MyRR, &MySQLRR{
					idRR:     u,
					idRegion: v,
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
			return nil, MyError.NewError(MyError.ERROR_NORESULT, "No Result for RegionID: "+strconv.Itoa(int(regionId)))
		}
		return MyRR, nil
	}
	fmt.Println(utils.GetDebugLine(), e)
	return nil, MyError.NewError(MyError.ERROR_UNKNOWN, e.Error())

}
