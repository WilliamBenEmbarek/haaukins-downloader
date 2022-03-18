package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	websocket "github.com/gorilla/websocket"
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
	urlObj, _ := url.Parse("https://"+haaukinsURL);
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

	//for i := range challenges.Values {
	//	handleChallenge(challenges.Values[i].Challenge, path)
	//}
}

func handleChallenge(chal Challenge, rootPath string) {
	path := filepath.Join(rootPath, chal.Challenge.Tag)
	os.MkdirAll(path, os.ModePerm)
	file, err := os.Create(filepath.Join(path, "challenge.json"))
	if err != nil {
		panic(err)
	}
	defer file.Close()
	json.NewEncoder(file).Encode(chal)
}

func recieveChallenges(conn *websocket.Conn) {
	//defer close(done)
	_, challengesJSON, _ := conn.ReadMessage()
	fmt.Printf("%s\n", challengesJSON)
	/*
		if err != nil {
			fmt.Println(err)
			return
		}
		json.Unmarshal([]byte(challengesJSON), &challenges)
		fmt.Printf("%+v\n", challenges)
	*/
}
