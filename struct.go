// struct
package main

import (
	"bufio"
	"net"
)

type Config struct {
	wrapUrl    string
	wrapAddr   *net.TCPAddr
	cpuprofile string
	logfile    string
	debug      bool
	//	ssdbConn *net.TCPConn
}

type Conn struct {
	conn  *net.TCPConn
	cBuf  *bufio.ReadWriter
	cmds  []Cmd
	multi bool
}

type Cmd struct {
	cmd string
}
