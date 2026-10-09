package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/asahnoln/vote-bot/tg"
)

func main() {
	// TODO: Create http server for context
	ctx := context.Background()
	errCh := make(chan error)
	go func() {
		for err := range errCh {
			slog.ErrorContext(ctx, "bot", "err", err)
		}
	}()

	b := &tg.Bot{
		URL:       os.Getenv("BOT_URL"),
		Store:     &inMemoryStore{},
		ErrorChan: errCh,
	}

	err := http.ListenAndServe(":8080", b)
	if err != nil {
		slog.ErrorContext(ctx, "http srv", "err", err)
	}
}

type inMemoryStore struct{}

func (s *inMemoryStore) CurrentQuestion(context.Context) (tg.Question, error) {
	return tg.Question{
		Q:    "Кто убийца ваших снов и мечтаний?",
		Opts: []string{"Нечаев", "П-3", "Сеченов", "Левая"},
	}, nil
}
