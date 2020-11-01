package main

// Config - Application Config
type Config struct {
	useCompression          bool
	permittedFileExtensions []string
}

func getConfig() Config {

	var config = Config{
		useCompression: true,
		permittedFileExtensions: []string{
			"png",
			"jpg",
			"jpeg",
			"gif",
		},
	}

	return config
}
