package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strings"
	"unicode"

	"github.com/nlopes/slack"
	"lib"
)

func main() {
	http.HandleFunc("/", slashCommandHandler)

	log.Println("[INFO] Server listening")
	http.ListenAndServe(":8080", nil)
}

func slashCommandHandler(w http.ResponseWriter, r *http.Request) {
	signingSecret := os.Getenv("SLACK_SIGNING_SECRET")
	verifier, err := slack.NewSecretsVerifier(r.Header, signingSecret)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	r.Body = ioutil.NopCloser(io.TeeReader(r.Body, &verifier))
	s, err := slack.SlashCommandParse(r)
	if err != nil {
		log.Println("Slack error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err = verifier.Ensure(); err != nil {
		log.Println("Slack error:", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	log.Println("Command:", s.Command, s.Text)
	switch s.Command {
	case "/mark":
		text := []rune(s.Text)

		for index, c := range text {
			if rand.Int()%2 == 0 {
				text[index] = unicode.ToUpper(c)
			} else {
				text[index] = unicode.ToLower(c)
			}
		}
		params := &slack.Msg{
			Text:         string(text),
			ResponseType: slack.ResponseTypeInChannel,
		}
		b, err := json.Marshal(params)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)

	case "/score":
		scorelist := lib.ScoreList{
			Scores: []lib.Score{},
		}
		scorelist.Read(os.Getenv("SLACK_PIKA_SCOREFILE"))

		name := s.Text[:len(s.Text)-2]
		count := s.Text[len(s.Text)-2:]
		text := "invalid command"

		if s.Text == "top" {
			sort.Sort(sort.Reverse(scorelist))

			var topScores []string
			for index := 0; index < 3; index++ {
				topScores = append(topScores, fmt.Sprintf("%s: %d", scorelist.Scores[index].Name, scorelist.Scores[index].Count))
			}
			text = strings.Join(topScores, "\n")
		} else if count == "++" {
			score := scorelist.Increment(name)
			text = fmt.Sprintf("%s: %d", score.Name, score.Count)
		} else if count == "--" {
			score := scorelist.Decrement(name)
			text = fmt.Sprintf("%s: %d", score.Name, score.Count)
		}

		log.Println("Writing scores to file:", os.Getenv("SLACK_PIKA_SCOREFILE"))
		scorelist.Write(os.Getenv("SLACK_PIKA_SCOREFILE"))

		params := &slack.Msg{
			Text:         text,
			ResponseType: slack.ResponseTypeInChannel,
		}
		b, err := json.Marshal(params)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)

	default:
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
