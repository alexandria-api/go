package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
)

func createApplicationFolders() {
	var applicationFolders = []string{
		config.dir.storage,
		config.dir.images,
		config.dir.temporary,
		config.dir.queue,
		config.dir.errors,
		config.dir.backlog,
	}

	for _, dir := range applicationFolders {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err = os.Mkdir(dir, config.dir.permissions)
			if err != nil {
				log.Fatal("Failed to create dir '" + dir + "'.")
			}
		}
	}
}

func generateIdentifier() string {
	guid := xid.New()
	return guid.String()
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

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func moveImage(imagePath string, newPath string) bool {
	err := os.Rename(imagePath, newPath)
	if nil != err {
		return false
	}
	return true
}

func (id *imageIdentifier) compressAndFinish(imagePath string, finalPath string) bool {
	compress := exec.Command("imagecomp", imagePath)
	_, err := compress.Output()
	if nil != err {
		return false
	}

	imageMoved := moveImage(imagePath, finalPath)
	if !imageMoved {
		return false
	}
	id.updateState("finished")
	positionInQueue--

	return true
}

func respond(c *gin.Context, code int, res gin.H) {
	c.JSON(code, res)
}

func processQueue() {
	plan, _ := ioutil.ReadFile(config.stateFile)
	data := make(map[string]map[string]string)
	err := json.Unmarshal(plan, &data)
	if err != nil {
		log.Fatal(err)
	}

	for index, imageState := range data {
		if imageState["state"] == "queue" {

			root := config.dir.queue
			var files []string
			err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
				files = append(files, path)
				return nil
			})
			if err != nil {
				panic(err)
			}
			for _, file := range files {
				if strings.Contains(file, index) {
					split := strings.Split(file, ".")
					extension := split[len(split)-1]

					var (
						imageID     = &imageIdentifier{id: index}
						currentPath = config.dir.queue + "/" + index + "." + extension
						finalPath   = config.dir.images + "/" + index + "." + extension
					)

					log.Printf("Compressing image from queue: %s", imageID.id)
					go imageID.compressAndFinish(currentPath, finalPath)
				}
			}

		}
	}
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func (imd *imageIdentifier) updateState(state string) {
	if _, err := os.Stat(config.stateFile); os.IsNotExist(err) {
		stateFile, err := os.Create(config.stateFile)
		if err != nil {
			log.Fatal(err)
		}
		emptyStateMap := make(map[string]string)
		file, _ := json.Marshal(emptyStateMap)
		stateFile.Write(file)
	}

	plan, _ := ioutil.ReadFile(config.stateFile)
	data := make(map[string]map[string]string)
	err := json.Unmarshal(plan, &data)
	if err != nil {
		log.Fatal(err)
	}

	stateMap := make(map[string]string)
	stateMap["state"] = state

	data[imd.id] = stateMap

	file, _ := json.MarshalIndent(data, "", " ")

	_ = ioutil.WriteFile(config.stateFile, file, 0644)
}

type imageIdentifier struct {
	id string
}
