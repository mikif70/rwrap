// connections
package main

import (
	"fmt"
	"log"
	"net"
	"runtime"
)

func getFuncName() string {
	a, _, c, _ := runtime.Caller(2)
	e := runtime.FuncForPC(a).Name()

	name := fmt.Sprintf("%s[%d]: ", e, c)

	return name
}

func printLog(str string, inf ...interface{}) {

	log.Printf(getFuncName()+str, inf...)

}

func debugLog(str string, inf ...interface{}) {

	if cfg.debug {
		log.Printf(getFuncName()+str, inf...)
	}

}

func errorLog(str string, inf ...interface{}) {

	log.Printf(getFuncName()+str, inf...)
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
