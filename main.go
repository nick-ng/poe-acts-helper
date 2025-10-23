package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

type DataItem struct {
	LogsPath     string `json:"-"`
	Zone         string `json:"zone"`
	Level        int    `json:"level"`
	HtmlNote     string `json:"htmlNote"`
	LastUpdateMs int64  `json:"lastUpdateMs"`
}

type DataResponse struct {
	StandAlone DataItem `json:"stand_alone"`
}

type DataRequest struct {
	PoeClient string `json:"poe_client"`
}

type clientSettings struct {
	LogPath      string
	LinuxLogPath string
	FirstLine    string
	ByteOffset   int
}

var settings = map[string]clientSettings{
	"stand_alone": {
		LogPath:      "C:\\Program Files (x86)\\Grinding Gear Games\\Path of Exile\\logs\\Client.txt",
		LinuxLogPath: "",
		FirstLine:    "",
		ByteOffset:   0,
	},
	"steam": {
		LogPath:      "C:\\Program Files (x86)\\Steam\\steamapps\\common\\Path of Exile\\logs\\Client.txt",
		LinuxLogPath: "",
		FirstLine:    "",
		ByteOffset:   0,
	},
}

func GetPoe1SteamLogPath() string {
	if runtime.GOOS == "windows" {
		poe1Dir := filepath.Join("C:", "Program Files (x86)", "Steam", "steamapps", "common", "Path of Exile", "logs")
		return poe1Dir
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("\nCouldn't get home directory", err)
		os.Exit(1)
	}

	poe1Dir := filepath.Join(homeDir, ".steam", "steam", "steamapps", "common", "Path of Exile", "logs")
	return poe1Dir
}

func GetPoe1StandAlonePath() string {
	if runtime.GOOS == "windows" {
		poe1Dir := filepath.Join("C:", "Program Files (x86)", "Grinding Gear Games", "Path of Exile", "logs")
		return poe1Dir
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("\nCouldn't get home directory", err)
		os.Exit(1)
	}

	poe1Dir := filepath.Join(homeDir, "Games", "path-of-exile", "drive_c", "Program Files (x86)", "Grinding Gear Games", "Path of Exile", "logs")
	return poe1Dir
}

var actsData = map[string]DataItem{
	"stand_alone": {
		LogsPath: GetPoe1StandAlonePath(),
		Zone:     "Loading",
		Level:    1,
	},
	"steam": {
		LogsPath: GetPoe1SteamLogPath(),
		Zone:     "Loading",
		Level:    1,
	},
}

func parseLine(logLine string, dataItem *DataItem) (string, int) {
	zoneParts := strings.Split(logLine, "You have entered ")
	if len(zoneParts) > 1 {
		temp := strings.TrimSpace(zoneParts[1])
		dataItem.Zone = strings.TrimSuffix(temp, ".")
		dataItem.LastUpdateMs = time.Now().UnixMilli()
	} else {
		levelParts := strings.Split(logLine, "is now level ")
		if len(levelParts) > 1 {
			temp := strings.TrimSpace(levelParts[1])
			tempLevel, err := strconv.Atoi(temp)
			if err != nil {
				log.Println("error converting to integer ", temp, err)
			} else {
				dataItem.Level = tempLevel
				dataItem.LastUpdateMs = time.Now().UnixMilli()
			}
		}
	}

	return dataItem.Zone, dataItem.Level
}

func updateActsData(poeClient string) int {
	setting, exists := settings[poeClient]
	if !exists {
		return http.StatusNotFound
	}

	poeLogFile, err := os.OpenFile(setting.LogPath, os.O_RDONLY, 0666)
	if err != nil {
		poeLogFile.Close()
		return http.StatusInternalServerError
	}
	defer poeLogFile.Close()

	reader := bufio.NewReader(poeLogFile)
	// check if same file
	text, err := reader.ReadString('\n')
	if err != nil {
		log.Println("error reading ", setting.LogPath, err)
		return http.StatusInternalServerError
	}

	// new reader so we start from the start
	reader = bufio.NewReader(poeLogFile)
	if text == setting.FirstLine {
		reader.Discard(setting.ByteOffset)
	} else {
		setting.FirstLine = text
		setting.ByteOffset = 0
	}

	tempEntry, ok := actsData[poeClient]
	zone := ""
	level := 0
	if ok {
		zone = tempEntry.Zone
		level = tempEntry.Level
	}

	keepGoing := true
	bytesRead := 0
	for keepGoing {
		tempBytes, err := reader.ReadBytes('\n')
		if err != nil {
			keepGoing = false

			if !errors.Is(err, io.EOF) {
				log.Println("error:", err)
			}

			continue
		}

		setting.ByteOffset = setting.ByteOffset + len(tempBytes)
		bytesRead = bytesRead + len(tempBytes)
		tempLine := string(tempBytes)

		zoneParts := strings.Split(tempLine, "You have entered ")
		if len(zoneParts) > 1 {
			temp := strings.TrimSpace(zoneParts[1])
			zone = strings.TrimSuffix(temp, ".")
		} else {
			levelParts := strings.Split(tempLine, "is now level ")
			if len(levelParts) > 1 {
				temp := strings.TrimSpace(levelParts[1])
				level, err = strconv.Atoi(temp)
				if err != nil {
					log.Println("error converting to integer ", temp, err)
					level = 0
				}
			}
		}
	}

	newEntry := DataItem{
		Zone:  zone,
		Level: level,
	}

	actsData[poeClient] = newEntry
	settings[poeClient] = setting

	return 0
}

