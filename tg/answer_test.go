package tg_test

import (
	"bytes"
	_ "embed"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/asahnoln/vote-bot/tg"
)

//go:embed testdata/answer_callback_update.json
var answerCallbackUpdate []byte

func TestAnswer(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(answerCallbackUpdate))
	w := httptest.NewRecorder()

	b := tg.Bot{}
	b.ServeHTTP(w, r)

	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf("answer status: got %v; want %v", got, want)
	}
}
