package main

import "os"

// Config - Configuration
type Config struct {
	useCompression           bool
	simultaneousCompressions int
	permittedFileExtensions  []string
	dir                      ApplicationFolders
	stateFile                string
}

// ApplicationFolders - Folders used by the application.
type ApplicationFolders struct {
	permissions os.FileMode
	storage     string
	images      string
	temporary   string
	queue       string
	errors      string
	backlog     string
	logs        string
}

func getConfig() Config {

	/*
		Edit config here
	*/
	var config = Config{
		useCompression:           true, // Only png and jpg (jpeg) files can be compressed.
		simultaneousCompressions: 3,    // Any more than 6 may crash the system running. Ignored if useCompression is false.
		permittedFileExtensions: []string{
			"png",
			"jpg",
			"jpeg",
			"gif",
			"webp",
			"bmp",
		},
		dir: ApplicationFolders{ // These folders will be created if they do not exist.
			permissions: 0777, // Permissions the folders will be created with.
			storage:     "storage",
			images:      "storage/images",
			// temporary:   "storage/temporary",
			queue:   "storage/queue",
			errors:  "storage/errors",
			backlog: "storage/backlog",
			logs:    "storage/logs",
		},
		stateFile: "storage/state.json", // You don't need to change the name of this.
	}

	return config
}
