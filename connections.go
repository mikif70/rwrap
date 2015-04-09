// connections
package main

import (
	"log"
	"net"
)

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
	//	ssdb.SetDeadline(time.Now().Add(time.Nanosecond * 200000000))
	//	config.ssdbConn = ssdb
	return ssdb, nil
}
