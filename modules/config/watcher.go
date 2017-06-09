package config

import (
	"fmt"
	"gopkg.in/ini.v1"
	"io/ioutil"
	"justdevelop.it/goaway/utils"
	"time"
	"os"
	"crypto/md5"
	"io"
	"encoding/hex"
	"path/filepath"
)

var path = "/etc/goaway"
var files = make(map[string]FileMeta)

type FileMeta struct {
	lastModTime time.Time
	hash        string
}

type callback func() error

//checks config directory every 2 seconds for any changes and
//reloads the config
func WatchConfigDir(cfg *ini.File, fn callback) {
	fmt.Println("Watching configs for changes")
	timer := time.NewTicker(time.Second * 2)

	for range timer.C{
		filesInfo, err := ioutil.ReadDir(path)
		utils.CheckAndPanic(err)

		for _, file := range filesInfo {
			if(filepath.Ext(file.Name()) == ".conf"){
				fileMeta, inMap := files[file.Name()]
				fileHash := md5File(filepath.Join(path, file.Name()))
				if inMap {
					if file.ModTime().After(fileMeta.lastModTime) && fileHash != fileMeta.hash{
						//file has been modified - reload config
						fmt.Println("Change detected")
						cfg.Reload()
						fmt.Println("Reloading connections")
						if err := fn(); err != nil {
							utils.LogErrorMessage(err.Error())
						}
					}
				} else {
					files[file.Name()] = FileMeta{}
				}

				files[file.Name()] = FileMeta{
					lastModTime:file.ModTime(),
					hash: fileHash,
				}
			}
		}
	}
}

func md5File(filePath string) string {
	var returnMD5String string
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println(err)
		return returnMD5String
	}
	defer file.Close()
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		fmt.Println(err)
		return returnMD5String
	}
	hashInBytes := hash.Sum(nil)[:16]
	returnMD5String = hex.EncodeToString(hashInBytes)
	return returnMD5String
}
