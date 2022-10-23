## alexandria
[<img class="badge" tag="github.com/alexandria-api/go" src="https://goreportcard.com/badge/github.com/alexandria-api/go">](https://goreportcard.com/badge/github.com/alexandria-api/go)

#### version 1.0.2
GO Image Storage and Retrieval API using [Gin](https://github.com/gin-gonic/gin)

Made with images in mind, but I guess aside from compression it can handle most if not all file types.

### Software Requirements
- [Go](https://go.dev/)
- [imagecomp](https://github.com/aprimadi/imagecomp)

### Go Dependancies
- `go get github.com/gin-gonic/gin`
- `go get github.com/rs/xid`

### Run
- `go run ./`
- `go build -o alexandria && ./alexandria`

### Todo:
- ~~Upload~~
- ~~Retrieval~~
- ~~Compression~~
- ~~Verify deployment integrity (Creates necessary project folders)~~
- ~~Check if file is png or jpg for compression~~
- ~~Add uploads to queue, so they can be compressed while the http request is finished~~
- ~~Implement a proper job queue~~
- ~~Allow a limit to be put on the amount of simultaneous image compressions.~~
- Tests
- Function that processes job queue on startup
- Add logging to all errors
- Upload multiple files at once
- Check if image already exists in storage
  - By filename
  - Using ImageMagick
- Allow for optional compression
- Rate limit
- Add support for other file types
- Process temporary folder into queue on start up
- Add metadata signalling when image has been compressed
- Process queue folder by checking metadata on start up
- Add clear state command