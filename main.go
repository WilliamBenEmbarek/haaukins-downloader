package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	websocket "github.com/gorilla/websocket"
	"mvdan.cc/xurls/v2"
)

type Challenges struct {
	Msg           string      `json:"msg"`
	Values        []Challenge `json:"values"`
	IsLabAssigned bool        `json:"isLabAssigned"`
}

type Challenge struct {
	Challenge struct {
		Tag             string `json:"tag"`
		Name            string `json:"name"`
		Points          int    `json:"points"`
		Category        string `json:"category"`
		TeamDescription string `json:"teamDescription"`
		StaticChallenge bool   `json:"staticChallenge"`
	} `json:"challenge"`
	IsUserCompleted bool `json:"isUserCompleted"`
	TeamsCompleted  []struct {
		TeamName    string    `json:"teamName"`
		CompletedAt time.Time `json:"completedAt"`
	} `json:"teamsCompleted"`
	IsChalDisabled bool `json:"isChalDisabled"`
}

var done chan interface{}
var interrupt chan os.Signal
var haaukinsURL string
var haaukinsSessionCookie string
var filePath string

var challenges Challenges

var cookieJar cookiejar.Jar
var haaukinsDialer websocket.Dialer

func init() {
	flag.StringVar(&haaukinsURL, "url", "http://localhost", "Haaukins url fx: ddc.haaukins.com")
	flag.StringVar(&haaukinsSessionCookie, "session", "", "Haaukins session cookie")
	flag.StringVar(&filePath, "file", "challenges", "Path to save challenges, default is local subdirectory challenges")

	flag.Parse()
	done = make(chan interface{})          //Channel to indicate when the connection is closed
	interrupt = make(chan os.Signal, 1)    // Channel to listen for interrupts
	signal.Notify(interrupt, os.Interrupt) // Catch OS interrupts

	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}
	cookie := &http.Cookie{
		Name:  "session",
		Value: haaukinsSessionCookie,
	}
	urlObj, _ := url.Parse("https://" + haaukinsURL)
	cookieJar.SetCookies(urlObj, []*http.Cookie{cookie})
	haaukinsDialer = websocket.Dialer{
		Jar: cookieJar,
	}
}

func main() {

	socketURL := "wss://" + haaukinsURL + "/challengesFrontend"
	conn, resp, err := haaukinsDialer.Dial(socketURL, nil)
	if err == websocket.ErrBadHandshake {
		log.Printf("Bad handshake: %d", resp.StatusCode)
	}
	if err != nil {
		panic(err)
	}

	defer conn.Close()
	recieveChallenges(conn)

	path := filepath.Join(".", filePath)
	os.MkdirAll(path, os.ModePerm)

	var wg sync.WaitGroup
	wg.Add(len(challenges.Values))
	log.Println("Downloading files...")
	for i := range challenges.Values {
		go func(chal Challenge) {
			defer wg.Done()
			handleChallenge(chal, path)
			downloadFileIfExists(chal, path)
		}(challenges.Values[i])
	}
	wg.Wait()
	log.Println("Done! Files saved to:", path)
}

func downloadFileIfExists(chal Challenge, rootPath string) {
	fileUrl := xurls.Strict().FindString(chal.Challenge.TeamDescription)
	client := &http.Client{}

	fsPath := filepath.Join(rootPath, chal.Challenge.Tag)
	os.MkdirAll(fsPath, os.ModePerm)

	// Check for false positives
	if fileUrl != "" && !strings.Contains(fileUrl, "hkn") {
		req, err := http.NewRequest("GET", fileUrl, nil)
		if err != nil {
			panic(err)
		}
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		filePath := path.Base(req.URL.Path)
		if filePath == "download" {
			filePath = chal.Challenge.Tag
		}
		log.Println("Downloading file:", filePath)
		file, err := os.Create(filepath.Join(fsPath, filePath))
		if err != nil {
			panic(err)
		}
		defer file.Close()
		_, err = io.Copy(file, resp.Body)
		if err != nil {
			panic(err)
		}
	}
}

func handleChallenge(chal Challenge, rootPath string) {
	path := filepath.Join(rootPath, chal.Challenge.Tag)
	os.MkdirAll(path, os.ModePerm)
	file, err := os.Create(filepath.Join(path, "challenge.json"))
	if err != nil {
		panic(err)
	}
	defer file.Close()
	enc := json.NewEncoder(file)
	enc.SetIndent("", "    ")
	enc.Encode(chal)
}

func recieveChallenges(conn *websocket.Conn) {
	defer close(done)
	_, challengesJSON, err := conn.ReadMessage()

	if err != nil {
		fmt.Println(err)
		return
	}
	json.Unmarshal([]byte(challengesJSON), &challenges)
}
