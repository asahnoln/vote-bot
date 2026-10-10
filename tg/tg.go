package tg

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
)

const ApplicationJSONContentType = "application/json"

const (
	SendRichMessageMethod = "sendRichMessage"
)

type SendRichMessage struct {
	ChatID      int              `json:"chat_id"`
	RichMessage InputRichMessage `json:"rich_message"`
}

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
	Message Message
}

type Message struct {
	Chat Chat
	Text string
}

type Chat struct {
	ID int
}

type Question struct {
	Q    string
	Opts []string
}

type CurrentQuestioner interface {
	CurrentQuestion(context.Context) (Question, error)
}

type Response struct {
	OK          bool
	Description string
}

type Bot struct {
	Store     CurrentQuestioner
	URL       string
	ErrorChan chan<- error
}

var ErrStoreNil = errors.New("store is nil")

func (b *Bot) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	srv := http.NewServeMux()

	slog.InfoContext(r.Context(), "request", "r", r)

	srv.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	srv.Handle("/webhook", http.HandlerFunc(b.question))
	srv.ServeHTTP(w, r)
}

func (b *Bot) error(err error, w http.ResponseWriter, s int) {
	w.WriteHeader(s)
	if b.ErrorChan == nil {
		return
	}

	b.ErrorChan <- err
}
