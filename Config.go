// options
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
)

type Config struct {
	wrapUrl    string
	wrapAddr   *net.TCPAddr
	cpuprofile string
	logfname   string
	logfile    *os.File
	log        *log.Logger
	debug      bool
	ssdbAddr   *net.TCPAddr
	ssdbUrl    string
}

var (
	cfg = Config{
		wrapUrl:  "0.0.0.0:6380",
		ssdbUrl:  "10.39.80.182:8888",
		logfname: "",
		logfile:  os.Stdout,
		debug:    false,
	}
)

func configure() {
	flag.StringVar(&cfg.ssdbUrl, "s", cfg.ssdbUrl, "ssdb ip:port")
	flag.StringVar(&cfg.wrapUrl, "l", cfg.wrapUrl, "listen ip:port")
	flag.StringVar(&cfg.logfname, "log", cfg.logfname, "write log to file")
	flag.BoolVar(&cfg.debug, "debug", cfg.debug, "activate debug")

	flag.Parse()

	fmt.Println(cfg)

	var err error

	if cfg.logfname != "" {
		cfg.logfile, err = os.OpenFile(cfg.logfname, os.O_CREATE|os.O_TRUNC, 0777)
		if err != nil {
			log.Fatalln("Error opening log file: ", err.Error())
		}
	}

	cfg.log = log.New(cfg.logfile, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
}
