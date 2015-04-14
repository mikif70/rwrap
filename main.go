// rwrap

/*
 TODO:
	check ssdb connection:

*/

package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	//"time"
)

func main() {

	configure()

	if cfg.logfile != nil {
		defer cfg.logfile.Close()
	}

	cfg.ssdbAddr, _ = net.ResolveTCPAddr("tcp", cfg.ssdbUrl)
	cfg.wrapAddr, _ = net.ResolveTCPAddr("tcp", cfg.wrapUrl)

	ln := listen()
	defer ln.Close()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGHUP, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT)

	go func(ln *net.TCPListener) {
		for {
			//			ln.SetDeadline(time.Now().Add(time.Nanosecond * cfg.deadLine))
			conn, err := ln.AcceptTCP()
			if err != nil {
				log.Println("Accept err: ", err.Error())
				continue
			}

			c := Conn{
				conn: conn,
				cBuf: bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)),
				cmds: make([]Request, 0),
			}

			go c.handleConn()
		}
	}(ln)

	for sig := range sigchan {
		fmt.Printf("\nSignal: %+v\n", sig)
		break
	}
}
