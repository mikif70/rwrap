// rwrap

/*
 TODO:
	check ssdb connection:

*/

package main

import (
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

	if config.logfile != "" {
		fLog, err := os.OpenFile(config.logfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal("Failed to open log file", err)
		}
		defer fLog.Close()
		log.SetOutput(fLog)
	} else {
		log.SetOutput(os.Stdout)
	}

	if config.debug {
		log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	} else {
		log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	}

	log.Println(config)

	ln := listen()
	defer ln.Close()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGHUP, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT)

	go func() {
		for {
			conn, err := ln.AcceptTCP()
			if err != nil {
				log.Println("Accept err: ", err.Error())
				continue
			}

			log.SetPrefix(conn.RemoteAddr().String() + ":")
			go manageConnection(conn)
		}
	}()

	for sig := range sigchan {
		log.Println("Signal: ", sig)
		break
		//		os.Exit(0)
	}
}
