package main

import (
	"runtime"

	"github.com/pkg/profile"

	"config"
	"query"
	"server"
	"utils"
)

// param: lower case + Upper Case ,No _ spliter
// Struct unit: Upper Case
// Func: golang style

func main() {
	config.InitConfig()

	if config.EnableProfile {
		defer profile.Start(profile.CPUProfile).Stop()
	}

	runtime.GOMAXPROCS(runtime.NumCPU() * 3)
	utils.InitLogger()
	if config.RC.MySQLEnabled {
		query.RC_MySQLConf = config.RC.MySQLConf
		query.InitMySQL(query.RC_MySQLConf)
	}
	server.NewServer()
}
