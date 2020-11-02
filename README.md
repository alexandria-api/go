## alexandria
#### version 1.0.0
GO Image Storage and Retrieval API using [Gin](https://github.com/gin-gonic/gin)

### Requirements
[GO](https://www.php.net/)

### Setup
1. Create directory `storage/images`
2. Edit config.go with your preferred allowed file extensions.
3. Run `go build`
4. Run `./alexandria-go`

### Todo:
- ~~Upload~~
- ~~Retrieval~~
- ~~Compression~~
- ~~Verify deployment integrity (Creates necessary project folders)~~
- Add uploads to queue so they can be compressed later
- Versioning
- Check if image already exists in storage
- Tests
- Allow for optional compression
- Rate limit
