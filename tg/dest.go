package tg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

const (
	SendRichMessageMethod = "sendRichMessage"
	EditMessageTextMethod = "editMessageText"
)

type Dest struct {
	URL string
}

func NewDest(url string) Dest {
	return Dest{URL: url}
}

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

func (d Dest) Send(ctx context.Context, data URLer) error {
	if data == nil {
		return nil
	}

	url := data.URL()

	slog.InfoContext(ctx, "message", "data", data)

	body, err := json.Marshal(&data)
	if err != nil {
		return fmt.Errorf("message marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx,
		http.MethodPost, d.URL+"/"+url, bytes.NewReader(body))
	req.Header.Set("Content-Type", ApplicationJSONContentType)
	if err != nil {
		return fmt.Errorf("request creation: %w", err)
	}

	// TODO: Test for this err
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http client do: %w", err)
	}
	defer response.Body.Close()

	resp := Response{}
	err = json.NewDecoder(response.Body).Decode(&resp)
	if err != nil {
		return fmt.Errorf("tg srv response decode: %w", err)
	}

	if !resp.OK {
		return fmt.Errorf("tg srv not ok: %s", resp.Description)
	}

	return nil
}
