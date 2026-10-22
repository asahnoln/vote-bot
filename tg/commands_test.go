package tg_test

import (
	"testing"

	"github.com/asahnoln/vote-bot/tg"
	"github.com/google/go-cmp/cmp"
)

func TestSetMyCommandsURL(t *testing.T) {
	msg := tg.SetMyCommands{}

	url := msg.URL()

	if got, want := url, "setMyCommands"; got != want {
		t.Errorf("url: got %q; want %q", got, want)
	}
}

func TestSetMyCommandsUnmarshal(t *testing.T) {
	m := tg.SetMyCommands{}

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

func TestCommandsForAdmins(t *testing.T) {
	msg := tg.CommandsForAdmin(100)

	want := tg.SetMyCommands{
		Commands: []tg.BotCommand{
			{
				Command:     "/sendcurrentquestion",
				Description: "Выслать текущий вопрос",
			},
			{
				Command:     "/closecurrentquestion",
				Description: "Закрыть текущий вопрос",
			},
			{
				Command:     "/sendanswer",
				Description: "Выслать всем ответ",
			},
			{
				Command:     "/statsrichestusers",
				Description: "Статистика: богачи",
			},
		},
		Scope: tg.BotCommandScope{
			Type:   "chat",
			ChatID: 100,
		},
	}
	if diff := cmp.Diff(want, msg); diff != "" {
		t.Errorf("commands message mismatch (-want +got):\n%s", diff)
	}
}
