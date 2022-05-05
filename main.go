package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/viper"
)

type config struct {
	path []string
	log  string
}

var C config

func loadConfig() {
	ex, err := os.Executable()
	if err != nil {
		log.Fatalf("Can not get Executable file path, with error: %v", err)
	}
	exPath := filepath.Dir(ex)

	viper.SetConfigName("config")
	viper.SetConfigFile(exPath + "/" + "config.yml")
	viper.SetConfigType("yaml")
	err = viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Fatal error config file: %v \n", err)
	}

	C.path = viper.GetStringSlice("path")
	C.log = exPath + "/" + viper.GetString("log")
}

func mergeTs(dir, output string) {
	log.Printf("Merging TS files under: %v", dir)

	for i := 0; ; i++ {
		srcFilePath := dir + "/" + strconv.Itoa(i)
		src, err := ioutil.ReadFile(srcFilePath)
		if err != nil {
			log.Printf("No TS file under %s\n", dir)
			log.Printf("Can not open TS file with error: %v", err)

			break
		}

		log.Printf("Merging TS file: %s", srcFilePath)

		dst, err := os.OpenFile(output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {

			log.Fatalf("Error open output file, with error: %v", err)
		}
		if _, err := dst.Write(src); err != nil {

			log.Fatalf("Error write to output file, with error: %v", err)
		}
		if err := dst.Close(); err != nil {

			log.Fatalf("Error close output file, with error : %v", err)
		}

	}

}

func main() {

	loadConfig()
	logTo, err := os.OpenFile(C.log, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		log.Fatalf("Error opening logfile: %v", err)
	}

	defer logTo.Close()

	log.SetOutput(logTo)

	for _, basePath := range C.path {
		files, err := ioutil.ReadDir(basePath)
		if err != nil {
			log.Fatalf("Error opening dir: %v", err)
		}

		for _, f := range files {
			if f.IsDir() {
				mergeTs(basePath+f.Name(), basePath+f.Name()+".mp4")
			}
		}
	}
}
