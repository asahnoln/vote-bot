package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/asahnoln/vote-bot/tg"
)

func main() {
	// TODO: Create http server for context
	errCh := make(chan error)
	go func() {
		for err := range errCh {
			log.Printf("bot err: %v", err)
		}
	}()

	b := &tg.Bot{
		URL:       os.Getenv("BOT_URL"),
		Store:     &inMemoryStore{},
		ErrorChan: errCh,
	}

	err := http.ListenAndServe(":8080", b)
	if err != nil {
		log.Fatalf("http err: %v", err)
	}
}

type inMemoryStore struct{}

func (s *inMemoryStore) CurrentQuestion(context.Context) (tg.Question, error) {
	return tg.Question{
		Q:    "Кто убийца ваших снов и мечтаний?",
		Opts: []string{"Нечаев", "П-3", "Сеченов", "Левая"},
	}, nil
}
