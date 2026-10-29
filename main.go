package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
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

	fireStore := fstore.New(fc)
	dest := tg.NewDest(os.Getenv("BOT_URL"))

	errCh := make(chan error)
	go func() {
		for err := range errCh {
			slog.ErrorContext(ctx, "bot", "err", err)
			if err := sendErrorMsgsToAdmins(ctx, fireStore, dest, err); err != nil {
				if err != nil {
					slog.ErrorContext(ctx, "send error to telegram", "err", err)
				}
			}
		}
	}()

	err = setBotCommands(ctx, fireStore, dest)
	if err != nil {
		log.Fatalf("set bot commands: %v", err)
	}

	b := &tg.Bot{
		Dest:      dest,
		ErrorChan: errCh,
	}
	b.Handle(&tg.StartHandler{Store: fireStore, Text: tg.StartHandlerText{}})
	b.Handle(
		&tg.StateWrapper{
			Store:   fireStore,
			State:   store.StateQuestion,
			Handler: &tg.QuestionHandler{Store: fireStore},
		})
	b.Handle(&tg.AnswerHandler{Store: fireStore})
	b.Handle(&tg.BetHandler{Store: fireStore})
	b.Handle(&tg.CloseHandler{Store: fireStore})
	b.Handle(&tg.RevealHandler{Store: fireStore})
	b.Handle(&tg.StatsHandler{Store: fireStore})
	b.Handle(&tg.ModeHandler{Store: fireStore, Mode: store.TutorialMode})
	b.Handle(&tg.ModeHandler{Store: fireStore, Mode: store.RealMode})
	b.Handle(&tg.DefaultHandler{})

	srv := http.Server{
		Addr:        ":8080",
		Handler:     b,
		BaseContext: func(l net.Listener) context.Context { return ctx },
	}
	err = srv.ListenAndServe()
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

func setBotCommands(ctx context.Context, fStore *fstore.Firestore, dest tg.Dest) error {
	admins, err := fStore.Admins(ctx)
	if err != nil {
		return fmt.Errorf("fstore admins: %w", err)
	}

	for _, admin := range admins {
		m, err := dest.Send(ctx, tg.CommandsForAdmin(admin.UserID))
		if err != nil {
			return fmt.Errorf("dest send (resp %#v): %w", m, err)
		}

	}

	return nil
}

func sendErrorMsgsToAdmins(ctx context.Context, fStore *fstore.Firestore, dest tg.Dest, errToSend error) error {
	admins, err := fStore.Admins(ctx)
	if err != nil {
		return fmt.Errorf("fstore admins: %w", err)
	}

	for _, admin := range admins {
		m, err := dest.Send(ctx, tg.SendRichMessage{
			ChatID: admin.UserID,
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
