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
	"time"

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

// @todo(nick-ng): move to separate file
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
			"The Cavern of Anger",
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
	{
		Zones: []string{
			"Ancient Pyramid",
			"The City of Sarn",
			"The Sarn Encampment",
			"The Slums",
			"The Crematorium",
			"The Sewers",
			"The Marketplace",
			"The Catacombs",
			"The Battlefront",
			"Solaris Temple Level 1",
			"Solaris Temple Level 2",
			"The Docks",
			"The Ebony Barracks",
			"Lunaris Temple Level 1",
			"Lunaris Temple Level 2",
			"The Imperial Gardens",
			"The Sceptre of God",
			"The Upper Sceptre of God",
		},
		MinLevel: 0,
		MaxLevel: 45,
		MdNote: `
## Act 3

| Highest Level  | Everyone Else |
| :------------- | :------------ |
| Crematorium WP | XP            |

| David             | Jacky  | Mark | Nick          |
| :---------------- | :----- | :--- | :------------ |
| Crematorium Trial | Tolman | XP   | Sewers Portal |

### Chain 3.1

- Swirl Jacky => Click Tolman
- Portal David => Crematorium trial
- Turn in Bracelet, get Thief Tools
- Portal Nick => Sewers

| All         |
| :---------- |
| Solo Sewers |

| Highest Level  | Everyone Else |
| :------------- | :------------ |
| Marketplace WP | XP            |

| David | Jacky           | Mark | Nick            |
| :---- | :-------------- | :--- | :-------------- |
| XP    | Ribbon Spool WP | XP   | Catacombs Trial |
| Docks | Dialla WP       | XP   | XP              |

### Chain 3.2

- Swirl David => Thaumatic Sulfite
- Portal Nick => Catacombs trial
- Swirl Jacky => Take Infernal Talc
- Sewers => Burn Blockage => Get WP

| Highest Level     | Everyone Else                             |
| :---------------- | :---------------------------------------- |
| General Gravicius | Wait for General Gravicius tag then swirl |

| David | Jacky                | Mark | Nick       |
| :---- | :------------------- | :--- | :--------- |
| Piety | Explosive Concoction | XP   | Tower Door |

### Chain 3.3

- Swirl David => Kill Piety
- Portal Nick => Tower door

| Highest Level | Nick          | Everyone Else |
| :------------ | :------------ | :------------ |
| Dominus       | Gardens trial | XP            |

### Chain 3.4

- Swirl <someone> => Dominus
- Aqueduct WP => Garden Trial

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

func handleSSE(writer http.ResponseWriter, req *http.Request) {
	writer.Header().Set("Access-Control-Allow-Origin", "*")
	writer.Header().Set("Access-Control-Expose-Headers", "Content-Type")

	writer.Header().Set("Content-Type", "text/event-stream")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Connection", "keep-alive")

	flusher, ok := writer.(http.Flusher)
	if !ok {
		fmt.Println("error initialising flusher")
	}

	for i := 0; i < 300; i++ {
		select {
		case <-req.Context().Done():
			{
				fmt.Println("connection closed")
				return
			}
		default:
			{
				fmt.Println("sending event", i)
				fmt.Fprintf(writer, "test\n\n")
				flusher.Flush()
				time.Sleep(2 * time.Second)
			}
		}
	}
}

func main() {
	router := http.NewServeMux()
	router.HandleFunc("GET /data", handleGetData)
	router.HandleFunc("POST /data", handlePostData)
	router.HandleFunc("POST /reset", handleReset)
	router.HandleFunc("GET /note", handleGetNoteList)
	router.HandleFunc("/events", handleSSE)

	fs := http.FileServer(http.Dir("./static"))
	router.Handle("/", fs)

	log.Println("Listening on :3232")
	err := http.ListenAndServe(":3232", router)
	if err != nil {
		log.Fatal(err)
	}
}
