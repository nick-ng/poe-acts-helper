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
	"strconv"
	"strings"
)

type DataItem struct {
	Zone  string `json:"zone"`
	Level int    `json:"level"`
}

type DataResponse struct {
	StandAlone DataItem `json:"stand_alone"`
}

type DataRequest struct {
	PoeClient string `json:"poe_client"`
}

type clientSettings struct {
	LogPath    string
	FirstLine  string
	ByteOffset int
}

var settings = map[string]clientSettings{
	"stand_alone": {
		LogPath:    "C:\\Program Files (x86)\\Grinding Gear Games\\Path of Exile\\logs\\Client.txt",
		FirstLine:  "",
		ByteOffset: 0,
	},
	"steam": {
		LogPath:    "C:\\Program Files (x86)\\Steam\\steamapps\\common\\Path of Exile\\logs\\Client.txt",
		FirstLine:  "",
		ByteOffset: 0,
	},
}

var actsData = map[string]DataItem{
	"stand_alone": {
		Zone: "Loading", Level: 0},
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

func handleGetHelperList(writer http.ResponseWriter, req *http.Request) {
	dirEntries, err := os.ReadDir("static/helpers")
	if err != nil {
		log.Println("error listing helpers ", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	helperPaths := []string{}
	for _, dirEntry := range dirEntries {
		fullPath := fmt.Sprintf("helpers/%s", dirEntry.Name())
		helperPaths = append(helperPaths, fullPath)
	}

	jsonBytes, err := json.Marshal(helperPaths)
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

func main() {
	http.HandleFunc("GET /data", handleGetData)
	http.HandleFunc("POST /data", handlePostData)
	http.HandleFunc("GET /helper", handleGetHelperList)

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	log.Println("Listening on :3232")
	err := http.ListenAndServe(":3232", nil)
	if err != nil {
		log.Fatal(err)
	}
}
