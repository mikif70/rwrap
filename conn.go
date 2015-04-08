// conn
package main

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"
)

func (c *Conn) readLine() ([]byte, error) {

	line, err := c.cBuf.ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	l := len(line)

	fmt.Println("Read: ", line, string(line))

	if line[l-2] != '\r' {
		return nil, errors.New("Malformed cmd")
	}

	return line[:l-2], nil
}

func (c *Conn) readArgs() ([]byte, error) {

	line, err := c.readLine()
	if err != nil {
		return nil, err
	}

	if line[0] != '$' {
		return nil, errors.New(fmt.Sprintf("Malformed args %s\n", string(line)))
	}

	line, err = c.readLine()
	if err != nil {
		return nil, err
	}

	fmt.Println("Args: ", string(line))
	return line, nil
}

func (c *Conn) parser() (*Request, error) {

	fmt.Println("Parsing")
	line, err := c.readLine()
	if err != nil {
		return nil, err
	}

	switch line[0] {
	case '*':
		num, err := strconv.Atoi(string(line[1:]))
		if err != nil {
			return nil, err
		}

		cmd, err := c.readArgs()
		if err != nil {
			return nil, err
		}

		switch string(cmd) {
		case "MULTI", "EXEC":
			return &Request{cmd: cmd}, nil
		default:
			params := make([][]byte, num-1)
			for i := 0; i < num-1; i++ {
				fmt.Printf("Reading %d of %d\n", i, num-1)
				if params[i], err = c.readArgs(); err != nil {
					return nil, err
				}
			}

			return &Request{
				cmd:   cmd,
				param: params,
			}, nil
		}

	default:
		return nil, errors.New(fmt.Sprintf("Malformed command: %s\n", string(line)))
	}

}

func (c *Conn) handleConn() error {
	log.Println("New Connection: ", c.conn.RemoteAddr())
	start := time.Now()
	startMsg := fmt.Sprintf("%v: Started %v", start, c.conn.RemoteAddr())

	for {
		request, err := c.parser()
		if err != nil {
			fmt.Println("Parser error: ", err)
			break
		}
		fmt.Println("Buffered: ", request)
		switch string(request.cmd) {
		case "GET":
			c.cBuf.Write([]byte("$1\r\n10\r\n"))
			c.cBuf.Flush()
			fmt.Println("Write $1")
			c.cmds = append(c.cmds, *request)
			//			c.dovecotStatus = DovecotWait
		case "MULTI":
			c.cBuf.Write([]byte("+OK\r\n"))
			c.cBuf.Flush()
			fmt.Println("Write +OK")
		case "EXEC":
			c.cBuf.Write([]byte("+QUEUED\r\n+QUEUED\r\n+QUEUED\r\n+QUEUED\r\n*4\r\n:1\r\n:1\r\n+OK\r\n+OK\r\n"))
			c.cBuf.Flush()
			fmt.Println("Write +QUEUED: ", c.cmds)
			//			c.dovecotStatus = DovecotExecReply
			//		case DovecotExecReply:
			//			c.dovecotStatus = DovecotWait
		default:
			fmt.Println("Default dovectoStatus")
			c.cmds = append(c.cmds, *request)
			//			c.dovecotStatus = DovecotWait
		}

		if err != nil {
			fmt.Println("Error: ", err)
			break
		}
	}

	fmt.Println("Cmds: ", c.cmds)
	log.Printf("%s - executed cmds in %v\n\n", startMsg, time.Since(start))

	return nil
}
