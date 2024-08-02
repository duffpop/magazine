package main

import (
	"bytes"
	"errors"
	"fmt"
	"image/jpeg"
	"image/png"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

var host string = "img.hayden.lol"

func main() {
	router := gin.Default()

	// Set a lower memory limit for multipart forms (default is 32 MiB)
	router.MaxMultipartMemory = 8 << 20 // 8 MiB

	// router.Static("/", "./public")
	router.POST("/upload", func(c *gin.Context) {
		// Source
		file, err := c.FormFile("file")
		if err != nil {
			c.String(http.StatusBadRequest, "get form err: %s", err.Error())
			return
		}

		// Convert to png
		fileBytes, err := file.Open()
		if err != nil {
			c.String(http.StatusBadRequest, "open file err: %s", err.Error())
			return
		}
		defer fileBytes.Close()

		imgBytes := new(bytes.Buffer)
		imgBytes.ReadFrom(fileBytes)
		imgBytes, err = imgToPng(imgBytes.Bytes(), err)

		urlGen := url.URL{
			Scheme: "https",
			Host:   host,
			Path:   "/" + filename,
		}

		c.String(http.StatusOK, "SUCCESS: %s/%s", urlGen)
	})

	router.Run(":8080")
}

// imgToPng converts an image to png
func imgToPng(imageBytes []byte, err error) ([]byte, error) {
	contentType := http.DetectContentType(imageBytes)

	switch contentType {
	case "image/png":
	case "image/jpeg":
		img, err := jpeg.Decode(bytes.NewReader(imageBytes))
		if err != nil {
			return nil, errors.Join(err, errors.New("unable to decode jpeg"))
		}

		buf := new(bytes.Buffer)
		if err := png.Encode(buf, img); err != nil {
			return nil, errors.Join(err, errors.New("unable to encode png"))
		}

		return buf.Bytes(), nil
	}

	return nil, fmt.Errorf("unable to convert %#v to png", contentType)
}

// gets image from multipart form upload
func getImageFromForm(c *gin.Context) ([]byte, error) {
	file, err := c.FormFile("file")
	if err != nil {
		return nil, err
	}

	fileBytes, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer fileBytes.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(fileBytes)

	return buf.Bytes(), nil
}
