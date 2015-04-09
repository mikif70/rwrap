// options
package main

import (
	"flag"
	"fmt"
)

var (
	config = &Config{
		wrapUrl:    "0.0.0.0:6380",
		ssdbUrl:    "10.39.80.182:8888",
		cpuprofile: "",
		logfile:    "",
		debug:      false,
	}
)

func configure() {
	flag.StringVar(&config.ssdbUrl, "s", config.ssdbUrl, "ssdb ip:port")
	flag.StringVar(&config.wrapUrl, "l", config.wrapUrl, "listen ip:port")
	flag.StringVar(&config.cpuprofile, "cpuprofile", config.cpuprofile, "write cpu profile to file")
	flag.StringVar(&config.logfile, "log", config.logfile, "write log to file")
	flag.BoolVar(&config.debug, "debug", config.debug, "activate debug")

	flag.Parse()

	fmt.Println(config)
}
