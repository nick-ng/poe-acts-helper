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
	"slices"
	"strconv"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

type DataItem struct {
	Zone     string `json:"zone"`
	Level    int    `json:"level"`
	HtmlNote string `json:"htmlNote"`
}

type Note struct {
	Zones    []string
	MinLevel int
	MaxLevel int
	MdNote   string
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

func GetPoe1SteamLogPath() string {
	if runtime.GOOS == "windows" {
		poe1Dir := filepath.Join("C:", "Program Files (x86)", "Steam", "steamapps", "common", "Path of Exile", "logs", "Client.txt")
		return poe1Dir
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("\nCouldn't get home directory", err)
		os.Exit(1)
	}

	poe1Dir := filepath.Join(homeDir, ".steam", "steam", "steamapps", "common", "Path of Exile", "logs", "Client.txt")
	return poe1Dir
}

func GetPoe1StandAlonePath() string {
	if runtime.GOOS == "windows" {
		poe1Dir := filepath.Join("C:", "Program Files (x86)", "Grinding Gear Games", "Path of Exile", "logs", "Client.txt")
		return poe1Dir
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("\nCouldn't get home directory", err)
		os.Exit(1)
	}

	poe1Dir := filepath.Join(homeDir, "Games", "path-of-exile", "drive_c", "Program Files (x86)", "Grinding Gear Games", "Path of Exile", "logs", "Client.txt")
	return poe1Dir
}

var settings = map[string]clientSettings{
	"stand_alone": {
		LogPath:    GetPoe1StandAlonePath(),
		FirstLine:  "",
		ByteOffset: 0,
	},
	"steam": {
		LogPath:    GetPoe1SteamLogPath(),
		FirstLine:  "",
		ByteOffset: 0,
	},
}

var notes = []Note{
	{
		Zones: []string{
			"The Twilight Strand",
			"Lioneye's Watch",
			"The Coast",
			"The Tidal Island",
			"The Mud Flats",
			"The Submerged Passage",
			"The Flooded Depths",
			"The Ledge",
			"The Climb",
			"The Lower Prison",
			"The Upper Prison",
			"The Wardern's Quarters",
			"Prisoner's Gate",
			"The Ship Graveyard",
			"The Ship Graveyard Cave",
			"The Cavern of Wrath",
			"The Cavern of Anger",
			"Mervil's Lair",
		},
		MinLevel: 0,
		MaxLevel: 20,
		MdNote: `
## Act 1

| All                    |
| :--------------------- |
| Solo Coast             |
| Group Mud Flats        |
| Solo Hailrake          |
| Solo Submerged passage |

| David | Jacky        | Mark         | Nick    |
| :---- | :----------- | :----------- | :------ |
| XP    | Lower Prison | XP           | XP      |
| Ledge | Prison Trial | Upper Prison | Dweller |

### Chain 1.1

- Swirl Mark => Brutus
- Portal Jacky => Prison trial
- Portal Nick => Dweller
- **Logout**

| David     | Jacky               | Mark    | Nick |
| :-------- | :------------------ | :------ | :--- |
| XP        | Ship Graveyard      | XP      | XP   |
| All Flame | Ship Graveyard Cave | Merveil | XP   |

| All              |
| :--------------- |
| Fairgraves Trick |
`,
	},
	{
		Zones: []string{
			"Mervil's Lair",
			"The Southern Forest",
			"The Forest Encampment",
			"The Old Fields",
			"The Crossroads",
			"Chamber of Sins Level 1",
			"Chamber of Sins Level 2",
			"The Fellshrine Ruins",
			"The Crypt Level 1",
			"The Crypt Level 2",
			"The Broken Bridge",
			"The Riverways",
			"The Wetlands",
			"The Western Forest",
			"The Weaver's Chamber",
			"Vaal Ruins",
			"Northern Forest",
			"The Dread Thicket",
			"The Caverns",
			"Ancient Pyramid",
		},
		MinLevel: 0,
		MaxLevel: 40,
		MdNote: `
## Act 2

| David         | Jacky        | Mark | Nick                  |
| :------------ | :----------- | :--- | :-------------------- |
| Crossroads WP | Riverways WP | XP   | XP                    |
| Crypt Trial   | Weaver       | Oak  | Chamber of Sins Trial |
| XP            | XP           | XP   | Fidelitas             |

### Chain 2.1

- Swirl Nick => Chamber of Sins trial => Portal to town
- Portal Nick => Baleful Gem
- Portal Jacky => Weaver
- Portal David => Crypt trial
- Portal Mark => Kill Oak
- Run to tree roots

| David       | Jacky | Mark | Nick    |
| :---------- | :---- | :--- | :------ |
| Golden Hand | Alira | Ball | Kraityn |

### Chain 2.2

- Swirl Mark => Ball => run to Northern Forest WP
- Portal Nick => Kill Kraityn
- Portal David => Golden Hand
- Portal Jacky => Help Alira => Run to Thaumatic Emblem
- Someone logout and get to Act 1 => Swirl that person

| Highest Level | Everyone Else |
| :------------ | :------------ |
| Vaal Oversoul | XP            |
`,
	},
}

var actsData = map[string]DataItem{
	"stand_alone": {
		Zone: "Loading", Level: 0, HtmlNote: ""},
	"steam": {
		Zone: "Loading", Level: 0, HtmlNote: ""},
}

func mdToHTML(md []byte) []byte {
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer)
}

