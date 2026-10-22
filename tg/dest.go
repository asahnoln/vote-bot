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
	UnmarshalResult(data []byte) (any, error)
}
type SendRichMessage struct {
	ChatID      int              `json:"chat_id"`
	RichMessage InputRichMessage `json:"rich_message"`
}

func (SendRichMessage) URL() string {
	return "sendRichMessage"
}

func (SendRichMessage) UnmarshalResult(data []byte) (any, error) {
	return baseUnmarshal[Message](data)
}

type EditMessageText struct {
	ChatID      int              `json:"chat_id"`
	MessageID   int              `json:"message_id"`
	RichMessage InputRichMessage `json:"rich_message"`
}

func (EditMessageText) URL() string {
	return "editMessageText"
}

func (EditMessageText) UnmarshalResult(data []byte) (any, error) {
	return baseUnmarshal[Message](data)
}

func baseUnmarshal[T any](data []byte) (result T, err error) {
	err = json.Unmarshal(data, &result)
	return
}

type DeleteMessage struct {
	ChatID    int `json:"chat_id"`
	MessageID int `json:"message_id"`
}

func (DeleteMessage) URL() string {
	return "deleteMessage"
}

func (DeleteMessage) UnmarshalResult(data []byte) (result any, err error) {
	err = json.Unmarshal(data, &result)
	return
}

type Dest struct {
	URL string
}

func NewDest(url string) Dest {
	return Dest{URL: url}
}

func (d Dest) Send(ctx context.Context, data URLer) (result any, err error) {
	if data == nil {
		return
	}

	methodURL := data.URL()

	slog.InfoContext(ctx, "message", "data", data)

	body, err := json.Marshal(&data)
	if err != nil {
		return result, fmt.Errorf("message marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx,
		http.MethodPost, d.URL+"/"+methodURL, bytes.NewReader(body))
	req.Header.Set("Content-Type", ApplicationJSONContentType)
	if err != nil {
		return result, fmt.Errorf("request creation: %w", err)
	}

	// TODO: Test for this err
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return result, fmt.Errorf("http client do: %w", err)
	}
	defer response.Body.Close()

	// TODO: Use whole response object for UnmarshalResult?
	resp := Response{}
	err = json.NewDecoder(response.Body).Decode(&resp)
	if err != nil {
		return result, fmt.Errorf("tg srv response decode: %w", err)
	}

	if !resp.OK {
		return result, fmt.Errorf("tg srv not ok: %s", resp.Description)
	}

	resultBytes, err := json.Marshal(resp.Result)
	if err != nil {
		return result, fmt.Errorf("result marshal: %w", err)
	}

	result, err = data.UnmarshalResult(resultBytes)
	if err != nil {
		return result, fmt.Errorf("data unmarshal result: %w", err)
	}

	return result, nil
}
