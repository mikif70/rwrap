// rwrap

/*
 TODO:
	check ssdb connection:

*/

package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
)

func main() {

	configure()

	//	config.ssdbAddr, _ = net.ResolveTCPAddr("tcp", config.ssdbUrl)
	config.wrapAddr, _ = net.ResolveTCPAddr("tcp", config.wrapUrl)

	if config.cpuprofile != "" {
		fProfile, err := os.OpenFile(config.cpuprofile, os.O_CREATE, 0666)
		if err != nil {
			log.Fatal("Failed to open profile file", err)
		}
		pprof.StartCPUProfile(fProfile)
		defer pprof.StopCPUProfile()
		defer fProfile.Close()
	}

	ln := listen()
	defer ln.Close()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGHUP, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT)

	go func(ln *net.TCPListener) {
		for {
			conn, err := ln.AcceptTCP()
			if err != nil {
				fmt.Println("Accept err: ", err.Error())
				continue
			}

			go manageConnection(conn)
		}
	}(ln)

	for sig := range sigchan {
		fmt.Println("Signal: ", sig)
		break
		//		os.Exit(0)
	}
}
