package main

import "os"

// Config - Configuration
type Config struct {
	useCompression          bool
	permittedFileExtensions []string
	dir                     ApplicationFolders
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
		useCompression: true, // Only png and jpg (jpeg) files can be compressed.
		permittedFileExtensions: []string{
			"png",
			"jpg",
			"jpeg",
			"gif",
		},
		dir: ApplicationFolders{ // These folders will be created if they do not exist.
			permissions: 0777, // Permissions the folders will be created with.
			storage:     "storage",
			images:      "storage/images",
			temporary:   "storage/temporary",
			queue:       "storage/queue",
			errors:      "storage/errors",
			backlog:     "storage/backlog",
			logs:        "storage/logs",
		},
	}

	return config
}
