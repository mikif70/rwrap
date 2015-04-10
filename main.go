// rwrap

/*
 TODO:
	check ssdb connection:

*/

package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	configure()

	cfg.ssdbAddr, _ = net.ResolveTCPAddr("tcp", cfg.ssdbUrl)
	cfg.wrapAddr, _ = net.ResolveTCPAddr("tcp", cfg.wrapUrl)

	ln := listen()
	defer ln.Close()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGHUP, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT)

	go func(ln *net.TCPListener) {
		for {
			conn, err := ln.AcceptTCP()
			if err != nil {
				cfg.log.Println("Accept err: ", err.Error())
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
		fmt.Println("Signal: ", sig)
		break
	}
}
