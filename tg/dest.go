package tg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

type URLer interface {
	URL() string
}
type SendRichMessage struct {
	ChatID      int              `json:"chat_id"`
	RichMessage InputRichMessage `json:"rich_message"`
}

func (SendRichMessage) URL() string {
	return "sendRichMessage"
}

type EditMessageText struct {
	ChatID      int              `json:"chat_id"`
	MessageID   int              `json:"message_id"`
	RichMessage InputRichMessage `json:"rich_message"`
}

func (EditMessageText) URL() string {
	return "editMessageText"
}

type DeleteMessage struct {
	ChatID    int `json:"chat_id"`
	MessageID int `json:"message_id"`
}

func (DeleteMessage) URL() string {
	return "deleteMessage"
}

type Dest struct {
	URL string
}

func NewDest(url string) Dest {
	return Dest{URL: url}
}

func (d Dest) Send(ctx context.Context, data URLer) (m any, err error) {
	if data == nil {
		return
	}

	methodURL := data.URL()

	slog.InfoContext(ctx, "message", "data", data)

	body, err := json.Marshal(&data)
	if err != nil {
		return m, fmt.Errorf("message marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx,
		http.MethodPost, d.URL+"/"+methodURL, bytes.NewReader(body))
	req.Header.Set("Content-Type", ApplicationJSONContentType)
	if err != nil {
		return m, fmt.Errorf("request creation: %w", err)
	}

	// TODO: Test for this err
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return m, fmt.Errorf("http client do: %w", err)
	}
	defer response.Body.Close()

	resp := Response{}
	err = json.NewDecoder(response.Body).Decode(&resp)
	if err != nil {
		return m, fmt.Errorf("tg srv response decode: %w", err)
	}

	if !resp.OK {
		return m, fmt.Errorf("tg srv not ok: %s", resp.Description)
	}

	if _, ok := resp.Result.(map[string]any); ok {
		data, err := json.Marshal(resp.Result)
		if err != nil {
			return m, fmt.Errorf("result marshal: %w", err)
		}

		msg := Message{}
		err = json.Unmarshal(data, &msg)
		if err != nil {
			return m, fmt.Errorf("result marshal: %w", err)
		}

		resp.Result = msg
	}

	return resp.Result, nil
}
