package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/vlad-pbr/imcode/src/codec"
)

var F_VERSION bool
var F_DECODE bool
var F_FROM string
var F_CYPHER string
var F_TO string

func Handle(version string) error {

	// define flags
	flag.BoolVar(&F_VERSION, "version", false, "display tool version")
	flag.BoolVar(&F_DECODE, "decode", false, "should the input file be decoded (encoding is default)")
	flag.StringVar(&F_FROM, "from", "", "path to input file ('-' for stdin) [required]")
	flag.StringVar(&F_CYPHER, "cypher", "", "path to cypher image ('-' for stdin) [required]")
	flag.StringVar(&F_TO, "to", "-", "path to output file ('-' for stdout) [required]")

	// parse args
	flag.Parse()

	// display version then quit
	if F_VERSION {
		fmt.Printf("imcode version: %s\n", version)
		os.Exit(0)
	}

	// check for undefined flags
	if len(F_FROM) == 0 || len(F_CYPHER) == 0 && len(F_TO) == 0 {
		return fmt.Errorf("one of the required flags is not defined")
	}

	// check for duplicate stdins
	if F_FROM == "-" && F_CYPHER == "-" {
		return fmt.Errorf("only one stdin flag can be specified")
	}

	// parse input
	var in *os.File
	var err error
	if F_FROM == "-" {
		in = os.Stdin
	} else {
		in, err = os.Open(F_FROM)
		if err != nil {
			return fmt.Errorf("error reading file %s: %s", F_FROM, err.Error())
		}
	}

	// parse cypher
	var cypher *os.File
	if F_CYPHER == "-" {
		cypher = os.Stdin
	} else {
		cypher, err = os.Open(F_CYPHER)
		if err != nil {
			return fmt.Errorf("error reading file %s: %s", F_CYPHER, err.Error())
		}
	}

	// parse output
	var out *os.File
	if F_TO == "-" {
		out = os.Stdout
	} else {
		out, err = os.Create(F_TO)
		if err != nil {
			return fmt.Errorf("error opening file %s: %s", F_CYPHER, err.Error())
		}
	}

	// call codec
	if F_DECODE {
		err = codec.Decode(in, cypher, out)
	} else {
		err = codec.Encode(in, cypher, out)
	}
	if err != nil {
		return fmt.Errorf("codec error: %s", err.Error())
	}

	return nil
}
