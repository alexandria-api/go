## alexandria
#### version 1.0.2
GO Image Storage and Retrieval API using [Gin](https://github.com/gin-gonic/gin)

Made with images in mind, but I guess aside from compression it can handle most if not all file types. I have not tested this.
### Requirements
[Go](https://www.php.net/), [imagecomp](https://github.com/aprimadi/imagecomp)

### Setup
1. Edit config.go with your preferred allowed file extensions (Only png and jpg supported for compression).
2. Run `./alexandria-go`

### Todo:
- ~~Upload~~
- ~~Retrieval~~
- ~~Compression~~
- ~~Verify deployment integrity (Creates necessary project folders)~~
- Check if file is png or jpg for compression
- Create backlog if queue is higher than 10
- Process backlog with cron
- Add logging to all errors
- ~~Add uploads to queue, so they can be compressed later~~
- Versioning
- Check if image already exists in storage
- Tests
- Allow for optional compression
- Rate limit
- Transform alexandria into a file hosting api with the same image features but support for other file types?