package main

import (
	"os"

	"github.com/vlad-pbr/imcode/src/codec"
)

func main() {

	// data to encode
	data, err := os.Open("../imcode-samples/chrome.png")
	if err != nil {
		panic(err)
	}
	defer data.Close()

	// cypher image
	cypher, err := os.Open("../imcode-samples/weird.png")
	if err != nil {
		panic(err)
	}
	defer cypher.Close()

	// destination file
	out, err := os.Create("../imcode-samples/weird.out.png")
	if err != nil {
		panic(err)
	}

	// encode
	if err := codec.Encode(data, cypher, out); err != nil {
		panic(err)
	}
}
