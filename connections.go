// connections
package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"time"
)

func listen() *net.TCPListener {
	log.Printf("Listen: %+v\n", config.wrapAddr)
	ln, err := net.ListenTCP("tcp", config.wrapAddr)
	if err != nil {
		log.Fatalln("Listen err: ", err.Error())
	}

	return ln
}

/*
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
*/

func manageConnection(conn *net.TCPConn) {

	log.Println("New Connection: ", conn.RemoteAddr())
	start := time.Now()
	startMsg := fmt.Sprintf("%v: Started %v", start, conn.RemoteAddr())
	defer conn.Close()

	/*
		log.Println("Connecting SSDB....")
		ssdb, err := ssdbConnect(3)
		if err != nil {
			log.Println("SSDB err: ", err.Error())
			conn.Close()
			return
		}
		defer ssdb.Close()
	*/

	c := &Conn{
		conn:  conn,
		cBuf:  bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)),
		cmds:  make([]Cmd, 0),
		multi: false,
	}

	//	c.sBuf = bufio.NewReadWriter(bufio.NewReader(ssdb), bufio.NewWriter(ssdb))

	for {
		err := c.parseCmd()
		fmt.Println("Buffered: ", c.cBuf.Writer.Buffered())
		c.cBuf.Flush()
		if err != nil {
			fmt.Println("Error: ", err)
			break
		}
	}

	fmt.Println("Cmds: ", c.cmds)

	/*
		counter := 1
		for {

			buf := make([]byte, 1024)
			n, err := c.conn.Read(buf)
			if err != nil {
				if err.Error() != "EOF" {
					log.Println("Read error: ", err.Error())
				} else {
					log.Println("EOF: ", err.Error())
				}

				break
			}

			log.Printf("Buffer (%d): %d -> %+v\n", counter, n, buf[:n])
			c.buf = string(buf[:n])
			err = c.wrapCmd()
			if err != nil {
				break
			}

			counter++
		}
	*/
	log.Printf("%s - executed cmds in %v\n\n", startMsg, time.Since(start))
}
