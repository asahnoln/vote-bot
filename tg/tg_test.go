package tg_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/asahnoln/vote-bot/store"
	"github.com/asahnoln/vote-bot/tg"
)

func TestDefaultNotFound(t *testing.T) {
	r := httptest.NewRequest("", "/", nil)
	w := httptest.NewRecorder()

	b := tg.Bot{}
	b.ServeHTTP(w, r)

	if got, want := w.Code, http.StatusNotFound; got != want {
		t.Errorf("bot response code: got %v; want %v", got, want)
	}
}

type storeSpy struct {
	err    error
	answer string
	userID int
}

func (s *storeSpy) CurrentQuestion(context.Context) (store.Question, error) {
	return store.Question{
		Q:    "X or Y?",
		Opts: []string{"X", "Y"},
	}, s.err
}

func (s *storeSpy) SaveAnswer(ctx context.Context, userID int, answer string) error {
	s.userID = userID
	s.answer = answer
	return s.err
}
