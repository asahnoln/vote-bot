package fstore

import (
	"context"

	"github.com/asahnoln/vote-bot/store"
)

func (f *Firestore) State(ctx context.Context) (store.State, error) {
	return store.StateQuestion, nil
}

func (f *Firestore) SaveState(ctx context.Context, state store.State) error {
	return nil
}
