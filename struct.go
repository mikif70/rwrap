// struct
package main

import ()

type Status int

const (
	CmdStar   Status = iota // *
	CmdDollar               // $
	CmdPlus                 // +
	CmdColon                // :
	CmdCmd                  // cmd
	CmdParam                // params
)

func getCmdStatus(status Status) string {
	switch status {
	case CmdStar:
		return "*"
	case CmdDollar:
		return "$"
	case CmdPlus:
		return "+"
	case CmdColon:
		return ":"
	case CmdCmd:
		return "_"
	}

	return ""
}
