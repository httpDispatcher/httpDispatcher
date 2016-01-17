package query

import (
	"MyError"
	"config"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	//	var RC_MySQLConf = &config.RuntimeConfiguration{
	//		MySQLEnabled:true,
	//		MySQLConf:&
	//	}
	config.ParseConf("/Users/chunsheng/Dropbox/Work/Sina/08.Projects/16.httpDispacher/conf/httpdispacher.toml")
	if config.RC.MySQLEnabled {
		once.Do(func() {
			RC_MySQLConf = config.RC.MySQLConf
			InitMySQL(RC_MySQLConf)
		})
	}
	m.Run()
	os.Exit(1)
}

func TestInitMySQL(t *testing.T) {
	db := InitMySQL(RC_MySQLConf)
	if db != false {
		t.Log("InitMySQL OK")
	} else {
		t.Fail()
	}
}

func TestGetDomainIDFromMySQL(t *testing.T) {
	d_a := []string{
		"www.sina.com.cn",
		"www.baidu.com",
		"www.a.shifen.com",
		"api.weibo.cn",
		"weibo.cn",
		"sinaedge.com",
		"ww2.sinaimg.cn",
	}

	for _, d := range d_a {
		t.Log(d)
		id, e := RRMySQL.GetDomainIDFromMySQL(d)
		if e != nil {
			t.Log(id)
			t.Log(e)
			//			t.Fail()
		} else {
			t.Log(d)
			t.Log(id)
		}
	}
}

func TestGetRegionWithIPFromMySQL(t *testing.T) {
	// d_a := []string{"www.sina.com.cn", "www.baidu.com", "www.a.shifen.com", "api.weibo.cn", "weibo.cn", "sinaedge.com"}
	ipuint32 := uint32(1790519448)
	id, e := RRMySQL.GetRegionWithIPFromMySQL(ipuint32)
	if e == nil {
		t.Log(id.Region)
	} else {
		t.Log(e)
	}
}

func TestGetRRFromMySQL(t *testing.T) {
	t.Log("Test..")
	d_a := []uint32{1, 2, 6, 7, 9}
	r_a := []uint32{0, 1, 2, 6, 7, 8}
	for _, d := range d_a {
		for _, id := range r_a {
			x, e := RRMySQL.GetRRFromMySQL(d, id)
			if e == nil {
				t.Log("DomainId: ", d, " RegionId: ", id, "result:", x.idRR, x.RR)
				//				for _, xx := range x {
				//					t.Log(xx, xx.RR)
				//					if xx.RR.RrType == dns.TypeA {
				//						t.Log("A RR")
				//					} else if xx.RR.RrType == dns.TypeCNAME {
				//						t.Log("CNAME RR")
				//					}
				//				}
			} else {
				t.Log(e)
				if e.ErrorNo == MyError.ERROR_NORESULT {
				} else {
					t.Fail()
				}
			}
		}
	}
}
