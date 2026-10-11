package tg_test

import (
	"bytes"
	_ "embed"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/asahnoln/vote-bot/tg"
	"github.com/google/go-cmp/cmp"
)

//go:embed testdata/answer_callback_update.json
var answerCallbackUpdate []byte

func TestAnswer(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(answerCallbackUpdate))
	w := httptest.NewRecorder()

	storeSpy := &storeSpy{}
	destSpy := &destSpy{}
	b := tg.Bot{
		Store: storeSpy,
		Dest:  destSpy,
	}
	b.ServeHTTP(w, r)

	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf("answer status: got %v; want %v", got, want)
	}

	want := tg.EditMessageText{
		ChatID:    45030020,
		MessageID: 79,
		RichMessage: tg.InputRichMessage{
			Blocks: []tg.InputRichBlock{
				{
					Type: "paragraph",
					Text: "Выберите ставку:",
				},
				{
					Type:  "buttons",
					Align: "left",
					Buttons: []tg.RichMessageButton{
						{
							Text:         "100",
							CallbackData: "100",
						},
						{
							Text:         "200",
							CallbackData: "200",
						},
						{
							Text:         "500",
							CallbackData: "500",
						},
					},
				},
			},
		},
	}
	if diff := cmp.Diff(want, destSpy.got); diff != "" {
		t.Errorf("edit message text mismatch (-want +got):%s", diff)
	}

	if got, want := storeSpy.userID, 45030020; got != want {
		t.Errorf("save answer: got %v; want %v", got, want)
	}
	if got, want := storeSpy.answer, "Ты"; got != want {
		t.Errorf("save answer: got %v; want %v", got, want)
	}
}

type destSpy struct {
	got any
}

func (d *destSpy) Send(data any) error {
	d.got = data
	return nil
}
