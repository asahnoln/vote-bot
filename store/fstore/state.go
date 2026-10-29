package fstore

import (
	"context"

	"github.com/asahnoln/vote-bot/store"
)

func (f *Firestore) State(ctx context.Context) (store.State, error) {
	return store.StateQuestion, nil
}
