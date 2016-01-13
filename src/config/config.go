package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/miekg/dns"
)

var ConfigFile string
var RC *RuntimeConfiguration

type MySQLConf struct {
	DomainsInMySQL []string `toml:"domains_in_mysql"`
	MySQLHost      string   `toml:"mysql_host"`
	MySQLPort      int32    `toml:"mysql_port"`
	MySQLDB        string   `toml:"mysql_database"`
	MySQLUser      string   `toml:"mysql_user"`
	MySQLPass      string   `toml:"mysql_password"`
}

type RuntimeConfiguration struct {
	Bind         string     `toml:"bind"`
	Domains      []string   `toml:"domains"`
	MySQLEnabled bool       `toml:"mysql_enable"`
	MySQLConf    *MySQLConf `toml:"mysql"`
	IPDB         string     `toml:"ipdb_path"`
	ServerLog    string     `toml:"server_log"`
	QueryLog     string     `toml:"query_log"`
	LogLevel     string     `toml:"log_level"`
}

func init() {
	ParseCommandline()
	ParseConf(ConfigFile)
}

//todo: complete this func
func InWhiteList(d string) bool {
	d = dns.Fqdn(d)
	for _, x := range RC.Domains {
		if d == dns.Fqdn(x) {
			return true
		}
	}
	return false
}

func IsLocalMysqlBackend(d string) bool {
	d = dns.Fqdn(d)
	if !RC.MySQLEnabled {
		return false
	}
	for _, x := range RC.MySQLConf.DomainsInMySQL {
		if d == dns.Fqdn(x) {
			return true
		}
	}
	return false
}

func ParseCommandline() {
	flag.StringVar(&ConfigFile, "conf", "", "The configuration file of TOML format")
	flag.Parse()
	if ConfigFile == "" {
		fmt.Println("ERROR: You must set the configuration file via -conf flag ")
		fmt.Println("\tuse -h to see more help ")
		os.Exit(1)
	}
	fd, e := os.Open(ConfigFile)
	if e != nil {
		fmt.Println("The configuration file open failed : " + e.Error())
		os.Exit(1)
	} else {
		fd.Close()
	}
}

func ParseConf(file string) bool {
	if x, e := toml.DecodeFile(file, &RC); e != nil {
		fmt.Println("Parse toml configuration file "+file+" error : ", e.Error())
		os.Exit(1)
	} else if len(x.Undecoded()) > 0 {
		for _, xx := range x.Undecoded() {
			fmt.Print(xx, " ")
		}
		fmt.Println(" Decode failed. Please review your configuration file: ", file)
		os.Exit(1)
	}
	fmt.Println("Runtime Configurations:")
	fmt.Println("\tBindTo:          ", RC.Bind)
	fmt.Println("\tEnabled domains: ", RC.Domains)
	fmt.Println("\tMySQL enabled:   ", RC.MySQLEnabled)
	fmt.Println("\tIPDB Path:       ", RC.IPDB)
	fmt.Println("\tServerLog:       ", RC.ServerLog)
	fmt.Println("\tQueryLog:        ", RC.QueryLog)
	fmt.Println("\tLoglevel:        ", RC.LogLevel)
	if RC.MySQLEnabled {
		fmt.Println("MySQL Conf: ")
		fmt.Println("\tMySQL Host: ", RC.MySQLConf.MySQLHost)
		fmt.Println("\tMySQL Port: ", RC.MySQLConf.MySQLPort)
		fmt.Println("\tMySQL DB:   ", RC.MySQLConf.MySQLDB)
		fmt.Println("\tMySQL User: ", RC.MySQLConf.MySQLUser)
		fmt.Println("\tMySQL Pass: ", RC.MySQLConf.MySQLPass)
		fmt.Println("\tDomains in MySQL: ", RC.MySQLConf.DomainsInMySQL)
		fmt.Println("\t\t")
	} else {
		fmt.Println("Notice: MySQL backend is disabled")
	}
	return true
}
