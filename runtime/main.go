package main

import (
	"github.com/xmazu/oexctl/cmd"
)

var (
	Version   = "dev"
	BuildTime = ""
)

func main() {
	cmd.Execute()
}
