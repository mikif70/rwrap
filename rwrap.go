// rwrap

/*
 TODO:
	check ssdb connection:

*/

package main

import (
	//	"bytes"
	"bufio"
	"flag"
	"fmt"
	//	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime/pprof"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Config struct {
	wrapUrl    string
	ssdbUrl    string
	wrapAddr   *net.TCPAddr
	ssdbAddr   *net.TCPAddr
	logfile    string
	logOut     *os.File
	logFlags   int
	logger     *log.Logger
	cpuprofile string
	debug      bool
	//	ssdbConn *net.TCPConn
}

type Conn struct {
	conn   *net.TCPConn
	ssdb   *net.TCPConn
	cBuf   *bufio.ReadWriter
	sBuf   *bufio.ReadWriter
	logger *log.Logger
	buf    []string
	cmds   []Cmds
	reply  []string
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
		wrapUrl:    "0.0.0.0:6380",
		ssdbUrl:    "10.39.80.182:8888",
		cpuprofile: "",
		logfile:    "",
		debug:      false,
	}

	state = &State{}
)

func (c *Conn) writeCmd(buf string) (string, error) {
	resp := make([]byte, 256)

	_, err := (*c.ssdb).Write([]byte(buf))
	if err != nil {
		c.logger.Println("Write error: ", err.Error())
		return "", err
	}

	l, err := (*c.ssdb).Read(resp)
	if err != nil {
		c.logger.Println("Response error: ", err.Error())
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
				c.logger.Println("Write Error: ", err)
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

func (c *Conn) parser(lines []string) int {
	ind := 0
	cmds := new(Cmds)
	enabled := true
	if lines[ind][0] == '*' {
		num, _ := strconv.Atoi(string(lines[ind][1:]))
		n := 1
		for n <= num*2 {
			cmd := new(Cmd)
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
				c.logger.Println("Malformed cmd: ", lines[ind+n])
				n++
			}
			cmds.cmds = append(cmds.cmds, *cmd)
		}
		ind += num * 2
		cmds.multi = state.multi
		cmds.enabled = enabled
		c.cmds = append(c.cmds, *cmds)
	} else {
		c.logger.Println("Malformed cmd")
	}

	return ind + 1
}

func (c *Conn) parseCmd() error {

	llen := len(c.buf)

	ind := 0
	for {
		ind += c.parser(c.buf[ind:])
		if ind >= llen-1 {
			break
		}
	}

	return nil
}

func (c *Conn) wrapCmd() error {

	c.parseCmd()

	c.logger.Println("Cmds: ", c.cmds)

	err := c.sendCmd()
	if err != nil {
		c.logger.Println("Write error: ", err.Error())
		return err
	}

	var reply string
	for k, _ := range c.cmds {
		reply += c.cmds[k].retval
	}

	c.logger.Printf("Reply: %+v\n", strings.Split(reply, "\r\n"))

	_, err = (*c.conn).Write([]byte(reply))
	if err != nil {
		c.logger.Println("Write reply error: ", err.Error())
		return err
	}

	return nil
}

/*
func (c *Conn) flush() {
	c.buf = ""
	c.cmds = make([]Cmds, 0)
	c.reply = make([]string, 0)
}
*/

func manageConnection(conn *net.TCPConn) {

	var err error

	c := new(Conn)
	c.conn = conn
	//	c.conn.SetReadDeadline(time.Now().Add(time.Nanosecond * 1000000 * 50))
	c.logger = log.New(config.logOut, conn.RemoteAddr().String()+"::", config.logFlags)
	//	defer c.flush()

	c.logger.Println("New Connection: ", conn.RemoteAddr())
	start := time.Now()
	startMsg := fmt.Sprintf("%v: Started %v", start, conn.RemoteAddr())
	defer conn.Close()

	c.logger.Println("Connecting SSDB....")
	c.ssdb, err = ssdbConnect(3)
	if err != nil {
		c.logger.Println("SSDB err: ", err.Error())
		conn.Close()
		return
	}
	defer c.ssdb.Close()

	c.cBuf = bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	c.sBuf = bufio.NewReadWriter(bufio.NewReader(c.ssdb), bufio.NewWriter(c.ssdb))
	c.buf = make([]string, 0)

	scanner := bufio.NewScanner(c.conn)
	counter := 1
	for {
		for {
			if ok := scanner.Scan(); !ok {
				break
			}
			line := scanner.Text()
			c.logger.Printf("Line (%d): %s\n", counter, line)
			c.buf = append(c.buf, line)
			counter++
		}
		if err := scanner.Err(); err != nil {
			c.logger.Printf("scanner error: %v\n\n", err.Error())
		}
		err = c.wrapCmd()
		if err != nil {
			c.logger.Printf("wrapCmd error: %v\n\n", err.Error())
			break
		}
	}

	c.logger.Printf("%s - executed %d cmds in %v\n\n", startMsg, counter, time.Since(start))
}

func listen() *net.TCPListener {
	config.logger.Printf("Listen: %+v\n", config.wrapAddr)
	ln, err := net.ListenTCP("tcp", config.wrapAddr)
	if err != nil {
		config.logger.Fatalln("Listen err: ", err.Error())
	}

	return ln
}

func ssdbConnect(count int) (*net.TCPConn, error) {
	ssdb, err := net.DialTCP("tcp", nil, config.ssdbAddr)
	if err != nil {
		config.logger.Println("Dial err: ", err.Error())
		if count > 1 {
			ssdb, err = ssdbConnect(count - 1)
			return ssdb, err
		} else {
			return nil, err
		}
	}
	ssdb.SetDeadline(time.Now().Add(time.Nanosecond * 200000000))
	//	config.ssdbConn = ssdb
	return ssdb, nil
}

func init() {
	flag.StringVar(&config.ssdbUrl, "s", config.ssdbUrl, "ssdb ip:port")
	flag.StringVar(&config.wrapUrl, "l", config.wrapUrl, "listen ip:port")
	flag.StringVar(&config.cpuprofile, "cpuprofile", config.cpuprofile, "write cpu profile to file")
	flag.StringVar(&config.logfile, "log", config.logfile, "write log to file")
	flag.BoolVar(&config.debug, "debug", config.debug, "activate debug")
}

func main() {

	var err error

	flag.Parse()

	config.ssdbAddr, _ = net.ResolveTCPAddr("tcp", config.ssdbUrl)
	config.wrapAddr, _ = net.ResolveTCPAddr("tcp", config.wrapUrl)

	if config.cpuprofile != "" {
		fProfile, err := os.OpenFile(config.cpuprofile, os.O_CREATE, 0666)
		if err != nil {
			log.Fatal("Failed to open profile file", err)
		}
		pprof.StartCPUProfile(fProfile)
		defer pprof.StopCPUProfile()
		defer fProfile.Close()
	}

	if config.logfile != "" {
		config.logOut, err = os.OpenFile(config.logfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal("Failed to open log file", err)
		}
		defer config.logOut.Close()
	} else {
		config.logOut = os.Stdout
	}

	if config.debug {
		config.logFlags = log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile

	} else {
		config.logFlags = log.Ldate | log.Ltime | log.Lshortfile
	}

	config.logger = log.New(config.logOut, "MAIN::", config.logFlags)

	config.logger.Println(config)

	ln := listen()
	defer ln.Close()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGHUP, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT)

	go func(ln *net.TCPListener) {
		for {
			conn, err := ln.AcceptTCP()
			if err != nil {
				config.logger.Println("Accept err: ", err.Error())
				continue
			}

			go manageConnection(conn)
		}
	}(ln)

	for sig := range sigchan {
		config.logger.Println("Signal: ", sig)
		break
		//		os.Exit(0)
	}
}
