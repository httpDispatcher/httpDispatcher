package main

import (
	"runtime"
	"server"
	"utils"

	"config"
	"github.com/pkg/profile"
)

// param: lower case + Upper Case ,No _ spliter
// Struct unit: Upper Case
// Func: golang style

func main() {
	config.InitConfig()
	defer profile.Start(profile.CPUProfile).Stop()
	utils.InitLogger()
	runtime.GOMAXPROCS(8)
	server.NewServer()

}
