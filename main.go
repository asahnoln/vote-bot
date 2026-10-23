package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
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

	dest := tg.NewDest(os.Getenv("BOT_URL"))

	errCh := make(chan error)
	go func() {
		for err := range errCh {
			slog.ErrorContext(ctx, "bot", "err", err)
			if err := sendErrorMsgsToAdmins(ctx, dest, err); err != nil {
				if err != nil {
					slog.ErrorContext(ctx, "send error to telegram", "err", err)
				}
			}
		}
	}()

	store := fstore.New(fc)

	err = setBotCommands(ctx, dest)
	if err != nil {
		log.Fatalf("set bot commands: %v", err)
	}

	b := &tg.Bot{
		Dest:      dest,
		ErrorChan: errCh,
	}
	b.Handle(&tg.StartHandler{Store: store})
	b.Handle(&tg.QuestionHandler{Store: store})
	b.Handle(&tg.AnswerHandler{Store: store})
	b.Handle(&tg.BetHandler{Store: store})
	b.Handle(&tg.CloseHandler{Store: store})
	b.Handle(&tg.RevealHandler{Store: store})
	b.Handle(&tg.StatsHandler{Store: store})
	b.Handle(&tg.DefaultHandler{})

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

func setBotCommands(ctx context.Context, dest tg.Dest) error {
	adminIDs := os.Getenv("ADMIN_IDS")
	ids := strings.Split(adminIDs, ",")
	for _, id := range ids {
		i, err := strconv.Atoi(id)
		if err != nil {
			return fmt.Errorf("atoi id %q: %w", id, err)
		}

		m, err := dest.Send(ctx, tg.CommandsForAdmin(i))
		if err != nil {
			return fmt.Errorf("dest send (resp %#v): %w", m, err)
		}

	}

	return nil
}

func sendErrorMsgsToAdmins(ctx context.Context, dest tg.Dest, errToSend error) error {
	adminIDs := os.Getenv("ADMIN_IDS")
	ids := strings.Split(adminIDs, ",")
	for _, id := range ids {
		i, err := strconv.Atoi(id)
		if err != nil {
			return fmt.Errorf("atoi id %q: %w", id, err)
		}

		m, err := dest.Send(ctx, tg.SendRichMessage{
			ChatID: i,
			RichMessage: tg.InputRichMessage{
				Blocks: []tg.InputRichBlock{
					{
						Type: tg.InputRichBlockParagraphType,
						Text: fmt.Sprintf("Ошибка! %v", errToSend),
					},
				},
			},
		})
		if err != nil {
			return fmt.Errorf("dest send (resp %#v): %w", m, err)
		}

	}

	return nil
}
