package main

import (
	"fmt"
	"gopkg.in/ini.v1"
	"io/ioutil"
	"justdevelop.it/goaway/modules/config"
	"justdevelop.it/goaway/modules/metrics"
	"justdevelop.it/goaway/modules/recorder"
	"justdevelop.it/goaway/utils"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"justdevelop.it/goaway/repos/whitelist"
	"justdevelop.it/goaway/repos/evidence"
	"justdevelop.it/goaway/repos/verdict"
	"runtime"
	"justdevelop.it/goaway/service"
	"justdevelop.it/goaway/httpserver"
)

var configFile = "/etc/goaway/goaway.conf"
var cfg *ini.File
var err error

var installPaths = []string{
	"/etc/goaway",
	"/var/log/goaway",
	"/var/run/goaway",
}
var serviceManager service.ServiceManager
var dispatcher *Dispatcher

func init() {
	cfg = ini.Empty()
}
func main() {
	var helpMsg = `Available commands:
	- start-agent: starts the goaway agent
	- start-server: starts the analyzer and janitor
	- record: records ip address from packets and stores in redis
	- record-debug: activate recorder with bpf filters specified
	- analyze: detects when your server's rps is above threshold and starts anaylzing your traffic for attack
	- guard: this puts all the bad IPs in jail (adds DROP rule to the iptable
	- clean: keeps your database within a sensible size by deleting old records
	- migrate: creates database schemas
	- install: creates required directory & files for goaway
	- install-dev: installs goaway for development use
	- build-config: builds the config file
	- remove: removes goaway directory & files installed
	- startwebnode: launch dummy server on port 8080 for testing
	- start-httpserver: launch api and ui server `
	//No args help
	if len(os.Args) < 2 {
		fmt.Println(helpMsg)
		os.Exit(1)
	}

	//install related commands
	switch os.Args[1] {
	case "install":
		fmt.Println("installing goaway...")
		createInstallPaths()
		fmt.Println("installation complete :)")
		fmt.Println("Now configure goaway using", configFile)
		os.Exit(0)
	case "install-dev":
		fmt.Println("installing goaway...")
		createInstallPaths()
		buildConfig("dev")
		installConfig()
		fmt.Println("installation complete :)")
		fmt.Println("Now configure goaway using", configFile)
		os.Exit(0)
	case "remove":
		uninstall()
		os.Exit(0)
	case "build-config":
		var env = "default"
		if len(os.Args) < 3 {
			fmt.Println("You should specify an env to build config for e.g qa, uat, prod.")
			fmt.Println("Default config was built")
		} else {
			env = os.Args[2]
		}
		buildConfig(env)
		os.Exit(0)
	}

	//running app commands
	if _, err := os.Stat(configFile); err != nil {
		helpMsg = "Install goway by running goaway install or install-dev"
		fmt.Println(helpMsg)
	} else {

		switch os.Args[1] {
		case "record":
			bootstrap()
			for {
				dispatcher.DispatchRecorder(cfg)
				fmt.Println("Gotoutine cnt: ", runtime.NumGoroutine())
				dispatcher.DispatcherWait()
			}
		case "record-debug":
			bootstrap()
			filter := os.Args[2]
			recorder.Debug(filter)
		case "analyze":
			bootstrap()
			utils.RegPid(utils.PROCESS_ANALYZER)
			for {
				dispatcher.DispatchAnalyzer(cfg)
				fmt.Println("Gotoutine cnt: ", runtime.NumGoroutine())
				dispatcher.DispatcherWait()
			}
		case "guard":
			bootstrap()
			utils.RegPid(utils.PROCESS_GUARDIAN)
			for {
				dispatcher.DispatchGuardian(cfg)
				fmt.Println("Gotoutine cnt: ", runtime.NumGoroutine())
				dispatcher.DispatcherWait()
			}
		case "clean":
			bootstrap()
			utils.RegPid(utils.PROCESS_JANITOR)
			for{
				dispatcher.DispatchJanitor(cfg)
				fmt.Println("Gotoutine cnt: ", runtime.NumGoroutine())
				dispatcher.DispatcherWait()
			}
		case "start-agent":
			bootstrap()
			utils.RegPid(utils.PROCESS_AGENT)
			for{
				dispatcher.DispatchRecorder(cfg)
				dispatcher.DispatchGuardian(cfg)
				fmt.Println("Gotoutine cnt: ", runtime.NumGoroutine())
				dispatcher.DispatcherWait()
			}
		case "start-server":
			bootstrap()
			utils.RegPid(utils.PROCESS_SERVER)
			for {
				dispatcher.DispatchServer(cfg)
				fmt.Println("Gotoutine cnt: ", runtime.NumGoroutine())
				dispatcher.DispatcherWait()
			}
		case "migrate":
			bootstrap()
			migrate()
		case "startwebnode":
			bootstrap()
			http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, "Hello!") // send data to client side
			}) // set router
			err := http.ListenAndServe(":8080", nil) // set listen port
			if err != nil {
				log.Fatal("ListenAndServe: ", err)
			}
		case "start-httpserver":
			bootstrap()
			httpserver.Setup(cfg, serviceManager)
			httpserver.Run()
		default:
			fmt.Println(helpMsg)
		}
	}

}

