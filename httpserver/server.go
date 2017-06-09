package httpserver

import (
	"net/http"
	"gopkg.in/ini.v1"
	"justdevelop.it/goaway/service"
	"sync"
	"justdevelop.it/goaway/utils"
	"log"
	"justdevelop.it/goaway/repos/verdict"
	"justdevelop.it/goaway/repos/whitelist"
)
var cfg *ini.File
var sm service.ServiceManager
var logger *log.Logger
var requestLogger *log.Logger
var verdictRepo verdict.Repository
var whiteListRepo whitelist.Repository
func Setup(conf *ini.File, s service.ServiceManager) {
	cfg = conf
	sm = s
	logger = utils.NewLogger(cfg.Section("http_server").Key("logFile").String())
	requestLogger = utils.NewLogger(cfg.Section("http_server").Key("requestsLogFile").String())
	verdictRepo = verdict.NewRepo(cfg, sm, logger)
	whiteListRepo = whitelist.NewRepo(cfg, sm, logger)
}

var wg sync.WaitGroup
func Run() {

	routerApi := NewRouter(apiRoutes)
	wg.Add(2)

	go func() {
		logger.Println("Start Api")
		http.ListenAndServe(":8085", routerApi)
	}()
	go func() {
		logger.Println("Start Frontend")
		http.Handle("/", http.FileServer(http.Dir("./httpserver/frontend")))
		http.ListenAndServe(":8088", nil)
	}()
	wg.Wait()
}

