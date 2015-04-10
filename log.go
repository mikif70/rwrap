// utils
package main

import (
	"fmt"
	"os"
)

type Log struct {
	debug   bool
	profile bool
	file    *os.File
}

func (l *Log) log() {
	fmt.Println("Start Log")
}

func NewLog() *Log {
	return &Log{
		debug:   false,
		profile: false,
	}
}
