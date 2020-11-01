package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
)

var (
	config Config
)

func main() {

	config = getConfig()

	// gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	os.Setenv("PORT", "8080")

	// Set a lower memory limit for multipart forms (default is 32 MiB)
	router.MaxMultipartMemory = 8 << 20 // 8 MiB

	router.POST("/upload", upload)
	router.GET("/retrieve/:id", retrieve)

	router.Run()

}

func generateIdentifier() string {
	guid := xid.New()
	return guid.String()
}

func retrieve(c *gin.Context) {
	id := c.Param("id")
	if !validIdentifier(id) {
		sendResponse(c, 400, gin.H{
			"error":   "bad request",
			"message": "Invalid image identifier",
		})
		return
	}

	var (
		root  = "storage/images"
		files []string
	)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})
	if err != nil {
		panic(err)
	}

	var foundFile string

	for _, file := range files {
		if strings.Contains(file, id) {
			foundFile = file
		}
	}

	c.File(foundFile)
}

func validIdentifier(identifier string) bool {
	if len(identifier) != 20 {
		return false
	}

	matched, err := regexp.MatchString(`([A-Za-z0-9\-]+)`, identifier)
	if nil != err {
		log.Fatal(err)
	}

	if !matched {
		return false
	}
	return true
}

func upload(c *gin.Context) {

	file, err := c.FormFile("file")
	if nil != err {
		sendResponse(c, 400, gin.H{
			"error":   "bad request",
			"message": "A file must be supplied as value in form data with key 'file'",
		})
		return
	}

	// file.Filename must be a file name and an extension
	matched, err := regexp.MatchString(`[^\\]*\.(\w+)$`, file.Filename)
	if nil != err {
		sendResponse(c, 400, gin.H{
			"error":   "bad request",
			"message": "File name must be a file name and an extension",
		})
		return
	}

	if !matched {
		sendResponse(c, 400, gin.H{
			"error":   "bad request",
			"message": "Content Disposition Filename must be supplied",
		})
		return
	}

	var (
		explodedFilename []string = strings.Split(file.Filename, ".")
		extension                 = explodedFilename[1]
	)

	if !stringInSlice(extension, config.permittedFileExtensions) {
		var pretty = strings.Join(config.permittedFileExtensions, ", ")
		sendResponse(c, 400, gin.H{
			"error":   "bad request",
			"message": "File extension must be one of: " + pretty,
		})
		return
	}

	var (
		filename = generateIdentifier()
		fullname = filename + "." + extension
	)

	c.SaveUploadedFile(file, "storage/images/"+fullname)
	sendResponse(c, http.StatusOK, gin.H{
		"success": "file uploaded",
		"message": "file has been saved as " + fullname,
	})
}

func sendResponse(c *gin.Context, code int, res gin.H) {
	c.JSON(code, res)
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
