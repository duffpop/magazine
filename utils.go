package main

import (
	"bytes"
	"errors"
	"github.com/gin-gonic/gin"
	"image"
	"image/jpeg"
	"image/png"
	"math/rand"
	"net/http"

	"github.com/chai2010/webp"
	"tailscale.com/words"
)

func convertToWebP(imageBytes []byte) (image.Image, error) {
	var buf bytes.Buffer
	contentType := http.DetectContentType(imageBytes)
	imgReader := bytes.NewReader(imageBytes)

	switch contentType {
	case "image/png":
		img, err := png.Decode(imgReader)
		if err != nil {
			return nil, errors.Join(err, errors.New("unable to decode png"))
		}

		if err = webp.Encode(&buf, img, &webp.Options{Lossless: true}); err != nil {
			return nil, errors.Join(err, errors.New("unable to encode webp"))
		}

		return img, nil
	case "image/jpeg":
		img, err := jpeg.Decode(imgReader)
		if err != nil {
			return nil, errors.Join(err, errors.New("unable to decode jpeg"))
		}

		if err = webp.Encode(&buf, img, &webp.Options{Lossless: true}); err != nil {
			return nil, errors.Join(err, errors.New("unable to encode webp"))
		}

		return img, nil
	}
	return nil, errors.New("unsupported image format")
}

func saveWebP(filename string, img image.Image) error {
	return webp.Save(filename, img, &webp.Options{Lossless: true})
}

func wordGen() string {
	scales := words.Scales()
	tails := words.Tails()

	var word string

	word += scales[rand.Intn(len(scales))]
	word += "-"
	word += tails[rand.Intn(len(tails))]

	return word
}

// tokenValidationMiddleware is a middleware function for validating the token present in the query parameters
func tokenValidationMiddleware(c *gin.Context) {
	token := c.Query("token")

	// Check if the token is present and valid
	if token == "" || token != expectedToken {
		c.String(http.StatusUnauthorized, "ERR: invalid or missing token")
		c.Abort()
		return
	}

	// If the token is valid, proceed to the next handler
	c.Next()
}