func parseLine(logLine string, dataItem *DataItem) (string, int) {
	zoneParts := strings.Split(logLine, "You have entered ")
	if len(zoneParts) > 1 {
		temp := strings.TrimSpace(zoneParts[1])
		dataItem.Zone = strings.TrimSuffix(temp, ".")
	} else {
		levelParts := strings.Split(logLine, "is now level ")
		if len(levelParts) > 1 {
			temp := strings.TrimSpace(levelParts[1])
			tempLevel, err := strconv.Atoi(temp)
			if err != nil {
				log.Println("error converting to integer ", temp, err)
			} else {
				dataItem.Level = tempLevel
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

	actsDatum, ok := actsData[poeClient]
	if !ok {
		return 1
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

		parseLine(tempLine, &actsDatum)

	}

	actsData[poeClient] = actsDatum
	settings[poeClient] = setting

	updateActsNote(poeClient)
	return 0
}

func updateActsNote(poeClient string) {
	actsDatum, ok := actsData[poeClient]
	if !ok {
		return
	}

	fullMd := ""
	matchFound := false
	for _, note := range notes {
		if note.MinLevel <= actsDatum.Level && actsDatum.Level <= note.MaxLevel && slices.Contains(note.Zones, actsDatum.Zone) {
			matchFound = true
			fullMd = fmt.Sprintf("%s\n\n%s", fullMd, note.MdNote)
		}
	}

	if matchFound {
		html := string(mdToHTML([]byte(fullMd)))
		actsDatum.HtmlNote = html
	}

	actsData[poeClient] = actsDatum
}

func handleReset(writer http.ResponseWriter, req *http.Request) {
	requestBody := DataRequest{}
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&requestBody)
	if err != nil {
		log.Println("error reading parameters ", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	poeClient := requestBody.PoeClient
	actsDatum, ok := actsData[poeClient]
	if !ok {
		return
	}

	actsDatum.Level = 1
	actsDatum.Zone = "The Twilight Strand"
	actsData[poeClient] = actsDatum

	updateActsNote(poeClient)
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

func main() {
	http.HandleFunc("GET /data", handleGetData)
	http.HandleFunc("POST /data", handlePostData)
	http.HandleFunc("POST /reset", handleReset)
	http.HandleFunc("GET /note", handleGetNoteList)

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	log.Println("Listening on :3232")
	err := http.ListenAndServe(":3232", nil)
	if err != nil {
		log.Fatal(err)
	}
}
