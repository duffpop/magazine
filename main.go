package main

import (
	"github.com/gin-gonic/gin"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

var (
	host          = "img.hayden.lol"
	expectedToken = os.Getenv("EXPECTED_TOKEN")
)

func main() {
	router := gin.Default()

	// Set a lower memory limit for multipart forms (default is 32 MiB)
	router.MaxMultipartMemory = 8 << 20 // 8 MiB

	if expectedToken == "" {
		panic("EXPECTED_TOKEN environment variable is not set") // Fail fast if the token is not set
	}

	// Apply the token validation middleware globally
	router.Use(tokenValidationMiddleware)

	router.POST("/upload", uploadHandler)

	err := router.Run(":8080")
	if err != nil {
		return
	}
}

func uploadHandler(c *gin.Context) {
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

		// Save webp
		filename = filename[:len(filename)-len(ext)] + ".webp"
		newFilename := wordGen() + ".webp"
		if err := saveWebP(newFilename, img); err != nil {
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
}
