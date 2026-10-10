package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"github.com/asahnoln/vote-bot/store"
	"github.com/asahnoln/vote-bot/store/fstore"
	"github.com/asahnoln/vote-bot/tg"
	"google.golang.org/api/option"
)

func main() {
	ctx := context.Background()

	fc, err := initFirestoreClient(ctx)
	if err != nil {
		log.Fatalf("init firebase client error: %v", err)
	}

	errCh := make(chan error)
	go func() {
		for err := range errCh {
			slog.ErrorContext(ctx, "bot", "err", err)
		}
	}()

	b := &tg.Bot{
		URL:       os.Getenv("BOT_URL"),
		Store:     fstore.New(fc),
		ErrorChan: errCh,
	}

	// TODO: Create http server for context
	err = http.ListenAndServe(":8080", b)
	if err != nil {
		log.Fatalf("http listen and server error: %v", err)
	}
}

func initFirestoreClient(ctx context.Context) (*firestore.Client, error) {
	app, err := firebase.NewApp(ctx, nil, option.WithAuthCredentialsFile(option.ServiceAccount, os.Getenv("FIREBASE_CREDS_PATH")))
	if err != nil {
		return nil, fmt.Errorf("new app: %w", err)
	}

	fc, err := app.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("app firestore: %w", err)
	}

	return fc, err
}

type inMemoryStore struct{}

func (s *inMemoryStore) CurrentQuestion(context.Context) (store.Question, error) {
	return store.Question{
		Q:    "Кто убийца ваших снов и мечтаний?",
		Opts: []string{"Нечаев", "П-3", "Сеченов", "Левая"},
	}, nil
}
