package main

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/chai2010/webp"
	"github.com/gin-gonic/gin"
	"tailscale.com/words"
)

var (
	host = "img.hayden.lol"
	//expectedToken = os.Getenv("EXPECTED_TOKEN")
	expectedToken = "hello"
)

func main() {
	router := gin.Default()

	// Set a lower memory limit for multipart forms (default is 32 MiB)
	router.MaxMultipartMemory = 8 << 20 // 8 MiB

	if expectedToken == "" {
		panic("EXPECTED_TOKEN environment variable is not set") // Fail fast if the token is not set
	}

	// Apply the token validation middleware globally or to the specific route
	router.Use(tokenValidationMiddleware)

	router.POST("/upload", func(c *gin.Context) {
		// Source
		file, err := c.FormFile("file")
		if err != nil {
			c.String(http.StatusBadRequest, "ERR: get form err: %s", err.Error())
			return
		}

		filename := filepath.Base(file.Filename)
		ext := filepath.Ext(filename)

		// Open the file
		imageBytes, err := file.Open()
		if err != nil {
			c.String(http.StatusBadRequest, "ERR: unable to open file: %s", err.Error())
			return
		}

		defer func(imageBytes multipart.File) {
			err := imageBytes.Close()
			if err != nil {
				c.String(http.StatusBadRequest, "ERR: unable to close file: %s", err.Error())
			}
		}(imageBytes)

		// Read the file content
		fileContent, err := io.ReadAll(imageBytes)
		if err != nil {
			c.String(http.StatusBadRequest, "ERR: unable to read file: %s", err.Error())
			return
		}

		if ext != ".webp" {
			// Convert to webp
			img, err := convertToWebP(fileContent)
			if err != nil {
				c.String(http.StatusBadRequest, "ERR: unable to convert to webp: %s", err.Error())
				return
			}

			// Save the webp file
			filename = filename[:len(filename)-len(ext)] + ".webp"
			newFilename := wordGen() + ".webp"
			if err := webp.Save(newFilename, img, &webp.Options{Lossless: true}); err != nil {
				c.String(http.StatusBadRequest, "ERR: unable to save webp: %s", err.Error())
				return
			}
		} else {
			// Save the original webp file
			if err := c.SaveUploadedFile(file, filename); err != nil {
				c.String(http.StatusBadRequest, "ERR: upload file err: %s", err.Error())
				return
			}
		}
		urlGen := url.URL{
			Scheme: "https", Host: host,
			Path: "/" + filename,
		}
		c.String(http.StatusOK, "SUCCESS: %s", urlGen.String())
	})

	err := router.Run(":8080")
	if err != nil {
		return
	}
}

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

func wordGen() string {
	scales := words.Scales()
	tails := words.Tails()

	var word string

	word += scales[rand.Intn(len(scales))]
	word += "-"
	word += tails[rand.Intn(len(tails))]

	return word
}

// tokenValidationMiddleware is a middleware function for validating the token present in the query parameters.
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
