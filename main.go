package main

import (
	"bytes"
	"errors"
	"image"
	"image/png"
	"io"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/chai2010/webp"
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

		filename := filepath.Base(file.Filename)
		if err := c.SaveUploadedFile(file, filename); err != nil {
			c.String(http.StatusBadRequest, "upload file err: %s", err.Error())
			return
		}

		// check syntax of the filename, it should be a webp file
		// if not, convert to webp
		if filepath.Ext(filename) != ".webp" {
			// Convert to webp
			imageBytes, err := file.Open()
			if err != nil {
				c.String(http.StatusBadRequest, "unable to open file: %s", err.Error())
				return
			}
			defer imageBytes.Close()

			// Read the file content
			fileContent, err := io.ReadAll(imageBytes)
			if err != nil {
				c.String(http.StatusBadRequest, "unable to read file: %s", err.Error())
				return
			}

			img, err := convertToWebP(fileContent)
			if err != nil {
				c.String(http.StatusBadRequest, "unable to convert to webp: %s", err.Error())
				return
			}

			// Save the webp file
			if err := webp.Save(filename, img, &webp.Options{Lossless: true}); err != nil {
				c.String(http.StatusBadRequest, "unable to save webp: %s", err.Error())
				return
			}
		}

		urlGen := url.URL{
			Scheme: "https",
			Host:   host,
			Path:   "/" + filename,
		}

		c.String(http.StatusOK, "SUCCESS: %s/%s", urlGen, file.Filename)
	})

	router.Run(":8080")
}

func convertToWebP(imageBytes []byte) (image.Image, error) {
	var buf bytes.Buffer
	contentType := http.DetectContentType(imageBytes)
	imgReader := bytes.NewReader(imageBytes)
	// strippedName := filepath.Base(name)
	// convertedName := strippedName[0 : len(strippedName)-len(filepath.Ext(strippedName))] // Remove file extension

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
	}
	return nil, nil
}
