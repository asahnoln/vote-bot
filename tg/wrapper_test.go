package tg_test

import (
	"context"
	"testing"

	"github.com/asahnoln/vote-bot/tg"
)

func TestWrapperUsesInternalHandlerByDefault(t *testing.T) {
	w := tg.Wrapper{
		Handler: &handlerStub{},
	}

	msgs, err := w.Handle(context.Background(), tg.Update{})
	if err != nil {
		t.Fatalf("got err %q; want nil", err)
	}

	if got, want := len(msgs), 1; got != want {
		t.Errorf("got %v; want %v", got, want)
	}

	if got, want := w.Condition(tg.Update{}), true; got != want {
		t.Errorf("condition: got %v; want %v", got, want)
	}

	if got, want := w.Name(), "stub"; got != want {
		t.Errorf("name: got %v; want %v", got, want)
	}

	err = w.HandleResponse(context.Background(), nil)
	if err != nil {
		t.Fatalf("got err %q; want nil", err)
	}
}
