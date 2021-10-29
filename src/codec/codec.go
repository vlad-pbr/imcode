package codec

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"math"
)

func Encode(dataStream io.Reader, cypherStream io.Reader, outStream io.Writer) error {

	// read cypher as image
	_img, _, err := image.Decode(cypherStream)
	if err != nil {
		return fmt.Errorf("could not decode cypher image: %s", err.Error())
	}

	// typecast to drawable image
	img, ok := _img.(draw.Image)
	if !ok {
		return fmt.Errorf("could not decode cypher as drawable image")
	}

	// amount of pixels needed to store metadata
	metaPixels := getMeta(img)
	maxBytes := ((img.Bounds().Max.X * img.Bounds().Max.Y) - metaPixels) * 4

	// encode loop
	buf := make([]byte, 4)
	dataLength := 0
	eof := false
	x, y := 0, 0
	for ; y < img.Bounds().Max.Y && !eof; y++ {
		for x = 0; x < img.Bounds().Max.X && !eof; x++ {

			// store current values
			r, g, b, a := img.At(x, y).RGBA()
			values := []uint8{byte(r / 257), byte(g / 257), byte(b / 257), byte(a / 257)}

			// read from data
			bytesRead, err := dataStream.Read(buf)
			if err != nil {

				// check for end of file
				if err == io.EOF {
					eof = true
				} else {
					return fmt.Errorf("could not read data byte from stream: %s", err.Error())
				}

			}

			// encode data into byte
			dataLength += bytesRead
			for i := 0; i < bytesRead; i++ {
				values[i] += buf[i]
			}

			// set pixel values in image
			img.Set(x, y, color.RGBA{
				R: values[0],
				G: values[1],
				B: values[2],
				A: values[3],
			})
		}
	}

	// ensure there's enough pixels
	if !eof || dataLength > maxBytes {
		fmt.Println(dataLength)
		return fmt.Errorf("only %d bytes can be stored using provided cypher", maxBytes)
	}

	// encode metadata
	axes := []int{x, y}
	for metaX, metaY := img.Bounds().Max.X-metaPixels, img.Bounds().Max.Y-1; metaX < img.Bounds().Max.X; metaX++ {

		// calculate colors
		values := [2][2]uint8{}
		for j := 0; j < 2; j++ {
			for i, value := 0, 255; i < 2; i, value = i+1, 255 {
				if axes[j] < value {
					value = axes[j]
				}
				axes[j] -= value
				values[j][i] = uint8(value)
			}
		}

		// set meta pixel color
		img.Set(metaX, metaY, color.RGBA{
			R: values[0][0],
			G: values[0][1],
			B: values[1][0],
			A: values[1][1],
		})
	}

	// write to given out stream
	if err = png.Encode(outStream, img); err != nil {
		return fmt.Errorf("could not encode image: %s", err.Error())
	}

	return nil
}

func getMeta(img image.Image) int {

	// get bounds
	boundX, boundY := img.Bounds().Max.X-1, img.Bounds().Max.Y-1

	// max of two bounds
	var maxBound int
	if boundX > boundY {
		maxBound = boundX
	} else {
		maxBound = boundY
	}

	// amount of pixels needed for metadata
	return int(math.Ceil(float64(maxBound) / 510))
}
