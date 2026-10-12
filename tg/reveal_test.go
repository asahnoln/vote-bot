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
		Dest: destSpy,
	}
	b.Handle(tg.RevealHandler{Store: &revealStoreStub{
		isRight: true,
	}})
	b.ServeHTTP(w, r)

	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf("reveal status: got %v; want %v", got, want)
	}

	want := tg.SendRichMessage{
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
	if diff := cmp.Diff(want, destSpy.got); diff != "" {
		t.Errorf("edit message text mismatch (-want +got):%s", diff)
	}
}

func TestRevealWrongAnswer(t *testing.T) {
	h := tg.RevealHandler{Store: &revealStoreStub{
		isRight: false,
	}}
	data, err := h.Handle(context.Background(), tg.Update{})
	if err != nil {
		t.Fatalf("got err %v; want nil", err)
	}

	if got, want := data.(tg.SendRichMessage).RichMessage.Blocks[1].Text, "Вы ответили неправильно."; got != want {
		t.Errorf("got %q; want %q", got, want)
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
		accountErr: wantErr,
	}}

	_, err := h.Handle(context.Background(), tg.Update{})
	if got, want := err, wantErr; !errors.Is(got, want) {
		t.Errorf("got err %q; want %q", got, want)
	}
}

type revealStoreStub struct {
	userID                int
	isRight               bool
	lastClosedQuestionErr error
	answerErr             error
	accountErr            error
}

func (s *revealStoreStub) LastClosedQuestion(context.Context) (store.Question, error) {
	return store.Question{
		Q:    "Да или Нет?",
		A:    0,
		Opts: []string{"Да", "Нет"},
	}, s.lastClosedQuestionErr
}

func (s *revealStoreStub) Answer(context.Context, int, string) (store.Answer, error) {
	return store.Answer{
		IsRight: s.isRight,
	}, s.answerErr
}

func (s *revealStoreStub) Account(ctx context.Context, userID int) (int, error) {
	s.userID = userID
	return 500, s.accountErr
}
