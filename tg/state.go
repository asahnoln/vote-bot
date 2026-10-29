package tg

import (
	"context"
	"errors"
	"fmt"

	"github.com/asahnoln/vote-bot/store"
)

var ErrWrongState = errors.New("state must be question")

type StateStore interface {
	State(context.Context) (store.State, error)
	SaveState(context.Context, store.State) error
}

type StateWrapper struct {
	Store     StateStore
	State     store.State
	NextState store.State
	Handler   Handler
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

func (w *StateWrapper) Condition(upd Update) bool {
	return w.Handler.Condition(upd)
}

func (w *StateWrapper) HandleResponse(ctx context.Context, msg any) error {
	return w.Handler.HandleResponse(ctx, msg)
}

func (w *StateWrapper) Name() string {
	return w.Handler.Name()
}
