package tg_test

import (
	"context"
	"testing"

	"github.com/asahnoln/vote-bot/tg"
	"github.com/google/go-cmp/cmp"
)

func TestDefaultHandler(t *testing.T) {
	h := tg.DefaultHandler{}

	msgs, err := h.Handle(context.Background(), tg.Update{
		Message: tg.Message{
			Chat: tg.Chat{
				ID: 567,
			},
			MessageID: 89,
		},
	})
	if err != nil {
		t.Fatalf("got err %v; want nil", err)
	}

	want := tg.DeleteMessage{
		ChatID:    567,
		MessageID: 89,
	}
	if diff := cmp.Diff(want, msgs[0]); diff != "" {
		t.Errorf("msg mismatch (-want +got):\n%s", diff)
	}
}

func TestDefaultCondition(t *testing.T) {
	h := tg.DefaultHandler{}

	ok := h.Condition(tg.Update{})

	if got, want := ok, true; got != want {
		t.Errorf("default condition: got %v; want %v", got, want)
	}
}

func TestDefaultName(t *testing.T) {
	h := tg.DefaultHandler{}

	n := h.Name()

	if got, want := n, "default"; got != want {
		t.Errorf("got name %q; want %q", got, want)
	}
}
