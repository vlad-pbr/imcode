package main

import (
	"fmt"
	"os"

	"github.com/vlad-pbr/imcode/src/cli"
)

var VERSION = "undefined"

func main() {

	if err := cli.Handle(VERSION); err != nil {
		fmt.Fprintf(os.Stderr, "imcode: %s\n", err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
