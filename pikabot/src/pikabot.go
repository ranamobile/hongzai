package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"net/http"
	"unicode"

	"github.com/nlopes/slack"
)

func main() {
	http.HandleFunc("/", slashCommandHandler)

	fmt.Println("[INFO] Server listening")
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
		fmt.Println("Error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err = verifier.Ensure(); err != nil {
		fmt.Println("Error:", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	fmt.Println("Command:", s.Command, s.Text)
	switch s.Command {
		case "/mark":
			text := []rune(s.Text)

			for index, c := range text {
				if index % 2 == 0 {
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

		default:
			w.WriteHeader(http.StatusInternalServerError)
			return
	}
}
