package tg_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "embed"

	"github.com/asahnoln/vote-bot/store"
	"github.com/asahnoln/vote-bot/tg"
	"github.com/google/go-cmp/cmp"
)

//go:embed testdata/reveal_command_update.json
var revealCommandUpdate []byte

func TestReveal(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(revealCommandUpdate))
	w := httptest.NewRecorder()

	destSpy := &destSpy{}
	b := tg.Bot{
		Store: &storeSpy{},
		Dest:  destSpy,
	}

	storeSpy := &revealStoreSpy{
		answers: []store.Answer{
			{
				UserID:  876,
				IsRight: true,
			},
			{
				UserID:  2345,
				IsRight: false,
			},
		},
		accounts: map[int]int{
			876:  500,
			2345: 300,
		},
	}
	b.Handle(&tg.RevealHandler{Store: storeSpy})
	b.ServeHTTP(w, r)

	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf("reveal status: got %v; want %v", got, want)
	}

	if got, want := storeSpy.questionID, "last-closed-q-99"; got != want {
		t.Errorf("question id for answer: got %q; want %q", got, want)
	}

	wantUIDs := []int{876, 2345}
	if diff := cmp.Diff(wantUIDs, storeSpy.userIDs); diff != "" {
		t.Errorf("accounts user IDs mismatch (-want +got):%s", diff)
	}

	if got, want := len(destSpy.got), 2; got != want {
		t.Fatalf("request count: got %v; want %v", got, want)
	}

	want1 := tg.SendRichMessage{
		ChatID: 876,
		RichMessage: tg.InputRichMessage{
			Blocks: []tg.InputRichBlock{
				{
					Type: "paragraph",
					Text: "Да или Нет? Правильный ответ: Да!",
				},
				{
					Type: "paragraph",
					Text: "Вы ответили правильно!",
				},
				{
					Type: "paragraph",
					Text: "Ваш текущий счет - 500.",
				},
			},
		},
	}

	if diff := cmp.Diff(want1, destSpy.got[0]); diff != "" {
		t.Errorf("send message 1 mismatch (-want +got):%s", diff)
	}

	want2 := tg.SendRichMessage{
		ChatID: 2345,
		RichMessage: tg.InputRichMessage{
			Blocks: []tg.InputRichBlock{
				{
					Type: "paragraph",
					Text: "Да или Нет? Правильный ответ: Да!",
				},
				{
					Type: "paragraph",
					Text: "Вы ответили неправильно.",
				},
				{
					Type: "paragraph",
					Text: "Ваш текущий счет - 300.",
				},
			},
		},
	}
	if diff := cmp.Diff(want2, destSpy.got[1]); diff != "" {
		t.Errorf("send message 2 mismatch (-want +got):%s", diff)
	}
}

func TestRevealWrongAnswer(t *testing.T) {
	h := tg.RevealHandler{Store: &revealStoreSpy{
		answers: []store.Answer{{}},
	}}
	data, err := h.Handle(context.Background(), tg.Update{})
	if err != nil {
		t.Fatalf("got err %v; want nil", err)
	}

	if got, want := data[0].(tg.SendRichMessage).RichMessage.Blocks[1].Text, "Вы ответили неправильно."; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestRevealLastClosedQuestionError(t *testing.T) {
	wantErr := errors.New("last question error")
	h := tg.RevealHandler{Store: &revealStoreSpy{
		lastClosedQuestionErr: wantErr,
	}}

	_, err := h.Handle(context.Background(), tg.Update{})
	if got, want := err, wantErr; !errors.Is(got, want) {
		t.Errorf("got err %q; want %q", got, want)
	}
}

func TestRevealAnswersError(t *testing.T) {
	wantErr := errors.New("answers error")
	h := tg.RevealHandler{Store: &revealStoreSpy{
		answersErr: wantErr,
	}}

	_, err := h.Handle(context.Background(), tg.Update{})
	if got, want := err, wantErr; !errors.Is(got, want) {
		t.Errorf("got err %q; want %q", got, want)
	}
}

func TestRevealAccountsError(t *testing.T) {
	wantErr := errors.New("accounts error")
	h := tg.RevealHandler{Store: &revealStoreSpy{
		accountsErr: wantErr,
	}}

	_, err := h.Handle(context.Background(), tg.Update{})
	if got, want := err, wantErr; !errors.Is(got, want) {
		t.Errorf("got err %q; want %q", got, want)
	}
}

type revealStoreSpy struct {
	questionID            string
	answers               []store.Answer
	accounts              map[int]int
	lastClosedQuestionErr error
	answersErr            error
	accountsErr           error
	userIDs               []int
}

func (s *revealStoreSpy) LastClosedQuestion(context.Context) (store.Question, error) {
	return store.Question{
		ID:   "last-closed-q-99",
		Q:    "Да или Нет?",
		A:    0,
		Opts: []string{"Да", "Нет"},
	}, s.lastClosedQuestionErr
}

func (s *revealStoreSpy) Answers(ctx context.Context, questionID string) ([]store.Answer, error) {
	s.questionID = questionID
	return s.answers, s.answersErr
}

func (s *revealStoreSpy) Accounts(ctx context.Context, userIDs []int) (map[int]int, error) {
	s.userIDs = userIDs
	return s.accounts, s.accountsErr
}
