package tg_test

import (
	"context"
	"errors"
	"testing"

	"github.com/asahnoln/vote-bot/store"
	"github.com/asahnoln/vote-bot/tg"
	"github.com/google/go-cmp/cmp"
)

func TestModeOptionalHandler(t *testing.T) {
	spy := &modeStoreSpy{}
	h := tg.ModeHandler{Store: spy, Mode: store.TutorialMode}

	msgs, err := h.Handle(context.Background(), tg.Update{})
	if err != nil {
		t.Fatalf("got err %v; want nil", err)
	}

	if got, want := len(msgs), 2; got != want {
		t.Fatalf("msgs count: got %v; want %v", got, want)
	}

	want := tg.SendRichMessage{
		ChatID: 10001,
		RichMessage: tg.InputRichMessage{
			Blocks: []tg.InputRichBlock{
				{
					Type: "paragraph",
					Text: "Установлен режим вопросов: учебный",
				},
			},
		},
	}
	if diff := cmp.Diff(want, msgs[0]); diff != "" {
		t.Errorf("send rich message mismatch (-want +got):\n%s", diff)
	}

	if got, want := msgs[1].(tg.SendRichMessage).ChatID, 10002; got != want {
		t.Errorf("2nd chat ID: got %v; want %v", got, want)
	}

	if got, want := spy.mode, store.TutorialMode; got != want {
		t.Errorf("mode set: got %v; want %v", got, want)
	}
}

func TestModeHandlerAdminsError(t *testing.T) {
	wantErr := errors.New("admins error")
	spy := &modeStoreSpy{
		adminsErr: wantErr,
	}
	h := tg.ModeHandler{Store: spy, Mode: store.TutorialMode}

	_, err := h.Handle(context.Background(), tg.Update{})
	if got, want := err, wantErr; !errors.Is(got, want) {
		t.Errorf("got err %q; want %q", got, want)
	}
}

func TestModeHandlerModeError(t *testing.T) {
	wantErr := errors.New("mode error")
	spy := &modeStoreSpy{
		modeErr: wantErr,
	}
	h := tg.ModeHandler{Store: spy, Mode: store.TutorialMode}

	_, err := h.Handle(context.Background(), tg.Update{})
	if got, want := err, wantErr; !errors.Is(got, want) {
		t.Errorf("got err %q; want %q", got, want)
	}
}

func TestModeHandlerCondition(t *testing.T) {
	h := tg.ModeHandler{Mode: store.TutorialMode}
	ok := h.Condition(tg.Update{
		Message: tg.Message{
			Text: "/settutorialmode",
		},
	})

	if got, want := ok, true; got != want {
		t.Errorf("got %v; want %v", got, want)
	}

	ok = h.Condition(tg.Update{})
	if got, want := ok, false; got != want {
		t.Errorf("got %v; want %v", got, want)
	}
}

func TestModeHandlerName(t *testing.T) {
	h := tg.ModeHandler{}
	n := h.Name()
	if got, want := n, "mode"; got != want {
		t.Errorf("got %v; want %v", got, want)
	}
}

type modeStoreSpy struct {
	mode      store.Mode
	adminsErr error
	modeErr   error
}

func (s *modeStoreSpy) Admins(ctx context.Context) ([]store.User, error) {
	return []store.User{
		{UserID: 10001},
		{UserID: 10002},
	}, s.adminsErr
}

func (s *modeStoreSpy) SetMode(ctx context.Context, mode store.Mode) error {
	s.mode = mode
	return s.modeErr
}
