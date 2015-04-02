// rwrap

/*
 TODO:
	check ssdb connection:

*/

package main

import (
	//	"bytes"
	"flag"
	"fmt"
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
	cpuprofile string
	logfile    string
	debug      bool
	//	ssdbConn *net.TCPConn
}

type Conn struct {
	conn  net.Conn
	ssdb  net.Conn
	buf   string
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

	_, err := c.ssdb.Write([]byte(buf))
	if err != nil {
		log.Println("Write error: ", err.Error())
		return "", err
	}

	l, err := c.ssdb.Read(resp)
	if err != nil {
		log.Println("Response error: ", err.Error())
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
		log.Println("Malformed cmd")
	}

	return ind + 1
}

func (c *Conn) parseCmd() error {

	lines := strings.Split(c.buf, "\r\n")

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

	_, err = c.conn.Write([]byte(reply))
	if err != nil {
		log.Println("Write reply error: ", err.Error())
		return err
	}

	return nil
}

func manageConnection(conn *net.TCPConn) {

	start := time.Now()
	startMsg := fmt.Sprintf("%v: Started %v", start, conn.RemoteAddr())
	defer conn.Close()

	log.Println("Connecting SSDB....")

	ssdb, err := ssdbConnect(3)
	if err != nil {
		log.Println("SSDB err: ", err.Error())
		conn.Close()
		return
	}
	defer ssdb.Close()

	counter := 1
	for {
		c := &Conn{
			conn: conn,
			ssdb: ssdb,
		}

		buf := make([]byte, 512)

		n, err := c.conn.Read(buf)
		if err != nil {
			if err.Error() != "EOF" {
				log.Println("Read error: ", err.Error())
			} else {
				log.Println("EOF: ", err.Error())
			}

			break
		}

		log.Printf("Buffer (%d): %+v\n", counter, buf[:n])
		c.buf = string(buf[:n])
		err = c.wrapCmd()
		if err != nil {
			break
		}
		counter++
	}
	log.Printf("%s - executed %d cmds in %v\n\n", startMsg, counter, time.Since(start))
}

func listen() *net.TCPListener {
	log.Printf("Listen: %+v\n", config.wrapAddr)
	ln, err := net.ListenTCP("tcp", config.wrapAddr)
	if err != nil {
		log.Fatalln("Listen err: ", err.Error())
	}

	return ln
}

func ssdbConnect(count int) (*net.TCPConn, error) {
	ssdb, err := net.DialTCP("tcp", nil, config.ssdbAddr)
	if err != nil {
		log.Println("Dial err: ", err.Error())
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
		fLog, err := os.OpenFile(config.logfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal("Failed to open log file", err)
		}
		defer fLog.Close()
		log.SetOutput(fLog)
	} else {
		log.SetOutput(os.Stdout)
	}

	if config.debug {
		log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	} else {
		log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	}

	log.Println(config)

	ln := listen()
	defer ln.Close()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGHUP, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT)

	go func() {
		for {
			log.Println("Waiting...")
			conn, err := ln.AcceptTCP()
			if err != nil {
				log.Println("Accept err: ", err.Error())
				continue
			}

			go manageConnection(conn)
		}
	}()

	for sig := range sigchan {
		log.Println("Signal: ", sig)
		break
		//		os.Exit(0)
	}
}
