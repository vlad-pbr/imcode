# Imcode

Encode your data into a PNG.

## In short

The tool receives data and a cypher PNG image and then encodes that data into the given PNG:

<p align="center">
  <img width="400" height="400" src="doc/cypher.png">
  <img width="400" height="400" src="doc/result.png">
</p>

It can then decode the data back using the encoded image and the original cypher.

> Due to lossless compression of the PNG format, data is also compressed when encoded.

## Usage

Imcode can be used as a binary package:

``` bash
# encode data with cypher and save from stdout
./imcode --from path/to/data.txt --cypher path/to/cypher.png > path/to/encoded.png

# read encoded result from stdin and restore the original data
cat path/to/encoded.png | ./imcode --from=- --cypher path/to/cypher.png --to path/to/result.txt --decode
```

It can also be imported and used as a Go package:

``` golang
package main

import (
	"os"

	"github.com/vlad-pbr/imcode/src/codec"
)

func main() {

	// open data to encode
	data, err := os.Open("path/to/data.txt")
	if err != nil {
		panic(err)
	}
	defer data.Close()

	// open cypher png
	cypher, err := os.Open("path/to/cypher.png")
	if err != nil {
		panic(err)
	}
	defer cypher.Close()

	// create destination file
	out, err := os.Create("path/to/encoded.png")
	if err != nil {
		panic(err)
	}
	defer out.Close()

	// encode
	if err := codec.Encode(data, cypher, out); err != nil {
		panic(err)
	}
}
```

## Compile from source

Clone this repository, `cd` to cloned directory, `git checkout` to a release tag and build:

``` bash
go build -ldflags="-X 'main.VERSION=$(git rev-parse --abbrev-ref HEAD)'"
```