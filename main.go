package main

import (
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"sync"

	"github.com/spf13/viper"
)

type config struct {
	path []string
	log  string
}

var C config

var (
	Trace 	*log.Logger
	Info 	*log.Logger
	Warning *log.Logger
	Error 	*log.Logger
)

func loadConfig() {
	ex, err := os.Executable()
	if err != nil {
		Error.Printf("Can not get Executable file path, with error: %v", err)
	}
	exPath := filepath.Dir(ex)

	viper.SetConfigName("config")
	viper.SetConfigFile(filepath.Join(exPath, "config.yml"))
	viper.SetConfigType("yaml")
	err = viper.ReadInConfig()
	if err != nil {
		Error.Printf("Fatal error config file: %v \n", err)
	}

	C.path = viper.GetStringSlice("path")
	C.log = filepath.Join(exPath, viper.GetString("log"))
}

func mergeTs(dir, output string) {
	Info.Printf("Merging TS files under: %v", dir)
	if _, err := os.Stat(output); err == nil {
		Info.Printf("Output file already exists: %s", output);
		return
	} else {
		inputs_files, err := ioutil.ReadDir(dir)
		if err != nil {
			Error.Printf("Failed to read source directory:%s with error: %v",dir, err)
		}
		regex, err := regexp.Compile("[0-9]+")
		if err != nil {
			Error.Printf("Could not compile regexp: %v", err)
		}
		var input_file_names []int
		for _, i := range inputs_files {
			if err != nil {
				Error.Printf("Failed to convert string to regex: %v", err)
			}
			if regex.MatchString(i.Name()) {
				// Info.Printf("%s Match regex",i.Name())
				file_name_int, err := strconv.Atoi(i.Name())
				if err != nil {
					Error.Printf("Failed to convert string to int: %v", err)
				}
				input_file_names = append(input_file_names, file_name_int)
			} else {
				Info.Printf("%s does not match regex.", i.Name())
			}
		}
		sort.Ints(input_file_names)
		Info.Printf("input_file_names: %v", input_file_names)
		for _, input_file_name := range input_file_names {
			file_name_string := strconv.Itoa(input_file_name)
			srcFilePath := filepath.Join(dir, file_name_string)
			_, err := os.Stat(srcFilePath)
			if os.IsNotExist(err) {
				Info.Printf("TS file %v not exist.", srcFilePath)
				return
			}
			if err != nil {
				Error.Printf("Check file %v state error: %v\n", srcFilePath, err)
			}
			src, err := ioutil.ReadFile(srcFilePath)
			if err != nil {
				Info.Printf("Can not read TS file %v with error: %v", srcFilePath, err)
				return
			}
			Info.Printf("Merging TS file: %s", srcFilePath)
			dst, err := os.OpenFile(output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
			log.Println(dst.Name())
			if err != nil {
				Error.Printf("Error open output file, with error: %v", err)
			}
			if _, err := dst.Write(src); err != nil {
				Error.Printf("Error write to output file, with error: %v", err)
			}
			if err := dst.Close(); err != nil {
				Error.Printf("Error close output file, with error : %v", err)
			}
		}
	}
}

func init() {
	loadConfig()
	file, err := os.OpenFile(C.log, os.O_CREATE | os.O_APPEND | os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalln(err)
	}

	Trace   = log.New(io.MultiWriter(file, os.Stdout), "TRACE:   ", log.Ldate|log.Ltime|log.Lshortfile)
	Info    = log.New(io.MultiWriter(file, os.Stdout), "INFO:    ", log.Ldate|log.Ltime|log.Lshortfile)
	Warning = log.New(io.MultiWriter(file, os.Stdout), "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	Error   = log.New(io.MultiWriter(file, os.Stderr), "ERROR:   ", log.Ldate|log.Ltime|log.Lshortfile)
}

func main() {
	var wg sync.WaitGroup
	for _, basePath := range C.path {
		files, err := ioutil.ReadDir(basePath)
		if err != nil {
			Error.Printf("Error opening dir: %v", err)
		}

		for _, f := range files {
			if f.IsDir() {
				wg.Add(1)
				go func(basePath string, f fs.FileInfo) {
					defer wg.Done()
					src_dir := filepath.Join(basePath, f.Name())
					dest_file := filepath.Join(basePath, f.Name()+".mp4")
					mergeTs(src_dir, dest_file)
				}(basePath, f)
			}
		}
	}
	wg.Wait()
}
