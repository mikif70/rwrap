// connections
package main

import (
	"log"
	"net"
)

func printLog(str string, inf ...interface{}) {

	log.Printf(str, inf...)

}

func debugLog(str string, inf ...interface{}) {

	if cfg.debug {
		log.Printf(str, inf...)
	}

}

func errorLog(str string, inf ...interface{}) {

	log.Printf(str, inf...)

}

func listen() *net.TCPListener {
	printLog("Listen: %+v - DB: %+v\n", cfg.wrapAddr, cfg.ssdbAddr)
	ln, err := net.ListenTCP("tcp", cfg.wrapAddr)
	if err != nil {
		log.Fatalln("Listen err: ", err.Error())
	}

	return ln
}

func ssdbConnect(count int) (*net.TCPConn, error) {
	ssdb, err := net.DialTCP("tcp", nil, cfg.ssdbAddr)
	if err != nil {
		errorLog("Dial err: %+v\n", err.Error())
		if count > 1 {
			ssdb, err = ssdbConnect(count - 1)
			return ssdb, err
		} else {
			return nil, err
		}
	}
	return ssdb, nil
}
