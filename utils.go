package main

import (
	"bufio"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
)

type Logger struct {
	info         *log.Logger
	warn         *log.Logger
	err          *log.Logger
	access       chan string
	accessWriter *bufio.Writer
	accessFile   *os.File
	runFile      *os.File
}

var logger *Logger

var levelMap = map[string]int{"info": 1, "warn": 2, "error": 3}

var GConfig = make(map[string]interface{})

func AssertSuccess(err error) {
	if err != nil {
		panic(err)
	}
}

func InitLog() {
	var err error

	if logger != nil {
		if logger.runFile != nil {
			logger.runFile.Close()
		}
		if logger.accessFile != nil {
			logger.accessWriter.Flush()
			logger.accessFile.Close()
		}
		logger.info = nil
		logger.warn = nil
		logger.err = nil
		logger.access = nil
	} else {
		logger = new(Logger)
	}

	// init run log
	path := GConfig["runlog.path"].(string)
	logger.runFile, err = os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	AssertSuccess(err)

	var level = GConfig["runlog.level"].(string)
	var lv, ok = levelMap[level]
	if !ok {
		panic("level [" + level + "] not \"error\" \"warn\" or \"info\"")
	}

	if lv <= 3 {
		logger.err = log.New(logger.runFile, "[ERROR] ", log.LstdFlags)
	}

	if lv <= 2 {
		logger.warn = log.New(logger.runFile, "[WARN] ", log.LstdFlags)
	}

	if lv <= 1 {
		logger.info = log.New(logger.runFile, "[INFO] ", log.LstdFlags)
	}

	// init access log
	path = GConfig["accesslog.path"].(string)
	logger.accessFile, err = os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	AssertSuccess(err)

	bufsize := GConfig["accesslog.buf"].(int)
	logger.access = make(chan string, 64)
	logger.accessWriter = bufio.NewWriterSize(logger.accessFile, bufsize)

	go writeAccess()
}

func writeAccess() {
	for {
		logger.accessWriter.WriteString(<-logger.access)
	}
}

func Linfo(v ...interface{}) {
	if logger.info != nil {
		logger.info.Println(v...)
	}
}

func Lwarn(v ...interface{}) {
	if logger.warn != nil {
		logger.warn.Println(v...)
	}
}

func Lerror(v ...interface{}) {
	if logger.err != nil {
		logger.err.Println(v...)
	}
}

func Laccess(r *Request) {
	if logger.access != nil {
		logger.access <- r.String()
	}
}

func readYaml(path string) (conf map[string]interface{}) {
	if content, err := ioutil.ReadFile(path); err != nil {
		if os.IsNotExist(err) {
			return conf
		}
		panic(err)
	} else {
		AssertSuccess(yaml.Unmarshal(content, &conf))
		return conf
	}
}

func LoadConf(path string, localPath string) {
	GConfig = readYaml(path)
	lconf := readYaml(localPath)

	for k, v := range lconf {
		GConfig[k] = v
	}
}
