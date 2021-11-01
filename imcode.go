package main

import (
	"fmt"
	"os"

	"github.com/vlad-pbr/imcode/src/cli"
)

func main() {

	if err := cli.Handle(); err != nil {
		fmt.Fprintf(os.Stderr, "imcode: %s\n", err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
