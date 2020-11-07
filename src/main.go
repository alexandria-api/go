package main

import (
	"github.com/gin-gonic/gin"
)

var (
	config          Config
	positionInQueue = 0
)

func main() {

	config = getConfig()
	createApplicationFolders()
	processQueue()

	// gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	router.MaxMultipartMemory = 8 << 20

	router.POST("/upload", Upload)
	router.GET("/retrieve/:id", Retrieve)

	router.Run(":8080")

}
