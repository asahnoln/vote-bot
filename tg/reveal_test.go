package tg_test

import (
	"bytes"
	"context"
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
	b.Handle(tg.RevealHandler{Store: &revealStoreStub{}})
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

type revealStoreStub struct{}

func (s *revealStoreStub) LastClosedQuestion(context.Context) (store.Question, error) {
	return store.Question{
		Q:    "Да или Нет?",
		A:    0,
		Opts: []string{"Да", "Нет"},
	}, nil
}

func (s *revealStoreStub) Answer(context.Context, int, string) (store.Answer, error) {
	return store.Answer{
		IsRight: true,
	}, nil
}

func (s *revealStoreStub) Account(context.Context) (int, error) {
	return 500, nil
}