func buildConfig(env string) {
	cfg, err = ini.LoadSources(ini.LoadOptions{Loose: true, AllowBooleanKeys: true}, "conf/default.conf", "conf/" + env + ".conf")

	utils.CheckAndPanic(err)

	f, err := os.Create("conf/goaway.conf")
	utils.CheckAndPanic(err)
	_, err = cfg.WriteTo(f)
	utils.CheckAndPanic(err)
}

func createInstallPaths() {
	for _, path := range installPaths {
		err = os.MkdirAll(path, os.ModePerm)
		if err != nil && strings.Contains(err.Error(), "permission denied") {
			fmt.Println("PERMISSION ERROR: You have to run install as root")
			os.Exit(1)
		}
		filepath.Walk(path, func(name string, info os.FileInfo, err error) error {
			if err == nil {
				err = os.Chmod(name, os.ModePerm)
			}
			return err
		})
	}
}

func installConfig() {
	data, err := ioutil.ReadFile("conf/goaway.conf")
	if err == nil {
		err = ioutil.WriteFile(configFile, data, 0744)
		if err != nil {
			fmt.Println(err)
			log.Fatal(err)
		}
	}
	os.Remove("conf/goaway.conf")
}

func uninstall(){
	fmt.Println("removing goaway...")
	for _, path := range installPaths {
		//first remove all sub files or folders
		files, err := ioutil.ReadDir(path)
		if err != nil {
			fmt.Println(err.Error())
		}

		for _, file := range files {
			fileName := filepath.Join(path, file.Name())
			err = os.Remove(fileName)
			if err != nil {
				fmt.Println("could not delete ", fileName)
			}
		}

		err = os.Remove(path)
		if err != nil && strings.Contains(err.Error(), "permission denied") {

			fmt.Println("PERMISSION ERROR: You have to run remove as root")
			log.Fatal(err)
		}
	}
	fmt.Println("uninstallation complete")
}

func bootstrap() {

	utils.RegPid(utils.PROCESS_RECORDER)
	dispatcher = NewDispatcher()
	//load config
	cfg, err = ini.LoadSources(ini.LoadOptions{AllowBooleanKeys: true}, configFile)
	utils.CheckAndPanic(err)
	//set raygun client
	utils.SetRaygun(cfg)

	defer utils.HandleErrors()

	//set database connections and services
	serviceManager, err = service.New(cfg)
	utils.CheckAndPanic(err)

	// dispatcher
	//go func() {
	//	defer utils.HandleErrors()
	//	dispatcher.ListenQuitChan()
	//}()
	//watch config change
	go func() {
		defer utils.HandleErrors()
		config.WatchConfigDir(cfg, configWatcherCallback)
	}()

	if enabled := cfg.Section("metrics").Key("enabled").MustBool(false); enabled {
		metrics.Init(cfg, serviceManager.GetInfluxDb())
	}else {
		go metrics.MetricsBlackholeTicker()
	}
}

func migrate() {
	logger := utils.NewLogger("/var/log/goaway/error.log")
	err := whitelist.NewRepo(cfg, serviceManager, logger).Migrate()
	utils.CheckAndPanic(err)
	err = evidence.NewRepo(cfg, serviceManager, logger).Migrate()
	utils.CheckAndPanic(err)
	err = verdict.NewRepo(cfg, serviceManager, logger).Migrate()
	utils.CheckAndPanic(err)
}


func configWatcherCallback() error{
	err := serviceManager.Reload(cfg)
	dispatcher.StopRunningModules()

	return err
}
