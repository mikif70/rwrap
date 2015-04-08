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
	conn          *net.TCPConn
	cBuf          *bufio.ReadWriter
	cmds          []Request
	dovecotStatus Status
	cmdStatus     Status
}

type Request struct {
	cmd   []byte
	param [][]byte
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

const (
	DovecotWait      Status = iota //waiting CMD
	DovecotSelect                  //expecting +OK reply for SELECT
	DovecotGet                     //expecting $-1 / $<size> followed by GET reply
	DovecotMulti                   //expecting +QUEUED
	DovecotDiscard                 //expecting +OK reply for DISCARD
	DovecotExec                    //expecting *<nreplies>
	DovecotExecReply               //expecting EXEC reply
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

func getDovecotStatus(status Status) string {
	switch status {
	case DovecotWait:
		return "WAIT"
	case DovecotSelect:
		return "SELECT"
	case DovecotGet:
		return "GET"
	case DovecotMulti:
		return "MULTI"
	case DovecotDiscard:
		return "DISCARD"
	case DovecotExec:
		return "EXEC"
	case DovecotExecReply:
		return "EXEC_REPLY"
	}

	return ""
}
