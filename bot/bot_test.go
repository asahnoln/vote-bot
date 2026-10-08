package bot_test

import (
	"errors"
	"testing"

	"github.com/asahnoln/vote-bot/bot"
	"github.com/google/go-cmp/cmp"
)

func TestQuestion(t *testing.T) {
	spy := &spyQuestionOutput{}
	b := &bot.Bot{
		QuestionStore: &stubQuestionStore{
			q:    "A or B?",
			opts: []string{"A", "B"},
		},
		QuestionOutput: spy,
	}

	err := b.CurrentQuestion()
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
	b := &bot.Bot{
		QuestionStore: &stubQuestionStore{
			err: errors.New("store error"),
		},
	}

	err := b.CurrentQuestion()

	if err == nil {
		t.Error("got err nil; want not nil")
	}
}

func TestQuestionOutputErr(t *testing.T) {
	b := &bot.Bot{
		QuestionStore: &stubQuestionStore{},
		QuestionOutput: &spyQuestionOutput{
			err: errors.New("output error"),
		},
	}

	err := b.CurrentQuestion()

	if err == nil {
		t.Error("got err nil; want not nil")
	}
}

type stubQuestionStore struct {
	q    string
	opts []string
	err  error
}

func (q *stubQuestionStore) CurrentQuestion() (string, []string, error) {
	return q.q, q.opts, q.err
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
