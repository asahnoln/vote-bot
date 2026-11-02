package tg_test

import (
	"context"
	"errors"
	"testing"

	"github.com/asahnoln/vote-bot/store"
	"github.com/asahnoln/vote-bot/tg"
)

func TestStateWrapperChecksState(t *testing.T) {
	stub := &stateStoreSpy{
		state: store.StateRevealAnswer,
	}

	h := &tg.StateWrapper{
		Store:   stub,
		State:   store.StateSendQuestion,
		Handler: &handlerStub{},
	}
	_, err := h.Handle(context.Background(), tg.Update{})
	if got, want := err, tg.ErrWrongState; !errors.Is(got, want) {
		t.Fatalf("got err %q; want %q", got, want)
	}
}

func TestStateWrapperError(t *testing.T) {
	wantErr := errors.New("state error")
	stub := &stateStoreSpy{
		err: wantErr,
	}
	h := tg.StateWrapper{
		Store: stub,
	}
	_, err := h.Handle(context.Background(), tg.Update{})

	if got, want := err, wantErr; !errors.Is(got, want) {
		t.Fatalf("got err %v; want %v", got, want)
	}
}

func TestStateWrapperSavesNextState(t *testing.T) {
	spy := &stateStoreSpy{
		state: store.StateSendQuestion,
	}

	h := &tg.StateWrapper{
		Store:     spy,
		State:     store.StateSendQuestion,
		NextState: store.StateCloseQuestion,
		Handler:   &handlerStub{},
	}
	msgs, err := h.Handle(context.Background(), tg.Update{})
	if err != nil {
		t.Fatalf("got err %q; want nil", err)
	}

	if got, want := spy.state, store.StateCloseQuestion; got != want {
		t.Errorf("got state %v; want %v", got, want)
	}

	if got, want := len(msgs), 1; got != want {
		t.Fatalf("msgs count: got %d; want %d", got, want)
	}
}

func TestStateWrapperSaveError(t *testing.T) {
	wantErr := errors.New("save state error")
	stub := &stateStoreSpy{
		saveErr: wantErr,
	}
	h := tg.StateWrapper{
		Store: stub,
	}

	_, err := h.Handle(context.Background(), tg.Update{})
	if got, want := err, wantErr; !errors.Is(got, want) {
		t.Fatalf("got err %v; want %v", got, want)
	}
}

type stateStoreSpy struct {
	state   store.State
	err     error
	saveErr error
}

func (s *stateStoreSpy) State(context.Context) (store.State, error) {
	return s.state, s.err
}

func (s *stateStoreSpy) SaveState(ctx context.Context, state store.State) error {
	s.state = state
	return s.saveErr
}
