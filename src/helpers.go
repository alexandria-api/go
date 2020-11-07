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
	log.Printf("Compressing image: %s", id.id)

	id.updateState("compressing")
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

	log.Printf("current compressions: %d", getAmountOfCurrentCompressions())

	// Refresh queue
	if 0 == getAmountOfCurrentCompressions() {
		processQueue()
	}
	return true
}

func respond(c *gin.Context, code int, res gin.H) {
	c.JSON(code, res)
}

func processQueue() {
	stateFile, _ := ioutil.ReadFile(config.stateFile)
	data := make(map[string]map[string]string)
	err := json.Unmarshal(stateFile, &data)
	if err != nil {
		log.Fatal(err)
	}

	var files []string
	err = filepath.Walk(config.dir.queue, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})
	if err != nil {
		panic(err)
	}

	var imagesToCompress []*imageSnapshot

	for index, imageState := range data {

		if "queue" == imageState["state"] {

			for _, file := range files {

				if strings.Contains(file, index) {
					split := strings.Split(file, ".")
					extension := split[len(split)-1]

					var (
						imageID         = &imageIdentifier{id: index}
						currentPath     = config.dir.queue + "/" + index + "." + extension
						finalPath       = config.dir.images + "/" + index + "." + extension
						imageToCompress = &imageSnapshot{
							imd:         imageID,
							currentPath: currentPath,
							finalPath:   finalPath,
						}
					)
					imagesToCompress = append(imagesToCompress, imageToCompress)
				}
			}

		}
	}

	if 0 == len(imagesToCompress) {
		return
	}

	stateFile, _ = ioutil.ReadFile(config.stateFile)
	data = make(map[string]map[string]string)
	err = json.Unmarshal(stateFile, &data)
	if err != nil {
		log.Fatal(err)
	}

	for _, imageState := range data {
		if "compressing" == imageState["state"] {
			currentlyCompressing++
		}
	}

	var canCompress = config.simultaneousCompressions - currentlyCompressing

	if config.simultaneousCompressions < currentlyCompressing {
		log.Fatal("Too many compressions happening.")
	}

	log.Printf("imagesToCompress: %d", len(imagesToCompress))

	for _, image := range imagesToCompress {
		if canCompress > 0 {
			go image.imd.compressAndFinish(image.currentPath, image.finalPath)
			canCompress--
			positionInQueue++
		}
	}
}

type imageSnapshot struct {
	imd         *imageIdentifier
	currentPath string
	finalPath   string
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func createStateFile() {
	if _, err := os.Stat(config.stateFile); os.IsNotExist(err) {
		stateFile, err := os.Create(config.stateFile)
		if err != nil {
			log.Fatal(err)
		}
		emptyStateMap := make(map[string]string)
		file, _ := json.Marshal(emptyStateMap)
		stateFile.Write(file)
	}
}

func (id *imageIdentifier) updateState(state string) {
	stateFile, _ := ioutil.ReadFile(config.stateFile)
	data := make(map[string]map[string]string)
	err := json.Unmarshal(stateFile, &data)
	if err != nil {
		log.Fatal(err)
	}

	stateMap := make(map[string]string)
	stateMap["state"] = state

	data[id.id] = stateMap

	file, _ := json.MarshalIndent(data, "", " ")

	_ = ioutil.WriteFile(config.stateFile, file, 0644)
}

func resetCompressionStates() {
	stateFile, _ := ioutil.ReadFile(config.stateFile)
	data := make(map[string]map[string]string)
	err := json.Unmarshal(stateFile, &data)
	if err != nil {
		log.Fatal(err)
	}

	queueStateMap := make(map[string]string)
	queueStateMap["state"] = "queue"

	for index, imageState := range data {
		if "compressing" == imageState["state"] {
			data[index] = queueStateMap
		}
	}

	file, _ := json.MarshalIndent(data, "", " ")

	_ = ioutil.WriteFile(config.stateFile, file, 0644)
}

func getAmountOfCurrentCompressions() int {
	stateFile, _ := ioutil.ReadFile(config.stateFile)
	data := make(map[string]map[string]string)
	err := json.Unmarshal(stateFile, &data)
	if err != nil {
		log.Fatal(err)
	}

	var count = 0

	for _, imageState := range data {
		if "compressing" == imageState["state"] {
			count++
		}
	}

	return count
}

type imageIdentifier struct {
	id string
}
