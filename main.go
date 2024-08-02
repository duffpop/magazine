package main

import (
	"net/http"
	"net/url"
	"path/filepath"

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

		urlGen := url.URL{
			Scheme: "https",
			Host:   host,
			Path:   "/" + filename,
		}

		c.String(http.StatusOK, "SUCCESS: %s/%s", urlGen, file.Filename)
	})

	router.Run(":8080")
}
