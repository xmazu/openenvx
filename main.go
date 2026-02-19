package main

import (
	"github.com/xmazu/openenvx/cmd"
)

var (
	Version   string
	BuildTime string
)

func main() {
	cmd.SetVersion(Version)
	cmd.Execute()
}
