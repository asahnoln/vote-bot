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
	h := tg.ModeHandler{Store: spy, Mode: store.OptionalMode}

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
				{Text: "Установлен режим вопросов: опциональный"},
			},
		},
	}
	if diff := cmp.Diff(want, msgs[0]); diff != "" {
		t.Errorf("send rich message mismatch (-want +got):\n%s", diff)
	}

	if got, want := msgs[1].(tg.SendRichMessage).ChatID, 10002; got != want {
		t.Errorf("2nd chat ID: got %v; want %v", got, want)
	}

	if got, want := spy.mode, store.OptionalMode; got != want {
		t.Errorf("mode set: got %v; want %v", got, want)
	}
}

func TestModeHadnlerAdminsError(t *testing.T) {
	wantErr := errors.New("admins error")
	spy := &modeStoreSpy{
		adminsErr: wantErr,
	}
	h := tg.ModeHandler{Store: spy, Mode: store.OptionalMode}

	_, err := h.Handle(context.Background(), tg.Update{})
	if got, want := err, wantErr; !errors.Is(got, want) {
		t.Errorf("got err %q; want %q", got, want)
	}
}

func TestModeHadnlerModeError(t *testing.T) {
	wantErr := errors.New("mode error")
	spy := &modeStoreSpy{
		modeErr: wantErr,
	}
	h := tg.ModeHandler{Store: spy, Mode: store.OptionalMode}

	_, err := h.Handle(context.Background(), tg.Update{})
	if got, want := err, wantErr; !errors.Is(got, want) {
		t.Errorf("got err %q; want %q", got, want)
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
