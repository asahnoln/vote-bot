package tg_test

import (
	"testing"

	"github.com/asahnoln/vote-bot/tg"
)

func TestDeleteMessageURL(t *testing.T) {
	m := tg.DeleteMessage{}

	u := m.URL()

	if got, want := u, "deleteMessage"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestEditMessageTextURL(t *testing.T) {
	m := tg.EditMessageText{}

	u := m.URL()

	if got, want := u, "editMessageText"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSednRichMessageURL(t *testing.T) {
	m := tg.SendRichMessage{}

	u := m.URL()

	if got, want := u, "sendRichMessage"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}
