// rwrapBench
package main

import (
	"bytes"
	"crypto/sha1"
	"flag"
	"fmt"
	"math/rand"
	"os/exec"
	"time"
)

type Config struct {
	repeat int
	sleep  time.Duration
	cmd    string
}

type Bench struct {
	params   []string
	countErr float64
	countOk  float64
	timeErr  float64
	cmd      string
	reply    [20]byte
	startS   time.Time
	endS     time.Duration
}

var (
	_DOVEADM = "/usr/bin/doveadm"
	_USERS   = []string{"miki", "mfadda", "user1", "user2", "test1", "test2"}
	_OPTS    = []string{"-v", "quota"} // , "get", "-u"}
	config   = &Config{
		repeat: 100,
		sleep:  10 * time.Nanosecond * 1000000,
		cmd:    "get",
	}
	cmds  = []string{"get", "recalc"}
	reply = map[string][20]byte{
		"get":    sha1.Sum([]byte{85, 115, 101, 114, 32, 113, 117, 111, 116, 97, 32, 83, 84, 79, 82, 65, 71, 69, 32, 32, 32, 32, 32, 48, 32, 49, 48, 52, 56, 53, 55, 54, 48, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 48, 10, 85, 115, 101, 114, 32, 113, 117, 111, 116, 97, 32, 77, 69, 83, 83, 65, 71, 69, 32, 32, 32, 32, 32, 48, 32, 32, 32, 32, 32, 32, 32, 32, 45, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 48, 10}),
		"recalc": sha1.Sum([]byte{}),
	}
)

func init() {
	flag.IntVar(&config.repeat, "r", config.repeat, "iterations")
	flag.DurationVar(&config.sleep, "s", config.sleep, "sleep")
	flag.StringVar(&config.cmd, "c", config.cmd, "cmd g[et]|r[ecalc]|a[ll]")
}

func execute(bench *Bench) {
}

func main() {
	flag.Parse()

	bench := &Bench{
		countErr: 0.0,
		countOk:  0.0,
		timeErr:  0.0,
	}

	rand.Seed(time.Now().UnixNano())
	start := time.Now()
	for int(bench.countErr+bench.countOk) < config.repeat {
		bench.startS = time.Now()

		bench.params = make([]string, 0)

		switch config.cmd {
		case "r", "recalc":
			bench.cmd = "recalc"
		case "a", "all":
			bench.cmd = cmds[rand.Intn(len(cmds))]
		default:
			bench.cmd = "get"
		}

		bench.reply = reply[bench.cmd]
		bench.params = append(bench.params, _OPTS...)
		bench.params = append(bench.params, bench.cmd)
		bench.params = append(bench.params, "-u")
		bench.params = append(bench.params, _USERS[rand.Intn(len(_USERS))])
		fmt.Println("Running: ", bench.params)
		var out bytes.Buffer
		var err bytes.Buffer
		cmd := exec.Command(_DOVEADM, bench.params...)
		cmd.Stdout = &out
		cmd.Stderr = &err
		er := cmd.Run()
		if er != nil {
			fmt.Println("Exec Error: ", er.Error())
			bench.countErr++
			bench.endS = time.Since(bench.startS)
			bench.timeErr += bench.endS.Seconds()
		}

		if bench.reply != sha1.Sum(out.Bytes()) {
			fmt.Println("Output: ", out.String())
			fmt.Println("Error: ", err.String())
			bench.countErr++
			bench.endS = time.Since(bench.startS)
			bench.timeErr += bench.endS.Seconds()
			return
		} else {
			bench.countOk++
		}
		time.Sleep(config.sleep)
	}
	end := time.Since(start)
	fmt.Println("End: ", end)
	fmt.Printf("OK: %d - Err: %d - %%Err: %.2f - Req/s: %.2f \n", int(bench.countOk), int(bench.countErr), (bench.countErr/bench.countOk)*100.0, bench.countOk/(end.Seconds()-bench.timeErr))
}
