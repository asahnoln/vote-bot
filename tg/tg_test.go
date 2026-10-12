package tg_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/synctest"

	"github.com/asahnoln/vote-bot/store"
	"github.com/asahnoln/vote-bot/tg"
)

// TODO: Check if update is correct for every type of action, have a default action as well
func TestDefaultNotFound(t *testing.T) {
	r := httptest.NewRequest("", "/", nil)
	w := httptest.NewRecorder()

	b := tg.Bot{}
	b.ServeHTTP(w, r)

	if got, want := w.Code, http.StatusNotFound; got != want {
		t.Errorf("bot response code: got %v; want %v", got, want)
	}
}

func TestBotDestError(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := httptest.NewRequest("", "/webhook", strings.NewReader("{}"))
		w := httptest.NewRecorder()

		errCh := make(chan error)
		wantErr := errors.New("dest error")
		go func() {
			err := <-errCh
			if got, want := err, wantErr; !errors.Is(got, want) {
				t.Fatalf("got err %v; want %v", got, want)
			}
		}()

		b := tg.Bot{
			Dest:      &destSpy{err: wantErr},
			Store:     &storeSpy{},
			ErrorChan: errCh,
		}
		b.ServeHTTP(w, r)

		if got, want := w.Code, http.StatusInternalServerError; got != want {
			t.Errorf("bot response code: got %v; want %v", got, want)
		}
	})
}

type storeSpy struct {
	err        error
	userID     int
	questionID string
	answer     string
	bet        int
}

func (s *storeSpy) CurrentQuestion(context.Context) (store.Question, error) {
	return store.Question{
		ID:   "opt-q-99",
		Q:    "X or Y?",
		Opts: []string{"X", "Y"},
	}, s.err
}

func (s *storeSpy) SaveAnswer(ctx context.Context, userID int, questiongID string, answer string) error {
	s.userID = userID
	s.questionID = questiongID
	s.answer = answer
	return s.err
}

func (s *storeSpy) SaveBet(ctx context.Context, userID int, questionID string, bet int) error {
	s.userID = userID
	s.questionID = questionID
	s.bet = bet
	return s.err
}
