// options
package main

import (
	"flag"
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

type Config struct {
	wrapUrl    string
	wrapAddr   *net.TCPAddr
	cpuprofile string
	logfname   string
	logfile    *os.File
	debug      bool
	ssdbAddr   *net.TCPAddr
	ssdbUrl    string
	deadLine   time.Duration
}

var (
	cfg = Config{
		wrapUrl:  "0.0.0.0:6380",
		ssdbUrl:  "10.39.80.182:8888",
		logfname: "",
		logfile:  nil,
		debug:    false,
		deadLine: 100,
	}
)

func configure() {
	flag.StringVar(&cfg.ssdbUrl, "s", cfg.ssdbUrl, "ssdb ip:port")
	flag.StringVar(&cfg.wrapUrl, "l", cfg.wrapUrl, "listen ip:port")
	flag.StringVar(&cfg.logfname, "log", cfg.logfname, "write log to file")
	flag.BoolVar(&cfg.debug, "debug", cfg.debug, "activate debug")
	flag.DurationVar(&cfg.deadLine, "t", cfg.deadLine, "read/write deadline [Valid time units are ns, us, ms, s, m, h]")
	flag.Parse()

	var err error

	if cfg.logfname != "" {
		if _, err := os.Stat(cfg.logfname); err == nil {
			newname := cfg.logfname + "." + strconv.FormatInt(time.Now().Unix(), 10)
			os.Rename(cfg.logfname, newname)
		}
		cfg.logfile, err = os.OpenFile(cfg.logfname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			log.Fatalln("Error opening log file: ", err.Error())
		}
		log.SetOutput(cfg.logfile)
	}

	if cfg.debug {
		log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	} else {
		log.SetFlags(log.Ldate | log.Ltime)
	}

	printLog("Config: %+v\n", cfg)
}
