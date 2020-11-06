package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"

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

func compressAndFinishUploadedImage(id string, imagePath string, finalPath string) bool {
	saveState(id, "temporary")

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

func respond(c *gin.Context, code int, res gin.H) {
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

func saveState(id string, state string) {
	if _, err := os.Stat(config.stateFile); os.IsNotExist(err) {
		_, err := os.Create(config.stateFile)
		if err != nil {
			log.Fatal(err)
		}
	}

	plan, _ := ioutil.ReadFile(config.stateFile)
	data := make(map[string]map[string]string)
	err := json.Unmarshal(plan, &data)
	if err != nil {
		log.Fatal(err)
	}

	stateMap := make(map[string]string)
	stateMap["state"] = state

	data[id] = stateMap

	file, _ := json.MarshalIndent(data, "", " ")

	_ = ioutil.WriteFile(config.stateFile, file, 0644)
}

func updateState() {
	// TODO: Implement updateState function
}
