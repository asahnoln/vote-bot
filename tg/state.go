package tg

import (
	"context"
	"errors"
	"fmt"

	"github.com/asahnoln/vote-bot/store"
)

var ErrWrongState = errors.New("wrong state")

type StateStore interface {
	State(context.Context) (store.State, error)
	SaveState(context.Context, store.State) error
}

type StateWrapper struct {
	Wrapper
	Store     StateStore
	State     store.State
	NextState store.State
}

func (w *StateWrapper) Handle(ctx context.Context, upd Update) (msgs []Request, err error) {
	state, err := w.Store.State(ctx)
	if err != nil {
		return nil, fmt.Errorf("state: %w", err)
	}

	if state != w.State {
		return nil, fmt.Errorf("%w: current state %q; will run at state %q", ErrWrongState, state, w.State)
	}

	err = w.Store.SaveState(ctx, w.NextState)
	if err != nil {
		return nil, fmt.Errorf("save state: %w", err)
	}

	return w.Handler.Handle(ctx, upd)
}
