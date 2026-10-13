package tg_test

import (
	"context"
	"errors"
	"testing"

	"github.com/asahnoln/vote-bot/store"
	"github.com/asahnoln/vote-bot/tg"
	"github.com/google/go-cmp/cmp"
)

func TestCloseHandler(t *testing.T) {
	spy := &closeStoreSpy{
		userIDs: []store.User{
			{UserID: 91, LastMessageID: 67521},
			{UserID: 56, LastMessageID: 303},
			{UserID: 23, LastMessageID: 10002},
		},
	}
	h := tg.CloseHandler{
		Store: spy,
	}

	msgs, err := h.Handle(context.Background(), tg.Update{
		Message: tg.Message{
			Text: "/closequestion",
		},
	})
	if err != nil {
		t.Fatalf("got err %v; want nil", err)
	}

	if got, want := spy.closed, true; got != want {
		t.Errorf("question closed: got %v; want %v", got, want)
	}

	if got, want := len(msgs), 3; got != want {
		t.Errorf("msgs count: got %v; want %v", got, want)
	}

	want := tg.EditMessageText{
		ChatID:    91,
		MessageID: 67521,
		RichMessage: tg.InputRichMessage{
			Blocks: []tg.InputRichBlock{
				{
					Type: "paragraph",
					Text: "Вопрос закрыт!",
				},
			},
		},
	}
	if diff := cmp.Diff(want, msgs[0]); diff != "" {
		t.Errorf("request mismatch (-want +got):%s", diff)
	}

	if got, want := msgs[1].(tg.EditMessageText).ChatID, 56; got != want {
		t.Errorf("request 2nd user id: got %v; want %v", got, want)
	}
	if got, want := msgs[1].(tg.EditMessageText).MessageID, 303; got != want {
		t.Errorf("request 2nd user id: got %v; want %v", got, want)
	}
}

func TestCloseHandlerUsersStoreError(t *testing.T) {
	wantErr := errors.New("store error")
	spy := &closeStoreSpy{
		usersErr: wantErr,
	}
	h := tg.CloseHandler{
		Store: spy,
	}

	_, err := h.Handle(context.Background(), tg.Update{})
	if got, want := err, wantErr; !errors.Is(got, want) {
		t.Errorf("got err %v; want %v", got, want)
	}
}

func TestCloseHandlerQuestionStoreError(t *testing.T) {
	wantErr := errors.New("store error")
	spy := &closeStoreSpy{
		err: wantErr,
	}
	h := tg.CloseHandler{
		Store: spy,
	}

	_, err := h.Handle(context.Background(), tg.Update{})
	if got, want := err, wantErr; !errors.Is(got, want) {
		t.Errorf("got err %v; want %v", got, want)
	}
}

func TestCloseCondition(t *testing.T) {
	h := tg.CloseHandler{}

	ok := h.Condition(tg.Update{
		Message: tg.Message{
			Text: "/closequestion",
		},
	})

	if got, want := ok, true; got != want {
		t.Errorf("close condition: got %v; want %v", got, want)
	}
}

func TestCloseName(t *testing.T) {
	h := tg.CloseHandler{}

	n := h.Name()

	if got, want := n, "close"; got != want {
		t.Errorf("got name %q; want %q", got, want)
	}
}

type closeStoreSpy struct {
	userIDs  []store.User
	closed   bool
	err      error
	usersErr error
}

func (s *closeStoreSpy) Users(context.Context) ([]store.User, error) {
	return s.userIDs, s.usersErr
}

func (s *closeStoreSpy) CloseCurrentQuestion(ctx context.Context) error {
	s.closed = true
	return s.err
}
