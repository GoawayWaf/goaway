package utils

import (
	"fmt"
	"os"
	"github.com/MindscapeHQ/raygun4go"
	"gopkg.in/ini.v1"
)

var raygunClient *raygun4go.Client

func SetRaygun(cfg *ini.File) {

	section, err := cfg.GetSection("raygun")
	if err == nil {
		raygunClient, err = raygun4go.New(section.Key("appName").MustString("appName"), section.Key("key").MustString(""))
		if err != nil {
			LogErrorMessage("Unable to create Raygun client: " + err.Error())
		}
	}
}

func HandleErrors() {
	//log error to /var/log/goaway/error.log
	e := recover()
	if e != nil {

		_, ok := e.(string)
		if ok {
			LogErrorMessage(e.(string))
			//only report to raygun if we have config
			if raygunClient != nil{
				raygunClient.CreateError(e.(string))
			}
		}

		fmt.Println(e)
		os.Exit(8)
	}
}
