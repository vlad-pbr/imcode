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

func Encode(dataStream io.Reader, cypherStream io.Reader, outStream io.Writer, no_padding bool) error {

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
	maxPixels := (img.Bounds().Max.X * img.Bounds().Max.Y) - metaPixels
	maxBytes := maxPixels * 3

	// encode loop
	buf := make([]byte, 3)
	dataLength, bytesRead := 0, 0
	x, y := 0, 0
	for eof := false; y < img.Bounds().Max.Y && !eof; y++ {
		for x = 0; x < img.Bounds().Max.X && !eof; x++ {

			// store current values
			r, g, b, _ := img.At(x, y).RGBA()
			values := [3]uint8{byte(r / 257), byte(g / 257), byte(b / 257)}

			// read from data
			bytesRead, err = dataStream.Read(buf)
			if bytesRead < 3 {
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
					A: 255,
				})
			}
		}
	}

	// ensure there's enough pixels to store data
	if dataLength > maxBytes {
		return fmt.Errorf("only %d bytes can be stored using provided cypher", maxBytes)
	}

	// pad remaining pixels if padding is enabled and there are pixels left
	// remainingPixels := maxPixels - (dataLength / 3)
	// if !no_padding && remainingPixels > 0 {

	// 	for ; y < img.Bounds().Max.Y && remainingPixels != 0; y++ {
	// 		for ; x < img.Bounds().Max.X && remainingPixels != 0; x, remainingPixels = x+1, remainingPixels-1 {
	// 			//fmt.Println(x, y)
	// 			//rx, ry := rand.Intn(x), rand.Intn(y)
	// 			//fmt.Println(rx, ry)
	// 			//img.Set(x, y, img.At(rx, ry))
	// 		}
	// 	}

	// }

	// adjust x,y coordinates
	y -= 1
	if x == 0 {
		x, y = img.Bounds().Max.X-1, y-1
	} else {
		x -= 1
	}

	// encode metadata
	axes := []int{x, y}
	for metaY := img.Bounds().Max.Y - 1; metaY > 0 && metaPixels > 1; metaY-- {
		for metaX := img.Bounds().Max.X - 2; metaX > 0 && metaPixels > 1; metaX, metaPixels = metaX-1, metaPixels-1 {

			// calculate colors
			values := [2]uint8{}
			for i, value := 0, 255; i < 2; i, value = i+1, 255 {
				if axes[i] < value {
					value = axes[i]
				}
				axes[i] -= value
				values[i] = uint8(value)
			}

			// set meta pixel color
			img.Set(metaX, metaY, color.RGBA{
				R: values[0],
				G: values[1],
				B: 0,
				A: 255,
			})

		}
	}

	// set amount of channels used in the final pixel
	img.Set(img.Bounds().Max.X-1, img.Bounds().Max.Y-1, color.RGBA{
		R: uint8(bytesRead),
		G: 0,
		B: 0,
		A: 255,
	})

	// write to given out stream
	if err = png.Encode(outStream, img); err != nil {
		return fmt.Errorf("could not encode image: %s", err.Error())
	}

	return nil
}

func Decode(codedStream io.Reader, cypherStream io.Reader, outStream io.Writer) error {

	// read coded stream as image
	codedImage, _, err := image.Decode(codedStream)
	if err != nil {
		return fmt.Errorf("could not decode coded image: %s", err.Error())
	}

	// read cypher as image
	cypherImage, _, err := image.Decode(cypherStream)
	if err != nil {
		return fmt.Errorf("could not decode cypher image: %s", err.Error())
	}

	// ensure dimensions
	if !codedImage.Bounds().Eq(cypherImage.Bounds()) {
		return fmt.Errorf("provided cypher image and coded image are not equal in dimensions")
	}

	// read metadata pixels
	metaPixels := getMeta(codedImage)
	dataCoordinates := []int{0, 0}
	for metaY := codedImage.Bounds().Max.Y - 1; metaY > 0 && metaPixels > 1; metaY-- {
		for metaX := codedImage.Bounds().Max.X - 2; metaX > 0 && metaPixels > 1; metaX, metaPixels = metaX-1, metaPixels-1 {

			// decode pixel
			r, g, _, _ := codedImage.At(metaX, metaY).RGBA()
			values := []uint8{byte(r / 257), byte(g / 257)}

			// calculate data pixel coordinates
			for i := 0; i < 2; i++ {
				dataCoordinates[i] += int(values[i])
			}

		}
	}

	// parse channels used on final pixel
	r, _, _, _ := codedImage.At(codedImage.Bounds().Max.X-1, codedImage.Bounds().Max.Y-1).RGBA()
	channelsUsed := r / 257
	if channelsUsed > 3 {
		return fmt.Errorf("invalid metadata: channels used %d is bigger than 3", channelsUsed)
	}
	fileSize := (((dataCoordinates[1] * codedImage.Bounds().Max.X) + dataCoordinates[0]) * 3) + int(channelsUsed)

	// decode loop
	buf := make([]byte, 3)
	for y := 0; y < codedImage.Bounds().Max.Y && fileSize > 0; y++ {
		for x := 0; x < codedImage.Bounds().Max.X && fileSize > 0; x, fileSize = x+1, fileSize-3 {

			// store image values
			values := [2][3]uint8{}
			for i, img := range [2]image.Image{codedImage, cypherImage} {
				r, g, b, _ := img.At(x, y).RGBA()
				values[i] = [3]uint8{byte(r / 257), byte(g / 257), byte(b / 257)}
			}

			// adjust buffer size if needed
			if fileSize < 3 {
				buf = make([]byte, fileSize)
			}

			// decode byte
			for i := 0; i < len(buf); i++ {
				buf[i] = values[0][i] - values[1][i]
			}

			// write to output stream
			_, err := outStream.Write(buf)
			if err != nil {
				return fmt.Errorf("error occurred while writing to output stream: %s", err.Error())
			}

		}
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
	return int(math.Ceil(float64(maxBound)/255)) + 1
}
