package fstore

import (
	"context"
	"fmt"

	"github.com/asahnoln/vote-bot/store"
)

const StateDocID = "state"

func (f *Firestore) State(ctx context.Context) (store.State, error) {
	d, err := f.c.Collection(ConfigCollection).Doc(StateDocID).Get(ctx)
	if err != nil {
		return "", fmt.Errorf("get: %w", err)
	}

	var r struct {
		State store.State `firestore:"state"`
	}
	err = d.DataTo(&r)
	if err != nil {
		return "", fmt.Errorf("data to: %w", err)
	}

	return r.State, nil
}

func (f *Firestore) SaveState(ctx context.Context, state store.State) error {
	_, err := f.c.Collection(ConfigCollection).Doc(StateDocID).Set(ctx, map[string]any{
		"state": state,
	})
	if err != nil {
		return fmt.Errorf("set: %w", err)
	}

	return nil
}
