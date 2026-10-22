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

func TestDeleteMessageUnmarshal(t *testing.T) {
	m := tg.DeleteMessage{}

	r, err := m.UnmarshalResult([]byte("true"))
	if err != nil {
		t.Fatalf("got err %v; want nil", err)
	}

	got, ok := r.(bool)
	if got, want := ok, true; got != want {
		t.Fatalf("proper type: got %v; want %v", got, want)
	}
	if got, want := got, true; got != want {
		t.Errorf("got %v; want %v", got, want)
	}
}

func TestEditMessageTextURL(t *testing.T) {
	m := tg.EditMessageText{}

	u := m.URL()

	if got, want := u, "editMessageText"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestEditMessgeTextUnmarshal(t *testing.T) {
	m := tg.EditMessageText{}

	r, err := m.UnmarshalResult([]byte(`{"message_id":777}`))
	if err != nil {
		t.Fatalf("got err %v; want nil", err)
	}

	got, ok := r.(tg.Message)
	if got, want := ok, true; got != want {
		t.Fatalf("proper type: got %v; want %v", got, want)
	}
	if got, want := got.MessageID, 777; got != want {
		t.Errorf("got %v; want %v", got, want)
	}
}

// TODO: Remove dependency from other dest tests on it
func TestSendRichMessageURL(t *testing.T) {
	m := tg.SendRichMessage{}

	u := m.URL()

	if got, want := u, "sendRichMessage"; got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestSendRichMessageUnmarshal(t *testing.T) {
	m := tg.SendRichMessage{}

	r, err := m.UnmarshalResult([]byte(`{"message_id":778}`))
	if err != nil {
		t.Fatalf("got err %v; want nil", err)
	}

	got, ok := r.(tg.Message)
	if got, want := ok, true; got != want {
		t.Fatalf("proper type: got %v; want %v", got, want)
	}
	if got, want := got.MessageID, 778; got != want {
		t.Errorf("got %v; want %v", got, want)
	}
}
