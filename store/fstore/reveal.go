package fstore

import (
	"context"

	"github.com/asahnoln/vote-bot/store"
)

func (f *Firestore) LastClosedQuestion(context.Context) (store.Question, error) {
	return store.Question{}, nil
}

func (f *Firestore) Answer(context.Context, int, string) (store.Answer, error) {
	return store.Answer{}, nil
}

func (f *Firestore) Account(ctx context.Context, userID int) (int, error) {
	return 0, nil
}
