// connections
package main

import (
	"log"
	"net"
)

func listen() *net.TCPListener {
	log.Printf("Listen: %+v\n", cfg.wrapAddr)
	ln, err := net.ListenTCP("tcp", cfg.wrapAddr)
	if err != nil {
		log.Fatalln("Listen err: ", err.Error())
	}

	return ln
}

func ssdbConnect(count int) (*net.TCPConn, error) {
	ssdb, err := net.DialTCP("tcp", nil, cfg.ssdbAddr)
	if err != nil {
		log.Println("Dial err: ", err.Error())
		if count > 1 {
			ssdb, err = ssdbConnect(count - 1)
			return ssdb, err
		} else {
			return nil, err
		}
	}
	return ssdb, nil
}
