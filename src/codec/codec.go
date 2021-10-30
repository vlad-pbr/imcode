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
	bytesRead := 0
	eof := false
	x, y := 0, 0
	for ; y < img.Bounds().Max.Y && !eof; y++ {
		for x = 0; x < img.Bounds().Max.X && !eof; x++ {

			// store current values
			r, g, b, a := img.At(x, y).RGBA()
			values := []uint8{byte(r / 257), byte(g / 257), byte(b / 257), byte(a / 257)}

			// read from data
			bytesRead, err = dataStream.Read(buf)
			if bytesRead < 4 {
				eof = true
			}

			// check for unexpected error
			if err != nil {
				if err != io.EOF {
					return fmt.Errorf("could not read data byte from stream: %s", err.Error())
				}
			} else {

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
	}

	// ensure there's enough pixels
	if !eof || dataLength > maxBytes {
		return fmt.Errorf("only %d bytes can be stored using provided cypher", maxBytes)
	}

	// adjust x,y coordinates
	y -= 1
	if x == 0 {
		x, y = img.Bounds().Max.X-1, y-1
	} else {
		x -= 1
	}

	// encode metadata
	axes := []int{x, y}
	for metaX, metaY := img.Bounds().Max.X-metaPixels, img.Bounds().Max.Y-1; metaX < img.Bounds().Max.X-1; metaX++ {

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

		fmt.Println(img.At(metaX, metaY).RGBA())
		fmt.Println(values[0][0])
		fmt.Println(metaX, metaY)
		fmt.Println(img.Bounds())

		// set meta pixel color
		img.Set(metaX, metaY, color.RGBA{
			R: 22,
			G: 0,
			B: 29,
			A: 0,
		})

		fmt.Println(img.At(metaX, metaY).RGBA())
	}

	// fmt.Println(img.At(30, 30).RGBA())

	// set amount of channels used in the final pixel
	// fmt.Println(img.Bounds().Max.X-1, img.Bounds().Max.Y-1)
	img.Set(img.Bounds().Max.X-1, img.Bounds().Max.Y-1, color.RGBA{
		R: 0,
		G: 0,
		B: 0,
		A: uint8(bytesRead),
	})
	// fmt.Println(img.At(img.Bounds().Max.X-1, img.Bounds().Max.Y-1).RGBA())

	// write to given out stream
	if err = png.Encode(outStream, img); err != nil {
		return fmt.Errorf("could not encode image: %s", err.Error())
	}

	return nil
}

func Decode(codedStream io.Reader, cypherStream io.Reader, outStream io.Writer) error {

	// read coded stream as image
	codedImg, _, err := image.Decode(codedStream)
	if err != nil {
		return fmt.Errorf("could not decode coded image: %s", err.Error())
	}

	// read cypher as image
	cypherImg, _, err := image.Decode(cypherStream)
	if err != nil {
		return fmt.Errorf("could not decode cypher image: %s", err.Error())
	}

	// ensure dimensions
	if !codedImg.Bounds().Eq(cypherImg.Bounds()) {
		return fmt.Errorf("provided cypher image and coded image are not equal in dimensions")
	}

	metaPixels := getMeta(codedImg)
	dataCoordinates := []int{0, 0}
	for metaX, metaY := codedImg.Bounds().Max.X-metaPixels, codedImg.Bounds().Max.Y-1; metaX < codedImg.Bounds().Max.X-1; metaX++ {

		// decode pixel
		r, g, b, a := codedImg.At(metaX, metaY).RGBA()
		values := [][]uint8{{byte(r / 257), byte(g / 257)}, {byte(b / 257), byte(a / 257)}}

		// calculate data pixel coordinates
		for j := 0; j < 2; j++ {
			for i := 0; i < 2; i++ {
				dataCoordinates[j] += int(values[j][i])
			}
		}
	}

	// parse channels used on final pixel
	_, _, _, a := codedImg.At(codedImg.Bounds().Max.X-1, codedImg.Bounds().Max.Y-1).RGBA()
	channelsUsed := a / 257
	if channelsUsed > 3 {
		return fmt.Errorf("invalid metadata: channels used %d is bigger than 3", channelsUsed)
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
	return int(math.Ceil(float64(maxBound)/510)) + 1
}
