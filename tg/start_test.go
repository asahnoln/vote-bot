package tg_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/asahnoln/vote-bot/store"
	"github.com/asahnoln/vote-bot/tg"
	"github.com/google/go-cmp/cmp"
)

func TestStartHandler(t *testing.T) {
	spy := &usersStoreSpy{}
	h := tg.StartHandler{
		Store: spy,
		Text:  &textStub{},
	}

	upd := tg.Update{}
	json.Unmarshal(revealCommandUpdate, &upd)
	msgs, err := h.Handle(context.Background(), upd)
	if err != nil {
		t.Fatalf("got err %v; want nil", err)
	}

	if got, want := spy.userID, 876; got != want {
		t.Errorf("user ID save: got %d; want %d", got, want)
	}
	if got, want := spy.name, "Артур Асланян @asahnoln"; got != want {
		t.Errorf("user name save: got %q; want %q", got, want)
	}

	want := tg.SendRichMessage{
		ChatID: 876,
		RichMessage: tg.InputRichMessage{
			Blocks: []tg.InputRichBlock{
				{
					Type: "paragraph",
					Text: "TextWelcomeToTheGame",
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
		Text:  &textStub{},
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
		Text:  &textStub{},
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
					Text: "TextMissedFirstQuestion",
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

func TestStartDontSaveIfUserAlreadyJoined(t *testing.T) {
	spy := &usersStoreSpy{
		err: store.ErrUserAlreadySaved,
	}
	h := tg.StartHandler{
		Store: spy,
		Text:  &textStub{},
	}

	msgs, err := h.Handle(context.Background(), tg.Update{
		Message: tg.Message{
			Chat: tg.Chat{
				ID: 202,
			},
			Text: "/start",
		},
	})
	if err != nil {
		t.Fatalf("got err %v; want nil", err)
	}

	// TODO: What text should be used?
	if got, want := msgs[0].(tg.SendRichMessage).RichMessage.Blocks[0].Text, "TextAlreadyJoinedTheGame"; got != want {
		t.Errorf("text: got %v; want %v", got, want)
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
	name                string
	firstQuestionMissed bool
	err                 error
	questionErr         error
}

func (s *usersStoreSpy) SaveUser(ctx context.Context, userID int, name string) error {
	s.userID = userID
	s.name = name
	return s.err
}

func (s *usersStoreSpy) FirstQuestionMissed(ctx context.Context) (bool, error) {
	return s.firstQuestionMissed, s.questionErr
}
