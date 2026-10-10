package tg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

func (b *Bot) question(w http.ResponseWriter, r *http.Request) {
	upd := Update{}
	updBody, err := io.ReadAll(r.Body)
	// TODO: Test for this error
	if err != nil {
		b.error(fmt.Errorf("update read: %w", err), w, http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	slog.InfoContext(r.Context(), "update", "upd", updBody)

	err = json.Unmarshal(updBody, &upd)
	if err != nil {
		b.error(fmt.Errorf("update decode: %w", err), w, http.StatusInternalServerError)
		return
	}

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
