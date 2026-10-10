package bot_test

import (
	"errors"
	"testing"

	"github.com/asahnoln/vote-bot/bot"
	"github.com/asahnoln/vote-bot/store"
	"github.com/google/go-cmp/cmp"
)

func TestQuestion(t *testing.T) {
	spy := &spyQuestionOutput{}
	b := &bot.Question{
		Store: &stubQuestionStore{
			q:    "A or B?",
			opts: []string{"A", "B"},
		},
		Output: spy,
	}

	err := b.Current()
	if err != nil {
		t.Fatalf("got err %v; want nil", err)
	}

	if got, want := spy.q, "A or B?"; got != want {
		t.Errorf("question: got %q; want %q", got, want)
	}

	if diff := cmp.Diff([]string{"A", "B"}, spy.opts); diff != "" {
		t.Errorf("options mismatch (-want +got):\n%s", diff)
	}
}

func TestQuestionStoreErr(t *testing.T) {
	b := &bot.Question{
		Store: &stubQuestionStore{
			err: errors.New("store error"),
		},
	}

	err := b.Current()

	if err == nil {
		t.Error("got err nil; want not nil")
	}
}

func TestQuestionOutputErr(t *testing.T) {
	b := &bot.Question{
		Store: &stubQuestionStore{},
		Output: &spyQuestionOutput{
			err: errors.New("output error"),
		},
	}

	err := b.Current()

	if err == nil {
		t.Error("got err nil; want not nil")
	}
}

func TestAnswer(t *testing.T) {
	spy := &spyAnswerStore{}
	b := &bot.Answer{
		Store: spy,
	}

	err := b.Save("userID", "ANSWER!!!", 200)
	if err != nil {
		t.Fatalf("got err %v; want nil", err)
	}

	if got, want := spy.a, "ANSWER!!!"; got != want {
		t.Errorf("answer: got %q; want %q", got, want)
	}

	if got, want := spy.bet, 200; got != want {
		t.Errorf("bet: got %v; want %v", got, want)
	}

	if got, want := spy.uID, "userID"; got != want {
		t.Errorf("user: got %v; want %v", got, want)
	}
}

func TestAnswerErr(t *testing.T) {
	spy := &spyAnswerStore{
		err: errors.New("store error"),
	}
	b := &bot.Answer{
		Store: spy,
	}

	err := b.Save("", "", 0)
	if err == nil {
		t.Error("got err nil; want not nil")
	}
}

type stubQuestionStore struct {
	q    string
	opts []string
	err  error
}

func (q *stubQuestionStore) CurrentQuestion() (store.Question, error) {
	return store.Question{Q: q.q, Opts: q.opts}, q.err
}

type spyQuestionOutput struct {
	q    string
	opts []string
	err  error
}

func (o *spyQuestionOutput) OutputQuestion(q string, opts []string) error {
	o.q = q
	o.opts = opts
	return o.err
}

type spyAnswerStore struct {
	a   string
	bet int
	uID string
	err error
}

func (s *spyAnswerStore) SaveAnswer(uID string, a string, bet int) error {
	s.a = a
	s.bet = bet
	s.uID = uID
	return s.err
}
