package tg_test

import (
	"context"
	"testing"

	"github.com/asahnoln/vote-bot/tg"
)

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

	// TODO: Which msg to tg I need?
	//
	if got, want := spy.userID, 913; got != want {
		t.Errorf("user ID save: got %d; want %d", got, want)
	}
}

type usersStoreSpy struct{}
