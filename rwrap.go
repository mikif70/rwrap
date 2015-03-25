// rwrap

package main

import (
	"bufio"
	//	"errors"
	"bytes"
	"flag"
	"fmt"
	//	"io"
	"net"
	"os"
	"strconv"
	"time"
)

type Config struct {
	wrapUrl  string
	ssdbUrl  string
	wrapAddr *net.TCPAddr
	ssdbAddr *net.TCPAddr
}

type Conn struct {
	conn      net.Conn
	bufRead   *bufio.Reader
	bufWrite  *bufio.Writer
	wrapRead  *bufio.Reader
	wrapWrite *bufio.Writer
}

var (
	config = &Config{
		wrapUrl: "0.0.0.0:6380",
		ssdbUrl: "10.39.80.182:8888",
	}
)

func parseInt(data []byte) int {
	//	fmt.Println(string(data))
	num, err := strconv.Atoi(string(data))
	if err != nil {
		fmt.Println("Conversion error: ", err.Error())
		return 0
	}
	return num
}

func myScanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		fmt.Println("EOF")
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		fmt.Println("Line: ", i+1, string(data[:i+1]))
		return i + 1, data[:i+1], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		fmt.Println("EOF with data")
		return len(data), data, nil
	}
	// Request more data.
	fmt.Println("more data")
	return 0, nil, nil
}

func parser(buf *bufio.Reader) ([]byte, error) {
	scan := bufio.NewScanner(buf)
	scan.Split(myScanLines)
	num := 0
	lines := make([][]byte, 0)
	retval := make([][]byte, 0)
	cmdLine := true
	cmd := ""
	loop := true
	multi := false
	for loop {
		fmt.Printf("New Loop: loop: %v - multi: %v - cmd: %s - cmdLine: %v - num: %d\n", loop, multi, cmd, cmdLine, num)
		if scan.Scan() {
			switch scan.Bytes()[0] {
			case ':':
			case '+':
			case '-':
			case '$':
				if num == 0 {
					num = 1
					cmdLine = false
				}
			case '*':
				ln := len(scan.Bytes())
				num = parseInt(scan.Bytes()[1:ln-2]) * 2
				cmdLine = true
			default:
				if cmdLine {
					ln := len(scan.Bytes())
					cmd = string(scan.Bytes()[:ln-2])
					cmdLine = false
				}
			}

			lines = append(lines, scan.Bytes())
			if num == 0 {
				if cmd == "MULTI" {
					multi = true
				} else if cmd == "EXEC" {
					loop = false
				} else {
					retval = append(retval, bytes.Join(lines, []byte("")))
					if !multi {
						loop = false
					}
				}
				lines = make([][]byte, 0)
			} else {
				num--
			}
		} else {
			loop = false
		}
	}

	if err := scan.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading buffer: ", err)
		return nil, err
	}

	return bytes.Join(retval, []byte("")), nil
}

func (c *Conn) wrapCmd() error {

	loop := true

	for loop {
		ret, err := parser(c.bufRead)
		if err != nil {
			loop = false
			continue
		}
		fmt.Println("Writing: ", ret)
		num, err := c.wrapWrite.Write(ret)
		fmt.Println("Written: ", num, err)
		c.wrapWrite.Flush()
		ret, _ = parser(c.wrapRead)
		fmt.Println("Recv: ", string(ret))
		num, err = c.bufWrite.Write(ret)
		c.bufWrite.Flush()
	}

	return nil
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

	fmt.Printf("Listen: %+v\n", config.wrapAddr)
	ln, err := net.ListenTCP("tcp", config.wrapAddr)
	if err != nil {
		fmt.Println("Listen err: ", err.Error())
		return
	}
	defer ln.Close()

	file, err := os.Create("dovecot-redis.log")
	if err != nil {
		fmt.Println("File err: ", err.Error())
		return
	}
	defer file.Close()

	ssdb, err := net.DialTCP("tcp", nil, config.ssdbAddr)
	if err != nil {
		fmt.Println("Dial err: ", err.Error())
		return
	}
	defer ssdb.Close()

	for {
		conn, err := ln.AcceptTCP()
		if err != nil {
			fmt.Println("Accept err: ", err.Error())
			continue
		}

		go func() {
			conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			defer conn.Close()

			c := &Conn{
				conn:      conn,
				bufRead:   bufio.NewReader(conn),
				bufWrite:  bufio.NewWriter(conn),
				wrapRead:  bufio.NewReader(ssdb),
				wrapWrite: bufio.NewWriter(ssdb),
			}

			start := time.Now()
			err := c.wrapCmd()
			fmt.Println("wrapCmd: ", time.Since(start))

			if err != nil {
				fmt.Println(err)
			}
		}()
	}
}
