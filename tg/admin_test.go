package tg_test

import (
	"context"
	"errors"
	"testing"

	"github.com/asahnoln/vote-bot/tg"
)

func TestAdminWrapperHandle(t *testing.T) {
	spy := &adminsStoreStub{}

	h := &tg.AdminWrapper{
		Store:   spy,
		Handler: &handlerStub{},
	}

	msgs, err := h.Handle(context.Background(), tg.Update{
		Message: tg.Message{
			Chat: tg.Chat{
				ID: 10001,
			},
		},
	})
	if err != nil {
		t.Fatalf("got err %q; want nil", err)
	}

	if got, want := len(msgs), 1; got != want {
		t.Errorf("got %v; want %v", got, want)
	}
}

func TestAdminWrapperHandleNotAdmin(t *testing.T) {
	spy := &adminsStoreStub{}

	h := &tg.AdminWrapper{
		Store:   spy,
		Handler: &handlerStub{},
	}

	_, err := h.Handle(context.Background(), tg.Update{
		Message: tg.Message{
			Chat: tg.Chat{
				ID: 333,
			},
		},
	})

	if got, want := err, tg.ErrUserIsNotAdmin; !errors.Is(got, want) {
		t.Fatalf("got err %v; want %v", got, want)
	}
}

func TestAdminWrapperStoreError(t *testing.T) {
	wantErr := errors.New("store error")
	spy := &adminsStoreStub{
		err: wantErr,
	}

	h := &tg.AdminWrapper{
		Store:   spy,
		Handler: &handlerStub{},
	}

	_, err := h.Handle(context.Background(), tg.Update{})

	if got, want := err, wantErr; !errors.Is(got, want) {
		t.Fatalf("got err %v; want %v", got, want)
	}
}
