package tg

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

const ApplicationJSONContentType = "application/json"

type InputRichMessage struct {
	Blocks []InputRichBlock `json:"blocks"`
}

type (
	InputRichBlockType string
	Align              string
)

const (
	InputRichBlockParagraphType InputRichBlockType = "paragraph"
	InputRichBlockButtonsType   InputRichBlockType = "buttons"
)

const (
	AlignCenter Align = "center"
	AlignLeft   Align = "left"
)

type InputRichBlock struct {
	Type      InputRichBlockType     `json:"type"`
	Text      string                 `json:"text"`
	Align     Align                  `json:"align"`
	Buttons   []RichMessageButton    `json:"buttons"`
	Cells     [][]RichBlockTableCell `json:"cells"`
	Caption   string                 `json:"caption"`
	IsStriped bool                   `json:"is_striped"`
}

type RichMessageButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data"`
}

type RichBlockTableCell struct {
	Text     string `json:"text"`
	Align    Align  `json:"align"`
	IsHeader bool   `json:"is_header"`
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
	MessageID   int `json:"message_id"`
	Chat        Chat
	Text        string
	RichMessage InputRichMessage `json:"rich_message"`
}

type Chat struct {
	ID        int
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string
}

type Response struct {
	OK          bool
	Description string
	Result      any
}

type Sender interface {
	Send(ctx context.Context, data Request) (any, error)
}

type Bot struct {
	Dest      Sender
	ErrorChan chan<- error
	hs        []Handler
}

type Handler interface {
	Name() string
	Condition(Update) bool
	Handle(context.Context, Update) (msgs []Request, err error)
	HandleResponse(context.Context, any) error
}

func (b *Bot) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	srv := http.NewServeMux()

	slog.InfoContext(r.Context(), "request", "r", r)

	srv.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	srv.Handle("/webhook", http.HandlerFunc(b.webhook))

	srv.ServeHTTP(w, r)
}

func (b *Bot) Handle(h Handler) {
	b.hs = append(b.hs, h)
}

func (b *Bot) webhook(w http.ResponseWriter, r *http.Request) {
	upd, err := parseUpdate(w, r)
	if err != nil {
		b.error(err, w, http.StatusInternalServerError)
		return
	}

	var handler Handler
	for _, h := range b.hs {
		if !h.Condition(upd) {
			continue
		}

		handler = h
		break
	}

	msgs, err := handler.Handle(r.Context(), upd)
	if err != nil {
		b.error(fmt.Errorf("proc %q: %w", handler.Name(), err), w, http.StatusInternalServerError)
		return
	}

	for _, m := range msgs {
		ctx := r.Context()
		resp, err := b.Dest.Send(ctx, m)
		if err != nil {
			b.error(fmt.Errorf("dest send: %w", err), w, http.StatusInternalServerError)
			continue
		}

		err = handler.HandleResponse(ctx, resp)
		if err != nil {
			b.error(fmt.Errorf("dest resp save: %w", err), w, http.StatusInternalServerError)
		}
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
	if b.ErrorChan == nil {
		return
	}

	b.ErrorChan <- err
}

func adminMessages(ctx context.Context, store Adminser, blocks []InputRichBlock) (msgs []Request, err error) {
	msg := SendRichMessage{
		RichMessage: InputRichMessage{
			Blocks: []InputRichBlock{
				{
					Type: InputRichBlockParagraphType,
					Text: "Сообщение только для админов",
				},
			},
		},
	}
	msg.RichMessage.Blocks = append(msg.RichMessage.Blocks, blocks...)
	admins, err := store.Admins(ctx)
	if err != nil {
		return nil, fmt.Errorf("admins store: %w", err)
	}

	for _, a := range admins {
		msg.ChatID = a.UserID
		msgs = append(msgs, msg)
	}

	return
}
