package tg

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	upd := Update{}
	err := json.NewDecoder(r.Body).Decode(&upd)
	if err != nil {
		b.error(fmt.Errorf("update decode: %w", err), w, http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	slog.InfoContext(r.Context(), "update", "upd", upd)

	if b.Store == nil {
		b.error(ErrStoreNil, w, http.StatusInternalServerError)
		return
	}

	q, err := b.Store.CurrentQuestion(r.Context())
	if err != nil {
		b.error(fmt.Errorf("question store: %w", err), w, http.StatusInternalServerError)
		return
	}

	msg := SendRichMessage{
		ChatID: upd.Message.Chat.ID,
		RichMessage: InputRichMessage{
			Blocks: []InputRichBlock{
				{
					Type: InputRichBlockParagraphType,
					Text: q.Q,
				},
				{
					Type:  InputRichBlockButtonsType,
					Align: InputRichBlockButtonsAlignCenter,
				},
			},
		},
	}
	for _, o := range q.Opts {
		// TODO: TDD utility for button auto row placement
		msg.RichMessage.Blocks[1].Buttons = append(
			msg.RichMessage.Blocks[1].Buttons,
			RichMessageButton{Text: o, CallbackData: o},
		)
	}

	slog.InfoContext(r.Context(), "message", "msg", msg)

	body, err := json.Marshal(&msg)
	if err != nil {
		b.error(fmt.Errorf("message marshal: %w", err), w, http.StatusInternalServerError)
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, b.URL+"/"+SendRichMessageMethod, bytes.NewReader(body))
	req.Header.Set("Content-Type", ApplicationJSONContentType)
	if err != nil {
		b.error(fmt.Errorf("request creation: %w", err), w, http.StatusInternalServerError)
		return
	}

	// TODO: Test for this err
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		b.error(fmt.Errorf("http client do: %w", err), w, http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	resp := Response{}
	err = json.NewDecoder(response.Body).Decode(&resp)
	if err != nil {
		b.error(fmt.Errorf("tg srv response decode: %w", err), w, http.StatusInternalServerError)
		return
	}

	if !resp.OK {
		b.error(fmt.Errorf("tg srv not ok: %s", resp.Description), w, http.StatusServiceUnavailable)
		return
	}
}

func (b *Bot) error(err error, w http.ResponseWriter, s int) {
	b.ErrorChan <- err
	w.WriteHeader(s)
}
