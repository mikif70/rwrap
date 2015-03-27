// rwrap

package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	wrapUrl  string
	ssdbUrl  string
	wrapAddr *net.TCPAddr
	ssdbAddr *net.TCPAddr
	ssdbConn *net.TCPConn
}

type Conn struct {
	conn  net.Conn
	ssdb  net.Conn
	cmds  []Cmds
	reply []string
}

type State struct {
	multi bool
}

type Cmd struct {
	length int
	cmd    string
}

type Cmds struct {
	cmds    []Cmd
	retval  string
	enabled bool
	multi   bool
}

type Empty struct{}

var (
	config = &Config{
		wrapUrl: "0.0.0.0:6380",
		ssdbUrl: "10.39.80.182:8888",
	}

	state = &State{}
)

func (c *Conn) parser(lines []string) int {
	ind := 0
	cmds := Cmds{}
	enabled := true
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
				fmt.Println("Malformed cmd: ", lines[ind+n])
				n++
			}
			cmds.cmds = append(cmds.cmds, cmd)
		}
		ind += num * 2
		cmds.multi = state.multi
		cmds.enabled = enabled
		c.cmds = append(c.cmds, cmds)
	} else {
		fmt.Println("Malformed cmd")
	}

	return ind + 1
}

func (c *Conn) parseCmd(buf string) error {

	lines := strings.Split(buf, "\r\n")

	lines = lines[:len(lines)-1]
	llen := len(lines)

	loop := true
	ind := 0
	for loop {
		ind += c.parser(lines[ind:])
		if ind >= llen-1 {
			loop = false
		}
	}

	return nil
}

func (c *Conn) writeCmd(buf string) (string, error) {
	resp := make([]byte, 256)

	r, err := c.ssdb.Write([]byte(buf))
	if err != nil {
		fmt.Println("Write error: ", err.Error())
		return "", err
	}
	fmt.Println("Write: ", r, err, len(buf))
	l, err := c.ssdb.Read(resp)
	if err != nil {
		fmt.Println("Response error: ", err.Error())
		return "", err
	}

	return string(resp[:l]), nil
}

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
				fmt.Println("Write Error: ", err)
				return err
			}
			if v.multi {
				fmt.Println("Multi ?: ", v, multi)
				multi++
			}
		} else if v.multi {
			c.cmds[k].retval = "+OK\r\n"
		} else {
			c.cmds[k].retval = fmt.Sprintf("*%d\r\n", multi)
			fmt.Println("Multi: ", multi)
			for multi > 0 {
				fmt.Println("Loop :", multi)
				c.cmds[k].retval += c.cmds[k-multi].retval
				c.cmds[k-multi].retval = "+QUEUED\r\n"
				multi--
			}
			fmt.Println("end loop: ", multi)
		}
	}

	return nil
}

func (c *Conn) wrapCmd(buf string) error {

	fmt.Println(buf)

	c.parseCmd(buf)

	if len(c.cmds) > 0 {
		fmt.Println("Cmds: ", c.cmds)
	}

	err := c.sendCmd()
	if err != nil {
		fmt.Println("Write error: ", err.Error())
		return err
	}

	var reply string
	for k, _ := range c.cmds {
		reply += c.cmds[k].retval
	}

	fmt.Println(reply)

	r, err := c.conn.Write([]byte(reply))
	if err != nil {
		fmt.Println("Write reply error: ", err.Error())
		return err
	}
	fmt.Println("Write reply: ", r, err, len(reply), string(reply))

	return nil
}

func manageConnection(conn *net.TCPConn) {

	fmt.Println("Manage conn: ", conn)
	defer conn.Close()

	counter := 1
	for {
		c := &Conn{
			conn: conn,
			ssdb: config.ssdbConn,
		}

		start := time.Now()
		fmt.Printf("Reading buf: %d\n", counter)

		buf := make([]byte, 512)

		n, err := c.conn.Read(buf)
		if err != nil {
			fmt.Println("Read error: ", err.Error())
			break
		}
		err = c.wrapCmd(string(buf[:n]))
		if err != nil {
			break
		}
		fmt.Printf("wrapCmd %d: %v\n\n", counter, time.Since(start))
		counter++
	}
}

func listen() *net.TCPListener {
	fmt.Printf("Listen: %+v\n", config.wrapAddr)
	ln, err := net.ListenTCP("tcp", config.wrapAddr)
	if err != nil {
		fmt.Println("Listen err: ", err.Error())
		os.Exit(-1)
	}

	return ln
}

func ssdbConnect() *net.TCPConn {
	ssdb, err := net.DialTCP("tcp", nil, config.ssdbAddr)
	if err != nil {
		fmt.Println("Dial err: ", err.Error())
		os.Exit(-2)
	}

	config.ssdbConn = ssdb
	return ssdb
}

func init() {
	flag.StringVar(&config.ssdbUrl, "s", config.ssdbUrl, "ssdb ip:port")
	flag.StringVar(&config.wrapUrl, "l", config.wrapUrl, "listen ip:port")
}

func main() {
	flag.Parse()

	config.ssdbAddr, _ = net.ResolveTCPAddr("tcp", config.ssdbUrl)
	config.wrapAddr, _ = net.ResolveTCPAddr("tcp", config.wrapUrl)

	fmt.Println(config)

	ln := listen()
	defer ln.Close()

	ssdb := ssdbConnect()
	defer ssdb.Close()

	for {
		conn, err := ln.AcceptTCP()
		if err != nil {
			fmt.Println("Accept err: ", err.Error())
			continue
		}

		go manageConnection(conn)
	}
}
