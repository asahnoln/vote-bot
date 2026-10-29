package tg

import (
	"context"
	"errors"

	"github.com/asahnoln/vote-bot/store"
)

var ErrWrongState = errors.New("state must be question")

type Stater interface {
	State(context.Context) (store.State, error)
}

type StateWrapper struct {
	Store   Stater
	State   store.State
	Handler Handler
}

func (w *StateWrapper) Handle(ctx context.Context, upd Update) (msgs []Request, err error) {
	state, err := w.Store.State(ctx)
	if err != nil {
		return nil, err
	}
	if state != w.State {
		return nil, ErrWrongState
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
