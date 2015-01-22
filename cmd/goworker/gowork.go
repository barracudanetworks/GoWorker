package main

import (
	"flag"
	"fmt"
	"log"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"reflect"
	"runtime/pprof"

	"strings"

	"github.com/barracudanetworks/GoWorker/config"
	"github.com/barracudanetworks/GoWorker/manager"
	"github.com/barracudanetworks/GoWorker/provider"
	_ "github.com/barracudanetworks/GoWorker/provider/http"
	_ "github.com/barracudanetworks/GoWorker/provider/redis"
	"github.com/barracudanetworks/GoWorker/worker"
	_ "github.com/barracudanetworks/GoWorker/worker/cli"
	_ "github.com/barracudanetworks/GoWorker/worker/disk"
	_ "github.com/barracudanetworks/GoWorker/worker/http"
)

var (
	configFileName = flag.String("conf", "config.json", "Config file to use")
	printConfigs   = flag.Bool("print-confs", false, "Print all of the config options")
	cpuProfile     = flag.String("prof-cpu", "", "write cpu profile to this file")
	memProfile     = flag.String("prof-mem", "", "write memory profile to this file")
)

func init() {

	// go http.ListenAndServe("0.0.0.0:6060", nil)
}

func CPUProf(file string) {
	// log.Println("here")
	// f, err := os.Create(file)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// log.Println("starting cpu profiler")
	// pprof.StartCPUProfile(f)

}

func MemProf(file string) {
	f, err := os.Create(file)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("starting memroy profiler")
	defer pprof.WriteHeapProfile(f)
}

// relaySignals send os signals caught to the manager to allow for graceful shutdowns
func relaySignals(m *manager.Manager) {
	interupt := make(chan os.Signal, 1)
	kill := make(chan os.Signal, 1)
	signal.Notify(kill, os.Kill)
	signal.Notify(interupt, os.Interrupt)

	for i := 0; i < 2; i++ {
		select {
		case s := <-kill:
			log.Println(s)
			os.Exit(1)
		case s := <-interupt:
			m.KillChan <- struct{}{}
			log.Println(s)
		}
	}
	os.Exit(0)
}

// printConfigStruct format and print out the available parameters to use in configuartion
func printConfigStruct(c config.Configer) {
	s := reflect.TypeOf(c.ConfigStruct())
	v := reflect.ValueOf(c.ConfigStruct()).Elem()
	e := s.Elem()
	for i := 0; i < e.NumField(); i++ {
		required := e.Field(i).Tag.Get("required") == "true"
		hasDefault := v.Field(i).String() != "" && v.Field(i).String() != "0"
		hasDescription := e.Field(i).Tag.Get("description") != ""
		fieldType := e.Field(i).Type.String()
		if strings.HasPrefix(fieldType, "config.ConfigBlock") {
			fieldType = "[{},{}]"
		}
		if e.Field(i).PkgPath == "" && e.Field(i).Tag.Get("json") != "" {

			fmt.Printf("\t%15s: %8s required: %8t ", e.Field(i).Tag.Get("json"), fieldType, required)
			if hasDefault {
				fmt.Printf("default: %v", v.Field(i).Interface())
			}
			if hasDescription {
				fmt.Printf("description: %s", e.Field(i).Tag.Get("description"))
			}
			fmt.Println()
		}
	}
}

func printProviderConfigs() {
	for k, v := range provider.Factories {
		fmt.Printf("Provider: %s\n", k)
		printConfigStruct(v())
	}
}

func printWorkerConfigs() {
	for k, v := range worker.Factories {
		fmt.Printf("Worker: %s\n", k)
		printConfigStruct(v())
	}
}

func printManagerConfig() {
	fmt.Printf("Manager:\n")
	printConfigStruct(manager.NewManager())
}

func printAllConfigs() {
	printManagerConfig()
	printProviderConfigs()
	printWorkerConfigs()
}

func main() {
	flag.Parse()
	if *printConfigs {
		printAllConfigs()
		os.Exit(0)
	}
	if *cpuProfile != "" {
		CPUProf(*cpuProfile)
		defer func() {
			log.Println("stopting cpu profiler")
			pprof.StopCPUProfile()
		}()
	}
	log.SetFlags(log.Llongfile)
	conf, confErr := config.LoadAppConfigFromFile(*configFileName)
	if confErr != nil {
		log.Fatal(confErr)
	}
	m := manager.NewManager()
	m.Init(conf)
	go relaySignals(m)
	m.Manage()
}
