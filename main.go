package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ChimeraCoder/anaconda"
	"github.com/joho/godotenv"
)

type GithubEvents struct {
	Etype     string  `json:"type"`
	Payload   Payload `json:"payload"`
	CreatedAt string  `json:"created_at"`
}

type Payload struct {
	PushId int `json:"pushid"`
	Size   int `json:"size"`
}

func init() {
	env_err := godotenv.Load()
	if env_err != nil {
		log.Fatal("Error loading .env file")
	}
}

func getTwitterApi() *anaconda.TwitterApi {
	anaconda.SetConsumerKey(os.Getenv("CONSUMER_KEY"))
	anaconda.SetConsumerSecret(os.Getenv("CONSUMER_SECRET"))
	return anaconda.NewTwitterApi(os.Getenv("ACCESS_TOKEN"), os.Getenv("ACCESS_TOKEN_SECRET"))
}

func main() {

	github_account := os.Getenv("GITHUB_ACCOUNT_NAME")
	resp, err := http.Get(fmt.Sprintf("https://api.github.com/users/%s/events", github_account))
	if err != nil {
		log.Fatal("http error", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("read error", err)
	}

	var g []GithubEvents
	json_err := json.Unmarshal(body, &g)
	if json_err != nil {
		log.Fatal("unmarshal error", json_err)
	}

	commits_sum := 0
	for _, event := range g {
		if event.Etype == "PushEvent" {
			t, _ := time.Parse(time.RFC3339, event.CreatedAt)
			t_now := time.Now()
			t_7ago := t_now.AddDate(0, 0, -7)

			if t_7ago.Equal(t) || t_7ago.Before(t) {
				commits_sum += event.Payload.Size
			}
		}
	}

	t_api := getTwitterApi()

	repository_url := fmt.Sprintf("https://github.com/%s", github_account)
	text := fmt.Sprintf("今週のcommit数は %v です。\n %s", commits_sum, repository_url)

	tweet, err := t_api.PostTweet(text, nil)
	if err != nil {
		log.Fatal("tweet error", err)
	}
	fmt.Println(tweet.Text)
}
