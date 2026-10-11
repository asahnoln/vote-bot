package tg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/asahnoln/vote-bot/store"
)

const ApplicationJSONContentType = "application/json"

type InputRichMessage struct {
	Blocks []InputRichBlock `json:"blocks"`
}

type InputRichBlock struct {
	Type    string              `json:"type"`
	Text    string              `json:"text"`
	Align   string              `json:"align"`
	Buttons []RichMessageButton `json:"buttons"`
}

const (
	InputRichBlockParagraphType = "paragraph"
	InputRichBlockButtonsType   = "buttons"
)

const (
	InputRichBlockButtonsAlignCenter = "center"
)

type RichMessageButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data"`
}

type Update struct {
	Message       Message
	CallbackQuery *CallbackQuery `json:"callback_query"`
}
type CallbackQuery struct {
	// NOTE: Actually MaybeInaccessibleMessage, probably should work it out
	Message Message
	Data    string
}

type Message struct {
	MessageID int `json:"message_id"`
	Chat      Chat
	Text      string
}

type Chat struct {
	ID int
}

type Store interface {
	CurrentQuestion(context.Context) (store.Question, error)
	SaveAnswer(ctx context.Context, userID int, answer string) error
}

type Response struct {
	OK          bool
	Description string
}

type Sender interface {
	Send(ctx context.Context, data URLer) error
}

type Bot struct {
	Store     Store
	Dest      Sender
	ErrorChan chan<- error
}

var ErrStoreNil = errors.New("store is nil")

func (b *Bot) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	srv := http.NewServeMux()

	slog.InfoContext(r.Context(), "request", "r", r)

	srv.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	srv.Handle("/webhook", http.HandlerFunc(b.webhook))

	srv.ServeHTTP(w, r)
}

func (b *Bot) webhook(w http.ResponseWriter, r *http.Request) {
	upd, err := parseUpdate(w, r)
	if err != nil {
		b.error(err, w, http.StatusInternalServerError)
		return
	}

	// TODO: Handle this beautifully
	proc := b.question
	procName := "question"
	if upd.CallbackQuery != nil {
		proc = b.answer
		procName = "answer"
	}

	msg, err := proc(w, r, upd)
	if err != nil {
		b.error(fmt.Errorf("proc %q: %w", procName, err), w, http.StatusInternalServerError)
		return
	}

	err = b.Dest.Send(r.Context(), msg)
	if err != nil {
		b.error(fmt.Errorf("dest send: %w", err), w, http.StatusInternalServerError)
		return
	}
}

func parseUpdate(w http.ResponseWriter, r *http.Request) (upd Update, err error) {
	updBody, err := io.ReadAll(r.Body)
	// TODO: Test for this error
	if err != nil {
		return upd, fmt.Errorf("update read: %w", err)
	}
	defer r.Body.Close()

	slog.InfoContext(r.Context(), "update", "upd", updBody)

	err = json.Unmarshal(updBody, &upd)
	if err != nil {
		return upd, fmt.Errorf("update decode: %w", err)
	}

	return upd, nil
}

func (b *Bot) error(err error, w http.ResponseWriter, s int) {
	w.WriteHeader(s)
	if b.ErrorChan == nil {
		return
	}

	b.ErrorChan <- err
}
