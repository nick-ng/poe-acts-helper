package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
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
	"stand_alone": clientSettings{
		LogPath:    "C:\\Program Files (x86)\\Grinding Gear Games\\Path of Exile\\logs\\Client.txt",
		FirstLine:  "",
		ByteOffset: 0},
}

var actsData = map[string]DataItem{
	"stand_alone": DataItem{
		Zone: "Loading", Level: 0},
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

var zoneRe = regexp.MustCompile(`You have entered (.+)\.$`)

func handlePostData(writer http.ResponseWriter, req *http.Request) {
	requestBody := DataRequest{}
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&requestBody)
	if err != nil {
		log.Println("error reading parameters ", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	setting, exists := settings[requestBody.PoeClient]
	if !exists {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	poeLogFile, err := os.OpenFile(setting.LogPath, os.O_RDONLY, 0666)
	if err != nil {
		poeLogFile.Close()
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer poeLogFile.Close()

	reader := bufio.NewReader(poeLogFile)
	// check if same file
	text, err := reader.ReadString('\n')
	if err != nil {
		log.Println("error reading ", setting.LogPath, err)
		return
	}

	// new reader so we start from the start
	reader = bufio.NewReader(poeLogFile)
	if text == setting.FirstLine {
		reader.Discard(setting.ByteOffset)
	} else {
		setting.FirstLine = text
		setting.ByteOffset = 0
	}

	keepGoing := true
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
		tempLine := string(tempBytes)

		foundStrings := zoneRe.FindStringSubmatch(tempLine)

		if len(foundStrings) > 0 {
			log.Println(foundStrings)
		}
	}

	handleGetData(writer, req)
}

func main() {
	http.HandleFunc("GET /data", handleGetData)
	http.HandleFunc("POST /data", handlePostData)

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	log.Println("Listening on :3232")
	err := http.ListenAndServe(":3232", nil)
	if err != nil {
		log.Fatal(err)
	}
}
