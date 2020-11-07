package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// Upload action
// Upload a file into a compression queue or storage.
func Upload(c *gin.Context) {

	file, err := c.FormFile("file")
	if nil != err {
		respond(c, 400, gin.H{
			"error":   "bad request",
			"message": "A file must be supplied as value in form data with key 'file'",
		})
		return
	}

	// file.Filename must be a file name and an extension
	matched, err := regexp.MatchString(`[^\\]*\.(\w+)$`, file.Filename)
	if nil != err {
		respond(c, 400, gin.H{
			"error":   "bad request",
			"message": "File name must be a file name and an extension",
		})
		return
	}

	if !matched {
		respond(c, 400, gin.H{
			"error":   "bad request",
			"message": "Content Disposition Filename must be supplied",
		})
		return
	}

	var (
		explodedFilename []string = strings.Split(file.Filename, ".")
		extension                 = explodedFilename[1]
	)

	if !stringInSlice(strings.ToLower(extension), config.permittedFileExtensions) {
		var pretty = strings.Join(config.permittedFileExtensions, ", ")
		respond(c, 400, gin.H{
			"error":   "bad request",
			"message": "File extension must be one of: " + pretty,
		})
		return
	}

	var (
		filename       = generateIdentifier()
		imageID        = &imageIdentifier{id: filename}
		fullname       = "/" + filename + "." + extension
		temporaryPath  = config.dir.temporary + fullname
		queuePath      = config.dir.queue + fullname
		successfulPath = config.dir.images + fullname
	)

	err = c.SaveUploadedFile(file, temporaryPath)
	if nil != err {
		respond(c, 500, gin.H{
			"error":   "server failure",
			"message": "Failed to save image to temporary location.",
		})
		return
	}
	imageID.updateState("temporary")

	movedToQueue := moveImage(temporaryPath, queuePath)
	if !movedToQueue {
		respond(c, 500, gin.H{
			"error":   "server failure",
			"message": "Failed to move image into queue.",
		})
		return
	}
	imageID.updateState("queue")

	var compressableExtensions = []string{"png", "jpg", "jpeg"}

	if contains(compressableExtensions, strings.ToLower(extension)) {

		// Move to backlog if queue is full

		go imageID.compressAndFinish(queuePath, successfulPath, false)
		positionInQueue++
		respond(c, http.StatusOK, gin.H{
			"success": "file added to queue",
			"message": "Position in queue: " + fmt.Sprintf("%v", positionInQueue),
			"id":      filename,
		})
		log.Println("File added to compression queue.")
		return
	}

	imageMoved := moveImage(queuePath, successfulPath)
	if !imageMoved {
		respond(c, 500, gin.H{
			"error":   "server failure",
			"message": "Failed to move image into final location.",
		})
		return
	}
	imageID.updateState("finished")

	respond(c, http.StatusOK, gin.H{
		"success": "file uploaded",
		"message": "File has been uploaded successfully.",
		"id":      filename,
	})
	log.Println("File skipped compression and uploaded.")
}

// Retrieve action
// Retrieve an uploaded file
func Retrieve(c *gin.Context) {
	id := c.Param("id")
	if !validIdentifier(id) {
		respond(c, 400, gin.H{
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

	var foundFile = "not found"

	for _, file := range files {
		if strings.Contains(file, id) {
			foundFile = file
		}
	}

	if "not found" == foundFile {
		respond(c, http.StatusOK, gin.H{
			"error":   "file not found",
			"message": "No file matching the supplied identifier was found.",
		})
		return
	}

	c.File(foundFile)
}
