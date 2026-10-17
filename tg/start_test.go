package tg_test

import (
	"context"
	"errors"
	"testing"

	"github.com/asahnoln/vote-bot/tg"
	"github.com/google/go-cmp/cmp"
)

// TODO: Send him some acknowledge that he already joined

func TestStartHandler(t *testing.T) {
	spy := &usersStoreSpy{}
	h := tg.StartHandler{
		Store: spy,
	}

	msgs, err := h.Handle(context.Background(), tg.Update{
		Message: tg.Message{
			Chat: tg.Chat{
				ID: 913,
			},
			Text: "/start",
		},
	})
	if err != nil {
		t.Fatalf("got err %v; want nil", err)
	}

	if got, want := spy.userID, 913; got != want {
		t.Errorf("user ID save: got %d; want %d", got, want)
	}

	want := tg.SendRichMessage{
		ChatID: 913,
		RichMessage: tg.InputRichMessage{
			Blocks: []tg.InputRichBlock{
				{
					Type: "paragraph",
					// TODO: Какое сообщение?
					Text: "Добро пожаловать в игру!",
				},
			},
		},
	}
	if diff := cmp.Diff(want, msgs[0]); diff != "" {
		t.Errorf("msg mismatch (-want +got):%s", diff)
	}
}

func TestStartHandlerStoreError(t *testing.T) {
	wantErr := errors.New("store error")
	spy := &usersStoreSpy{
		err: wantErr,
	}
	h := tg.StartHandler{
		Store: spy,
	}

	_, err := h.Handle(context.Background(), tg.Update{})
	if got, want := err, wantErr; !errors.Is(got, want) {
		t.Errorf("got err %v; want %v", got, want)
	}
}

func TestStartDontSaveIfQuestionMissed(t *testing.T) {
	spy := &usersStoreSpy{
		firstQuestionMissed: true,
	}
	h := tg.StartHandler{
		Store: spy,
	}

	msgs, err := h.Handle(context.Background(), tg.Update{
		Message: tg.Message{
			Chat: tg.Chat{
				ID: 111,
			},
			Text: "/start",
		},
	})
	if err != nil {
		t.Fatalf("got err %v; want nil", err)
	}

	if got, want := spy.userID, 0; got != want {
		t.Errorf("user ID save: got %d; want %d", got, want)
	}

	want := tg.SendRichMessage{
		ChatID: 111,
		RichMessage: tg.InputRichMessage{
			Blocks: []tg.InputRichBlock{
				{
					Type: "paragraph",
					// TODO: Какое сообщение?
					Text: "Вы пропустили первый обязательный вопрос и потому не можете быть допущены к игре.",
				},
			},
		},
	}
	if diff := cmp.Diff(want, msgs[0]); diff != "" {
		t.Errorf("msg mismatch (-want +got):%s", diff)
	}
}

func TestStartHandlerQuestionStoreError(t *testing.T) {
	wantErr := errors.New("question store error")
	spy := &usersStoreSpy{
		questionErr: wantErr,
	}
	h := tg.StartHandler{
		Store: spy,
	}

	_, err := h.Handle(context.Background(), tg.Update{})
	if got, want := err, wantErr; !errors.Is(got, want) {
		t.Errorf("got err %v; want %v", got, want)
	}
}

func TestStartCondition(t *testing.T) {
	h := tg.StartHandler{}

	ok := h.Condition(tg.Update{
		Message: tg.Message{
			Text: "/start",
		},
	})

	if got, want := ok, true; got != want {
		t.Errorf("start condition: got %v; want %v", got, want)
	}
}

func TestStartName(t *testing.T) {
	h := tg.StartHandler{}

	n := h.Name()

	if got, want := n, "start"; got != want {
		t.Errorf("got name %q; want %q", got, want)
	}
}

type usersStoreSpy struct {
	userID              int
	firstQuestionMissed bool
	err                 error
	questionErr         error
}

func (s *usersStoreSpy) SaveUser(ctx context.Context, userID int) error {
	s.userID = userID
	return s.err
}

func (s *usersStoreSpy) FirstQuestionMissed(ctx context.Context) (bool, error) {
	return s.firstQuestionMissed, s.questionErr
}
