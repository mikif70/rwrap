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
	ssdbAddr   *net.TCPAddr
	ssdbUrl    string
}

type Conn struct {
	conn *net.TCPConn
	ssdb *net.TCPConn
	cBuf *bufio.ReadWriter
	sBuf *bufio.ReadWriter
	cmds []Request
}

type Request struct {
	cmd   string
	param []string
}

type Status int

const (
	CmdStar   Status = iota // *
	CmdDollar               // $
	CmdPlus                 // +
	CmdColon                // :
	CmdCmd                  // cmd
	CmdParam                // params
)

func getCmdStatus(status Status) string {
	switch status {
	case CmdStar:
		return "*"
	case CmdDollar:
		return "$"
	case CmdPlus:
		return "+"
	case CmdColon:
		return ":"
	case CmdCmd:
		return "_"
	}

	return ""
}
