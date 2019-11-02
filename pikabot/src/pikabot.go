package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/nlopes/slack"
)

type Score struct {
	name string
	count int
}

type ScoreList struct {
	scores []Score
}

func (scorelist ScoreList) Len() int {
	return len(scorelist.scores)
}

func (scorelist ScoreList) Less(i, j int) bool {
	return scorelist.scores[i].count < scorelist.scores[j].count
}

func (scorelist ScoreList) Swap(i, j int) {
	scorelist.scores[i], scorelist.scores[j] = scorelist.scores[j], scorelist.scores[i]
}

func (scorelist *ScoreList) Read(filename string) {
	handler, err := os.Open(filename)
	if err != nil {
		log.Println("Read error:", err)
		return
	}
	defer handler.Close()

	reader := csv.NewReader(handler)
	reader.FieldsPerRecord = 2

	for {
		line, err := reader.Read()
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		count, err := strconv.Atoi(line[1])
		if err != nil {
			log.Println("Read error:", err)
			continue
		}
		scorelist.Update(line[0], count)
	}
}

func (scorelist *ScoreList) Write(filename string) {
	handler, err := os.Create(filename)
	if err != nil {
		log.Println("Write error:", err)
		return
	}
	defer handler.Close()

	writer := csv.NewWriter(handler)
	defer writer.Flush()

	log.Println("Scores:", len(scorelist.scores))
	for _, score := range scorelist.scores {
		log.Println("Writing score:", score.name, score.count)
		writer.Write([]string{score.name, strconv.Itoa(score.count)})
	}
	writer.Flush()
}

func (scorelist *ScoreList) Find(name string) int {
	for index, score := range scorelist.scores {
		if score.name == name {
			return index
		}
	}
	return -1
}

func (scorelist *ScoreList) Update(name string, count int) Score {
	index := scorelist.Find(name)
	if index < 0 {
		index = len(scorelist.scores)
		scorelist.scores = append(scorelist.scores, Score{name, count})
	} else {
		scorelist.scores[index].count = count
	}
	return scorelist.scores[index]
}

func (scorelist *ScoreList) Increment(name string) Score {
	index := scorelist.Find(name)
	if index < 0 {
		index = len(scorelist.scores)
		scorelist.scores = append(scorelist.scores, Score{name, 1})
	} else {
		scorelist.scores[index].count++
	}
	return scorelist.scores[index]
}

func (scorelist *ScoreList) Decrement(name string) Score {
	index := scorelist.Find(name)
	if index < 0 {
		index = len(scorelist.scores)
		scorelist.scores = append(scorelist.scores, Score{name, -1})
	} else {
		scorelist.scores[index].count--
	}
	return scorelist.scores[index]
}

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
				if rand.Int() % 2 == 0 {
					text[index] = unicode.ToUpper(c)
				} else {
					text[index] = unicode.ToLower(c)
				}
			}
			params := &slack.Msg{
				Text: string(text),
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
			scorelist := ScoreList{
				scores: []Score{},
			}
			scorelist.Read(os.Getenv("SLACK_PIKA_SCOREFILE"))

			name := s.Text[:len(s.Text) - 2]
			count := s.Text[len(s.Text) - 2:]
			text := "invalid command"

			if s.Text == "top" {
				sort.Sort(sort.Reverse(scorelist))

				var topScores []string
				for index := 0; index < 3; index++ {
					topScores = append(topScores, fmt.Sprintf("%s: %d", scorelist.scores[index].name, scorelist.scores[index].count))
				}
				text = strings.Join(topScores, "\n")
			} else if count == "++" {
				score := scorelist.Increment(name)
				text = fmt.Sprintf("%s: %d", score.name, score.count)
			} else if count == "--" {
				score := scorelist.Decrement(name)
				text = fmt.Sprintf("%s: %d", score.name, score.count)
			}

			log.Println("Writing scores to file:", os.Getenv("SLACK_PIKA_SCOREFILE"))
			scorelist.Write(os.Getenv("SLACK_PIKA_SCOREFILE"))

			params := &slack.Msg{
				Text: text,
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
