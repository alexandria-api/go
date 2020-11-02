package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
)

var (
	config          Config
	positionInQueue = 0
)

func main() {

	config = getConfig()
	createApplicationFolders()

	// gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	router.MaxMultipartMemory = 8 << 20

	router.POST("/upload", upload)
	router.GET("/retrieve/:id", retrieve)

	router.Run(":8080")

}

func createApplicationFolders() {
	const (
		filePermission = 0777
		dirStorage     = "storage"
		dirImages      = dirStorage + "/images"
		dirTemporary   = dirStorage + "/temporary"
		dirQueue       = dirStorage + "/queue"
		dirError       = dirStorage + "/error"
	)
	if _, err := os.Stat(dirStorage); os.IsNotExist(err) {
		createDirErr := os.Mkdir(dirStorage, filePermission)
		if createDirErr != nil {
			log.Fatal("Failed to create dir '" + dirStorage + "'.")
		}
	}
	if _, err := os.Stat(dirImages); os.IsNotExist(err) {
		createDirErr := os.MkdirAll(dirImages, filePermission)
		if createDirErr != nil {
			log.Fatal("Failed to create dir '" + dirImages + "'.")
		}
	}
	if _, err := os.Stat(dirTemporary); os.IsNotExist(err) {
		createDirErr := os.MkdirAll(dirTemporary, filePermission)
		if createDirErr != nil {
			log.Fatal("Failed to create dir '" + dirTemporary + "'.")
		}
	}
	if _, err := os.Stat(dirQueue); os.IsNotExist(err) {
		createDirErr := os.MkdirAll(dirQueue, filePermission)
		if createDirErr != nil {
			log.Fatal("Failed to create dir '" + dirQueue + "'.")
		}
	}
	if _, err := os.Stat(dirError); os.IsNotExist(err) {
		createDirErr := os.MkdirAll(dirError, filePermission)
		if createDirErr != nil {
			log.Fatal("Failed to create dir '" + dirError + "'.")
		}
	}
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
		filename       = generateIdentifier()
		fullname       = filename + "." + extension
		temporaryPath  = "storage/temporary/" + fullname
		queuePath      = "storage/queue/" + fullname
		successfulPath = "storage/images/" + fullname
	)

	err = c.SaveUploadedFile(file, temporaryPath)
	if nil != err {
		sendResponse(c, 500, gin.H{
			"error":   "server failure",
			"message": "Failed to save image to temporary location.",
		})
		return
	}

	movedToQueue := moveImage(temporaryPath, queuePath)
	if !movedToQueue {
		sendResponse(c, 500, gin.H{
			"error":   "server failure",
			"message": "Failed to move image into queue.",
		})
		return
	}

	go compressAndFinishUploadedImage(queuePath, successfulPath)

	positionInQueue++
	sendResponse(c, http.StatusOK, gin.H{
		"success": "file added to queue",
		"message": "Position in queue: " + fmt.Sprintf("%v", positionInQueue),
		"id":      filename,
	})
}

func moveImage(imagePath string, newPath string) bool {
	err := os.Rename(imagePath, newPath)
	if nil != err {
		return false
	}
	return true
}

func compressAndFinishUploadedImage(imagePath string, finalPath string) bool {
	compress := exec.Command("imagecomp", imagePath)
	_, err := compress.Output()
	if nil != err {
		return false
	}

	imageMoved := moveImage(imagePath, finalPath)
	if !imageMoved {
		return false
	}
	positionInQueue--

	return true
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
