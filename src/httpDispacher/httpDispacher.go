package main

import (
	"runtime"
	"server"
	"utils"

	"github.com/pkg/profile"
)

//import _ "net/http/pprof"

// param: lower case + Upper Case ,No _ spliter
// Struct unit: Upper Case
// Func: golang style

func main() {
	defer profile.Start(profile.CPUProfile).Stop()
	utils.InitUitls()

	//	go func() {
	//		utils.Logger.Fatal(http.ListenAndServe("localhost:6060", nil))
	//	}()
	//	f, err := os.Create("./cpuprofile.out")
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//
	//	pprof.StartCPUProfile(f)
	//	defer pprof.StopCPUProfile()
	runtime.GOMAXPROCS(8)
	//	utils.Logger.Println(ServerAddr + ":" + ServerPort)
	server.NewServer()

}
