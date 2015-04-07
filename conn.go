// conn
package main

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"
)

/*
func (c *Conn) writeCmd(buf string) (string, error) {
	resp := make([]byte, 256)

	_, err := (*c.ssdb).Write([]byte(buf))
	if err != nil {
		log.Println("Write error: ", err.Error())
		return "", err
	}

	l, err := (*c.ssdb).Read(resp)
	if err != nil {
		log.Println("Response error: ", err.Error())
		return "", err
	}

	return string(resp[:l]), nil
}
*/

/*
func (c *Conn) sendCmd() error {

	var err error
	multi := 0

	for k, v := range c.cmds {
		tot := len(v.cmds)
		var reply string

		if v.enabled {
			reply += fmt.Sprintf("*%d\r\n", tot)
			for _, e := range v.cmds {
				reply += fmt.Sprintf("$%d\r\n", e.length)
				reply += fmt.Sprintf("%s\r\n", e.cmd)
			}
			c.cmds[k].retval, err = c.writeCmd(reply)
			if err != nil {
				log.Println("Write Error: ", err)
				return err
			}
			if v.multi {
				multi++
			}
		} else if v.multi {
			c.cmds[k].retval = "+OK\r\n"
		} else {
			c.cmds[k].retval = fmt.Sprintf("*%d\r\n", multi)
			for multi > 0 {
				c.cmds[k].retval += c.cmds[k-multi].retval
				c.cmds[k-multi].retval = "+QUEUED\r\n"
				multi--
			}
		}
	}

	return nil
}
*/

/*
func (c *Conn) parser(lines string) int {
	if lines[ind][0] == '*' {
		num, _ := strconv.Atoi(string(lines[ind][1:]))
		n := 1
		for n <= num*2 {
			cmd := Cmd{}
			if lines[ind+n][0] == '$' {
				cmd.length, _ = strconv.Atoi(string(lines[ind+n][1:]))
				cmd.cmd = lines[ind+n+1]
				if cmd.cmd == "MULTI" {
					enabled = false
					state.multi = true
				} else if cmd.cmd == "EXEC" {
					enabled = false
					state.multi = false
				} else {
					enabled = true
				}
				n += 2
			} else {
				log.Println("Malformed cmd: ", lines[ind+n])
				n++
			}
			cmds.cmds = append(cmds.cmds, cmd)
		}
		ind += num * 2
		cmds.multi = state.multi
		cmds.enabled = enabled
		c.cmds = append(c.cmds, cmds)
	} else {
	}

	return ind + 1
}
*/

func (c *Conn) parser() error {

	fmt.Println("\nReading...")
	line, err := c.cBuf.ReadBytes('\n')
	if err != nil {
		return err
	}

	l := len(line)

	fmt.Println("Read: ", line, string(line))

	if line[l-2] != '\r' {
		return errors.New("Malformed cmd")
	}

	line = line[:l-2]

	sline := string(line)

	fmt.Println("Line: ", line, sline)

	fmt.Println("CmdStatus: ", getCmdStatus(c.cmdStatus))
	//	cmd := new(Cmd)
	switch line[0] {
	case '*':
		c.cmdStatus = CmdStar
		num, _ := strconv.Atoi(string(line[1:]))
		fmt.Println("*: ", num)
		for i := 0; i < num; i++ {
			err := c.parser()
			if err != nil {
				return err
			}
		}
		fmt.Println("Status: ", getCmdStatus(c.cmdStatus), getDovecotStatus(c.dovecotStatus))
	case '$':
		switch c.cmdStatus {
		case CmdStar:
			c.cmdStatus = CmdCmd
			err := c.parser()
			if err != nil {
				return err
			}
		case CmdCmd:
			c.cmdStatus = CmdParam
			err := c.parser()
			if err != nil {
				return err
			}
		default:
			fmt.Println("Protocol error: ", line)
			return errors.New("Protocol Error: " + sline)
		}
	case '+':
		c.cmdStatus = CmdPlus
	case ':':
		c.cmdStatus = CmdColon
	default:
		switch c.cmdStatus {
		case CmdCmd:
			fmt.Println("cmd: ", line)
			switch sline {
			case "GET":
				fmt.Println("GET")
				c.dovecotStatus = DovecotGet
				return nil
			case "MULTI":
				fmt.Println("MULTI")
				c.dovecotStatus = DovecotMulti
				return nil
			case "EXEC":
				fmt.Println("EXEC")
				c.dovecotStatus = DovecotExec
				return nil
			default:
				if c.dovecotStatus == DovecotMulti {
					c.cmds = append(c.cmds, sline)
				}
				fmt.Println("CMD: ", sline, getDovecotStatus(c.dovecotStatus))
			}
		case CmdParam:
		default:
		}
	}

	return nil
}

func (c *Conn) handleConn() error {
	log.Println("New Connection: ", c.conn.RemoteAddr())
	start := time.Now()
	startMsg := fmt.Sprintf("%v: Started %v", start, c.conn.RemoteAddr())

	for {
		err := c.parser()
		fmt.Println("Buffered: ", c.cBuf.Writer.Buffered())
		switch c.dovecotStatus {
		case DovecotGet:
			c.cBuf.Write([]byte("$1\r\n10\r\n"))
			c.cBuf.Flush()
			fmt.Println("Write $1")
			c.dovecotStatus = DovecotWait
		case DovecotMulti:
			c.cBuf.Write([]byte("+OK\r\n"))
			c.cBuf.Flush()
			fmt.Println("Write +OK")
		case DovecotExec:
			c.cBuf.Write([]byte(""))
			c.cBuf.Flush()
			fmt.Println("Write nil: ", c.cmds)
			c.dovecotStatus = DovecotExecReply
		case DovecotExecReply:
			c.dovecotStatus = DovecotWait
		default:
			fmt.Println("Default dovectoStatus")
			c.dovecotStatus = DovecotWait
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

/*
func (c *Conn) wrapCmd() error {

	c.parseCmd()

	log.Println("Cmds: ", c.cmds)

	err := c.sendCmd()
	if err != nil {
		log.Println("Write error: ", err.Error())
		return err
	}

	var reply string
	for k, _ := range c.cmds {
		reply += c.cmds[k].retval
	}

	log.Printf("Reply: %+v\n", strings.Split(reply, "\r\n"))

	_, err = (*c.conn).Write([]byte(reply))
	if err != nil {
		log.Println("Write reply error: ", err.Error())
		return err
	}

	return nil
}
*/
