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
	viper.SetConfigFile(filepath.Join(exPath, "config.yml"))
	viper.SetConfigType("yaml")
	err = viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Fatal error config file: %v \n", err)
	}

	C.path = viper.GetStringSlice("path")
	C.log = filepath.Join(exPath, viper.GetString("log"))
}

func mergeTs(dir, output string) {
	log.Printf("Merging TS files under: %v", dir)

	for i := 0; ; i++ {
		srcFilePath := filepath.Join(dir, strconv.Itoa(i))

		_, err := os.Stat(srcFilePath)
		if os.IsNotExist(err) {
			log.Printf("TS file %v not exist.", srcFilePath)
			break
		}

		if err != nil {
			log.Fatalf("Check file %v state error: %v\n", srcFilePath, err)
		}

		src, err := ioutil.ReadFile(srcFilePath)
		if err != nil {
			log.Printf("Can not read TS file %v with error: %v", srcFilePath, err)
			break
		}

		log.Printf("Merging TS file: %s", srcFilePath)

		dst, err := os.OpenFile(output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		log.Println(dst.Name())
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
				mergeTs(filepath.Join(basePath, f.Name()), filepath.Join(basePath, f.Name()+".mp4"))
			}
		}
	}
}
