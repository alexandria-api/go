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
	return false
}

func (id *imageIdentifier) compressAndFinish(imagePath string, finalPath string, afterProcessQueue bool) bool {
	log.Printf("Compressing image: %s", id.id)

	id.updateState("compressing")
	log.Printf("current compressions: %d", getAmountOfCurrentCompressions())
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

	// Refresh queue
	if afterProcessQueue {
		processQueue()
	}
	return true
}

func respond(c *gin.Context, code int, res gin.H) {
	c.JSON(code, res)
}

func processQueue() {
	states := getStates()

	var files []string
	err := filepath.Walk(config.dir.queue, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})
	if err != nil {
		panic(err)
	}

	var imagesToCompress []*imageSnapshot

	for index, imageState := range states {

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

	states = getStates()

	for _, imageState := range states {
		if "compressing" == imageState["state"] {
			currentlyCompressing++
		}
	}

	var canCompress = config.simultaneousCompressions - currentlyCompressing

	if config.simultaneousCompressions < currentlyCompressing {
		log.Fatal("Too many compressions happening.")
	}

	log.Printf("imagesToCompress: %d", len(imagesToCompress))

	for i := 0; i < len(imagesToCompress); i++ {
		relImg := imagesToCompress[i]

		var processQueueAfter = false
		if (i == 0 && len(imagesToCompress) == 1) || i == (len(imagesToCompress)-1) {
			processQueueAfter = true
		}

		go relImg.imd.compressAndFinish(relImg.currentPath, relImg.finalPath, processQueueAfter)
		canCompress--
		positionInQueue++
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
	states := getStates()

	stateMap := make(map[string]string)
	stateMap["state"] = state

	states[id.id] = stateMap

	statesJSON, _ := json.MarshalIndent(states, "", " ")

	_ = ioutil.WriteFile(config.stateFile, statesJSON, 0644)
}

func resetCompressionStates() {
	states := getStates()

	queueStateMap := make(map[string]string)
	queueStateMap["state"] = "queue"

	for index, imageState := range states {
		if "compressing" == imageState["state"] {
			states[index] = queueStateMap
		}
	}

	statesJSON, _ := json.MarshalIndent(states, "", " ")

	_ = ioutil.WriteFile(config.stateFile, statesJSON, 0644)
}

func getStates() map[string]map[string]string {
	stateFile, _ := ioutil.ReadFile(config.stateFile)
	data := make(map[string]map[string]string)
	err := json.Unmarshal(stateFile, &data)
	if err != nil {
		log.Fatal(err)
	}

	return data
}

func getAmountOfCurrentCompressions() int {
	states := getStates()
	var count = 0
	for _, imageState := range states {
		if "compressing" == imageState["state"] {
			count++
		}
	}

	return count
}

type imageIdentifier struct {
	id string
}
