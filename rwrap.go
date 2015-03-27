// rwrap

package main

import (
	// "bufio"
	//	"errors"
	// "bytes"
	"flag"
	"fmt"
	//"io"
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
	multi []bool
	reply []string
	count int
}

type Cmd struct {
	length byte
	cmd    string
}

type Cmds struct {
	cmds []Cmd
}

var (
	config = &Config{
		wrapUrl: "0.0.0.0:6380",
		ssdbUrl: "10.39.80.182:8888",
	}
)

func (c *Conn) parser(lines []string) int {
	//	fmt.Println("Lines: ", lines)
	ind := 0
	multi := false
	switch lines[ind][0] {
	case '*':
		//		fmt.Println("*: ", string(lines[ind]))
		num, _ := strconv.Atoi(string(lines[ind][1:]))
		cmd := make([]string, 0)
		cmd = append(cmd, string(lines[ind]))
		for n := 1; n <= num*2; n++ {
			switch lines[ind+n][0] {
			case '$':
				cmd = append(cmd, lines[ind+n])
			default:
				if lines[ind+n] == "MULTI" {
					multi = true
					c.multi = append(c.multi, true)
				} else if lines[ind+n] == "EXEC" {
					multi = true
					c.multi = append(c.multi, false)
				}
				cmd = append(cmd, lines[ind+n])

			}
			//			fmt.Println("For: ", ind, n, string(lines[ind+n]))
		}
		ind += num * 2
		if !multi {
			if c.multi {
				c.count++
				c.cmds[cmd] = true
			} else {
				c.cmds[cmd] = false
			}
		}

	default:
		fmt.Println("Malformed cmd")
	}

	//	fmt.Println("return ind: ", ind+1)

	return ind + 1
}

func (c *Conn) parseCmd(buf string) error {

	//	cmds := make([]string, 0)

	c.count = 0
	c.multi = false
	c.cmds = make(map[string]bool)

	lines := strings.Split(buf, "\r\n")

	lines = lines[:len(lines)-1]
	llen := len(lines)

	loop := true
	ind := 0
	for loop {
		//		fmt.Println("ind: ", ind, llen)
		ind += c.parser(lines[ind:])
		if ind >= llen-1 {
			loop = false
		}
	}

	fmt.Printf("Cmds: %v\n", c.cmds)

	return nil
}

func (c *Conn) wrapCmd() {

	fmt.Println("Reading buf: ")

	buf := make([]byte, 512)

	n, err := c.conn.Read(buf)
	if err != nil {
		fmt.Println("Read error: ", err.Error())
		return
	}

	//	fmt.Println("Read: ", n, buf[:n])

	c.parseCmd(string(buf[:n]))

	resp := make([]byte, 128)
	if c.count > 0 {
		c.reply = append(c.reply, "+OK\r\n")
		for i := 0; i < c.count; i++ {
			c.reply = append(c.reply, "+QUEUED\r\n")
		}
		fmt.Printf("multi: %d\n", c.count)
	}
	for k, v := range c.cmds {
		fmt.Printf("cmd: %v -> %+v\n", v, k)
		buf := []byte(strings.Join(k.([]string), "\r\n") + "\r\n")
		r, err := c.ssdb.Write(buf)
		if err != nil {
			fmt.Println("Write error: ", err.Error())
			return
		}
		fmt.Println("Write: ", r, err, len(buf))
		l, err := c.ssdb.Read(resp)
		if err != nil {
			fmt.Println("Response error: ", err.Error())
			return
		}
		c.reply = append(c.reply, string(resp[:l]))
	}

	retbuf := []byte(strings.Join(c.reply, "\r\n"))
	r, err := c.conn.Write(retbuf)
	if err != nil {
		fmt.Println("Write reply error: ", err.Error())
		return
	}
	fmt.Println("Write reply: ", r, err, len(retbuf), string(retbuf))

}

func manageConnection(conn *net.TCPConn) {

	fmt.Println("Manage conn: ", conn)
	defer conn.Close()

	c := &Conn{
		conn: conn,
		ssdb: config.ssdbConn,
	}

	start := time.Now()
	c.wrapCmd()
	fmt.Println("wrapCmd: ", time.Since(start))
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