func handleGetData(writer http.ResponseWriter, req *http.Request) {
	jsonBytes, err := json.Marshal(actsData)
	if err != nil {
		log.Println("error converting acts data to json ", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.Header().Add("Content-Type", "application/json; charset=utf-8")
	writer.Header().Add("Cache-Control", "no-store")
	writer.WriteHeader(http.StatusOK)
	writer.Write(jsonBytes)
}

func handlePostData(writer http.ResponseWriter, req *http.Request) {
	requestBody := DataRequest{}
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&requestBody)
	if err != nil {
		log.Println("error reading parameters ", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	updateActsData(requestBody.PoeClient)

	handleGetData(writer, req)
}

func handleGetNoteList(writer http.ResponseWriter, req *http.Request) {
	dirEntries, err := os.ReadDir("static/notes")
	if err != nil {
		log.Println("error listing notes ", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	notePaths := []string{}
	for _, dirEntry := range dirEntries {
		fullPath := fmt.Sprintf("/notes/%s", dirEntry.Name())
		notePaths = append(notePaths, fullPath)
	}

	jsonBytes, err := json.Marshal(notePaths)
	if err != nil {
		log.Println("error converting acts data to json ", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.Header().Add("Content-Type", "application/json; charset=utf-8")
	writer.Header().Add("Cache-Control", "no-store")
	writer.WriteHeader(http.StatusOK)
	writer.Write(jsonBytes)
}

func listenToLogs() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("error creating watcher", err)
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) {
					for _, actsDatum := range actsData {
						if strings.HasPrefix(event.Name, actsDatum.LogsPath) {
							fmt.Println(event.Name)
						}
					}
					if !strings.HasPrefix(event.Name, path1) {
						processAllFilters(path1, path2)
						return
					}
					if strings.HasSuffix(event.Name, ".filter") {
						filterPath := event.Name
						fmt.Printf("%s modified, reprocessing... ", filterPath)
						start := time.Now()
						temp := strings.Split(filterPath, "/")
						filterName := temp[1]
						filter, filterFlags, errList := processFilter(filterPath, false)

						filterData := []byte(filter)

						outputFilterPath := filepath.Join(OUTPUT_FILTERS_PATH, filterName)

						if filterName != "example.filter" && filterName != "example2.filter" {
							if filterFlags.Game == "poe2" {
								poe2FilterName := fmt.Sprintf("poe2-%s", filterName)
								outputFilterPath = filepath.Join(OUTPUT_FILTERS_PATH, poe2FilterName)

								gameFilterPath := utils.GetPoe2SteamPath(filterName)
								err = os.WriteFile(gameFilterPath, filterData, 0666)
								if err != nil {
									fmt.Println("\nError writing filter to PoE 2 directory", filterName, err)
									return
								}
							} else {
								gameFilterPath := utils.GetPoe1SteamPath(filterName)
								err = os.WriteFile(gameFilterPath, filterData, 0666)
								if err != nil {
									fmt.Println("\nError writing filter to PoE 1, Steam directory", filterName, err)
									return
								}

								gameFilterPath = utils.GetPoe1LutrisPath(filterName)
								err = os.WriteFile(gameFilterPath, filterData, 0666)
								if err != nil {
									fmt.Println("\nError writing filter to PoE 1, Lutris Steam directory", filterName, err)
									return
								}
							}
						}

						err := os.WriteFile(outputFilterPath, filterData, 0666)
						if err != nil {
							fmt.Println("\nError writing filter to output-filters", filterName, err)
							return
						}

						copySounds()

						if len(errList) > 1 {
							fmt.Printf("done with %d errors\n", len(errList))
							for err := range errList {
								fmt.Println(err)
							}
						} else if len(errList) == 1 {
							fmt.Println("done with 1 error")
							fmt.Println(errList[0])
						} else {
							elapsed := time.Since(start)
							fmt.Printf("done in %d ms\n", int(elapsed.Milliseconds()))
						}
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(path1)
	if err != nil {
		fmt.Println("error adding path to watch", err)
	}

	err = watcher.Add(baseFiltersPath)
	if err != nil {
		fmt.Println("error adding path to watch", err)
	}

	err = watcher.Add(thirdPartyFiltersPath)
	if err != nil {
		fmt.Println("error adding path to watch", err)
	}

	<-make(chan struct{})
}

func main() {
	http.HandleFunc("GET /data", handleGetData)
	http.HandleFunc("POST /data", handlePostData)
	http.HandleFunc("GET /note", handleGetNoteList)

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	log.Println("Listening on :3232")
	err := http.ListenAndServe(":3232", nil)
	if err != nil {
		log.Fatal(err)
	}
}
