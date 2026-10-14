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
	b.Handle(&tg.RevealHandler{Store: &revealStoreStub{
		users: []store.User{
			{UserID: 876},
			{UserID: 2345},
		},
		answers: map[int]store.Answer{
			876: {IsRight: true},
		},
		accounts: map[int]int{
			876:  500,
			2345: 300,
		},
	}})
	b.ServeHTTP(w, r)

	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf("reveal status: got %v; want %v", got, want)
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
	h := tg.RevealHandler{Store: &revealStoreStub{
		users: []store.User{{}},
	}}
	data, err := h.Handle(context.Background(), tg.Update{})
	if err != nil {
		t.Fatalf("got err %v; want nil", err)
	}

	if got, want := data[0].(tg.SendRichMessage).RichMessage.Blocks[1].Text, "Вы ответили неправильно."; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestRevealLastUsersError(t *testing.T) {
	wantErr := errors.New("users error")
	h := tg.RevealHandler{Store: &revealStoreStub{
		usersErr: wantErr,
	}}

	_, err := h.Handle(context.Background(), tg.Update{})
	if got, want := err, wantErr; !errors.Is(got, want) {
		t.Errorf("got err %q; want %q", got, want)
	}
}

func TestRevealLastClosedQuestionError(t *testing.T) {
	wantErr := errors.New("last question error")
	h := tg.RevealHandler{Store: &revealStoreStub{
		lastClosedQuestionErr: wantErr,
	}}

	_, err := h.Handle(context.Background(), tg.Update{})
	if got, want := err, wantErr; !errors.Is(got, want) {
		t.Errorf("got err %q; want %q", got, want)
	}
}

func TestRevealAnswerError(t *testing.T) {
	wantErr := errors.New("answer error")
	h := tg.RevealHandler{Store: &revealStoreStub{
		users:     []store.User{{}},
		answerErr: wantErr,
	}}

	_, err := h.Handle(context.Background(), tg.Update{})
	if got, want := err, wantErr; !errors.Is(got, want) {
		t.Errorf("got err %q; want %q", got, want)
	}
}

func TestRevealAccountError(t *testing.T) {
	wantErr := errors.New("answer error")
	h := tg.RevealHandler{Store: &revealStoreStub{
		users:      []store.User{{}},
		accountErr: wantErr,
	}}

	_, err := h.Handle(context.Background(), tg.Update{})
	if got, want := err, wantErr; !errors.Is(got, want) {
		t.Errorf("got err %q; want %q", got, want)
	}
}

type revealStoreStub struct {
	users                 []store.User
	questionID            string
	answers               map[int]store.Answer
	accounts              map[int]int
	lastClosedQuestionErr error
	answerErr             error
	accountErr            error
	usersErr              error
}

func (s *revealStoreStub) Users(ctx context.Context) ([]store.User, error) {
	return s.users, s.usersErr
}

func (s *revealStoreStub) LastClosedQuestion(context.Context) (store.Question, error) {
	return store.Question{
		Q:    "Да или Нет?",
		A:    0,
		Opts: []string{"Да", "Нет"},
	}, s.lastClosedQuestionErr
}

func (s *revealStoreStub) Answer(ctx context.Context, userID int, questionID string) (store.Answer, error) {
	s.questionID = questionID
	return s.answers[userID], s.answerErr
}

func (s *revealStoreStub) Account(ctx context.Context, userID int) (int, error) {
	return s.accounts[userID], s.accountErr
}
