// conn
package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

type Conn struct {
	conn  *net.TCPConn
	ssdb  *net.TCPConn
	cBuf  *bufio.ReadWriter
	sBuf  *bufio.ReadWriter
	cmds  []Request
	multi bool
}

type Request struct {
	cmd   string
	param []string
}

func (c *Conn) readLine() ([]byte, error) {

	line, err := c.cBuf.ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	l := len(line)

	//	fmt.Println("Read: ", line, string(line))

	if line[l-2] != '\r' {
		return nil, errors.New("Malformed cmd")
	}

	return line[:l-2], nil
}

func (c *Conn) readArgs() (string, error) {

	line, err := c.readLine()
	if err != nil {
		return "", err
	}

	if line[0] != '$' {
		return "", errors.New(fmt.Sprintf("Malformed args %s\n", string(line)))
	}

	line, err = c.readLine()
	if err != nil {
		return "", err
	}

	//	fmt.Println("Args: ", string(line))
	return string(line), nil
}

func (c *Conn) parser() (*Request, error) {

	//	fmt.Println("Parsing")
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

		scmd := strings.ToLower(string(cmd))

		switch scmd {
		case "multi", "exec", "ping":
			return &Request{cmd: scmd}, nil
		default:
			params := make([]string, num-1)
			for i := 0; i < num-1; i++ {
				//				fmt.Printf("Reading %d of %d\n", i, num-1)
				if params[i], err = c.readArgs(); err != nil {
					return nil, err
				}
			}

			return &Request{
				cmd:   scmd,
				param: params,
			}, nil
		}

	default:
		return nil, errors.New(fmt.Sprintf("Malformed command: %s\n", string(line)))
	}

}

func (c *Conn) makeReply(cmd string, buf string) string {

	list := strings.Split(buf, "\n")
	l := len(list)
	list = list[:l-2]
	l = len(list)
	//	fmt.Println(strings.Join(list, "-"))

	switch cmd {
	case "get":
		return "$" + strconv.Itoa(len(list[l-1])) + "\r\n" + list[l-1] + "\r\n"
	case "set":
		return "+OK\r\n"
	case "del", "incrby":
		return ":" + list[l-1] + "\r\n"
	}

	return ""
}

func (c *Conn) exec() (string, error) {

	var retval string

	//	fmt.Println("EXEC:")
	for i := range c.cmds {
		var reply string
		var cmd string
		//		fmt.Printf("%s ", string(c.cmds[i].cmd))
		if c.cmds[i].cmd == "incrby" {
			cmd = "incr"
		} else {
			cmd = c.cmds[i].cmd
		}
		reply += strconv.Itoa(len(cmd)) + "\n" + string(cmd) + "\n"
		for p := range c.cmds[i].param {
			//			fmt.Printf("%s ", string(c.cmds[i].param[p]))
			reply += strconv.Itoa(len(c.cmds[i].param[p])) + "\n" + string(c.cmds[i].param[p]) + "\n"
		}
		fmt.Println("Send cmd: ", reply)
		c.sBuf.WriteString(reply + "\n")
		c.sBuf.Flush()
		buf := make([]byte, 1024)
		n, err := c.sBuf.Read(buf)
		if err != nil {
			return "", err
		}
		fmt.Println("cmd response: ", string(buf[:n]))
		retval += c.makeReply(c.cmds[i].cmd, string(buf[:n]))
	}

	//fmt.Println("Retval: ", retval)

	return retval, nil
}

func (c *Conn) reply(reply string, multi bool) error {

	var retval string

	if multi {
		lines := strings.Split(reply, "\r\n")
		l := len(lines)
		lines = lines[:l-1]
		//		fmt.Println("Lines: ", len(lines), lines)
		for _ = range lines {
			retval += "+QUEUED\r\n"
		}
		retval += "*" + strconv.Itoa(l-1) + "\r\n"
	}

	fmt.Println("Retval: ", retval+reply)
	c.cBuf.WriteString(retval + reply)
	c.cBuf.Flush()
	return nil
}

func (c *Conn) handleConn() error {

	var err error
	defer c.conn.Close()

	log.Println("Connecting SSDB....")
	c.ssdb, err = ssdbConnect(3)
	if err != nil {
		log.Println("SSDB err: ", err.Error())
		return err
	}

	defer c.ssdb.Close()

	c.sBuf = bufio.NewReadWriter(bufio.NewReader(c.ssdb), bufio.NewWriter(c.ssdb))

	log.Println("New Connection: ", c.conn.RemoteAddr())
	start := time.Now()
	startMsg := fmt.Sprintf("%v: Started %v", start, c.conn.RemoteAddr())

	for {
		request, err := c.parser()
		if err != nil {
			fmt.Println("Parser error: ", err)
			break
		}
		//		fmt.Println("Buffered: ", request)
		switch string(request.cmd) {
		case "get", "set", "incr":
			if !c.multi {
				c.cmds = append(c.cmds, *request)
				reply, err := c.exec()
				if err != nil {
					fmt.Println("Exec error: ", err.Error())
					continue
				}
				fmt.Println("Request: ", request)
				fmt.Println("Reply: ", reply)
				c.reply(reply, false)
				c.cmds = make([]Request, 0)
			} else {
				fmt.Println("Default dovectoStatus: ", *request)
				c.cmds = append(c.cmds, *request)
			}
		case "multi":
			c.cBuf.Write([]byte("+OK\r\n"))
			c.cBuf.Flush()
			c.multi = true
			//			fmt.Println("Write +OK")
		case "exec":
			c.multi = false
			reply, err := c.exec()
			if err != nil {
				fmt.Println("Exec error: ", err.Error())
				continue
			}
			//			fmt.Println("Reply: ", reply)
			c.reply(reply, true)
			c.cmds = make([]Request, 0)
		case "ping":
			c.cBuf.Write([]byte("+PONG\r\n"))
			c.cBuf.Flush()
		default:
			fmt.Println("Default dovectoStatus: ", *request)
			c.cmds = append(c.cmds, *request)
		}

		if err != nil {
			fmt.Println("Error: ", err)
			break
		}
	}

	//	fmt.Println("Cmds: ", c.cmds)
	log.Printf("%s - executed cmds in %v\n\n", startMsg, time.Since(start))

	return nil
}
