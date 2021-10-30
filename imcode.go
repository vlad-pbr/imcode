package main

import (
	"os"

	"github.com/vlad-pbr/imcode/src/codec"
)

func main() {

	// data to encode
	data, err := os.Open("../imcode-samples/lorem.txt")
	if err != nil {
		panic(err)
	}
	defer data.Close()

	// cypher image
	cypher, err := os.Open("../imcode-samples/32.png")
	if err != nil {
		panic(err)
	}
	defer cypher.Close()

	// destination file
	out, err := os.Create("../imcode-samples/32.out.png")
	if err != nil {
		panic(err)
	}
	defer out.Close()

	// encode
	if err := codec.Encode(data, cypher, out); err != nil {
		panic(err)
	}

	// =================

	// data to encode
	coded, err := os.Open("../imcode-samples/32.out.png")
	if err != nil {
		panic(err)
	}
	defer coded.Close()

	// cypher image
	anotherCypher, err := os.Open("../imcode-samples/32.png")
	if err != nil {
		panic(err)
	}
	defer anotherCypher.Close()

	// destination file
	decoded, err := os.Create("../imcode-samples/32.decoded.png")
	if err != nil {
		panic(err)
	}
	defer decoded.Close()

	// decode
	if err = codec.Decode(coded, anotherCypher, decoded); err != nil {
		panic(err)
	}
}
