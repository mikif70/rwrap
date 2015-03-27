//
package main

import (
	"flag"
	"fmt"
	"net"
)

type Config struct {
	wrapUrl  string
	wrapAddr *net.TCPAddr
	cmd      string
	user     string
}

const (
	NL = "\r\n"
)

var (
	//	ErrInvalidSyntax = error.New("resp: invalid syntax")
	config = &Config{
		wrapUrl: "127.0.0.1:6380",
		cmd:     "deliver",
		user:    "dtest01@tiscali.it",
	}
)

func init() {
	flag.StringVar(&config.wrapUrl, "s", config.wrapUrl, "server ip:port")
	flag.StringVar(&config.cmd, "c", config.cmd, "cmd: deliver|recalc|get")
	flag.StringVar(&config.user, "u", config.user, "user")
}

func main() {
	flag.Parse()

	config.wrapAddr, _ = net.ResolveTCPAddr("tcp", config.wrapUrl)

	fmt.Println(config)

	wrap, err := net.DialTCP("tcp", nil, config.wrapAddr)
	if err != nil {
		fmt.Println("Dial wrap err: ", err.Error())
		return
	}
	defer wrap.Close()

	var msg []string

	switch config.cmd {
	case "get":
		msg = []string{
			"*2" + NL + "$3" + NL + "GET" + NL + "$32" + NL + config.user + "/quota/storage" + NL,
			"*2" + NL + "$3" + NL + "GET" + NL + "$33" + NL + config.user + "/quota/messages" + NL,
		}
	case "recalc":
		msg = []string{
			"*1" + NL + "$5" + NL + "MULTI" + NL + "*2" + NL + "$3" + NL + "DEL" + NL + "$32" + NL +
				config.user + "/quota/storage" + NL + "*2" + NL + "$3" + NL + "DEL" + NL + "$33" + NL +
				config.user + "/quota/messages" + NL + "*3" + NL + "$3" + NL + "SET" + NL + "$32" + NL +
				config.user + "/quota/storage" + NL + "$9" + NL + "190544684" + NL + "*3" + NL + "$3" + NL +
				"SET" + NL + "$33" + NL + config.user + "/quota/messages" + NL + "$4" + NL + "2963" + NL +
				"*1" + NL + "$4" + NL + "EXEC" + NL,
		}
	default:
		msg = []string{
			"*1" + NL + "$5" + NL + "MULTI" + NL + "*3" + NL + "$6" +
				NL + "INCRBY" + NL + "$32" + NL + config.user + "/quota/storage" +
				NL + "$5" + NL + "-1622" + NL + "*3" + NL + "$6" + NL + "INCRBY" + NL + "$33" +
				NL + config.user + "/quota/messages" + NL + "$2" + NL + "-2" + NL + "*1" +
				NL + "$4" + NL + "EXEC" + NL + "*2" + NL + "$3" + NL + "GET" + NL + "$32" +
				NL + config.user + "/quota/storage" + NL,
		}
	}

	for _, m := range msg {
		wrap.Write([]byte(m))
		buf := make([]byte, 512)
		n, err := wrap.Read(buf)
		if err != nil {
			fmt.Printf("Error: ", err)
			continue
		}
		fmt.Printf("Reply: %+v\n%+v\n", buf[:n], string(buf[:n]))
	}
}
