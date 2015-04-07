// conn
package main

import (
	"errors"
	"fmt"
	"strconv"
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

func (c *Conn) parseCmd() error {

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

	fmt.Println("Line: ", line, string(line))

	//	cmd := new(Cmd)
	switch line[0] {
	case '*':
		num, _ := strconv.Atoi(string(line[1:]))
		fmt.Println("*: ", num)
		for i := 0; i < num*2; i++ {
			c.parseCmd()
		}
	case '$':
		num, _ := strconv.Atoi(string(line[1:]))
		fmt.Println("$: ", num)
	case '+':
	case ':':
	default:
		fmt.Println("cmd: ", line)
		c.cmds = append(c.cmds, Cmd{cmd: string(line)})
		if c.multi {
			if string(line) == "EXEC" {
				c.cBuf.WriteString("+OK\r\n")
				c.multi = false
				c.cBuf.Flush()
			} else {
				c.cBuf.WriteString("+QUEUED\r\n")
			}
		} else {
			c.cBuf.WriteString("+OK\r\n")
		}
		if string(line) == "MULTI" {
			c.multi = true
		}
	}

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
