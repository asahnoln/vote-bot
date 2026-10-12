package fstore

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/asahnoln/vote-bot/store"
)

func (f *Firestore) LastClosedQuestion(context.Context) (store.Question, error) {
	return store.Question{}, nil
}

func (f *Firestore) Answer(context.Context, int, string) (store.Answer, error) {
	return store.Answer{}, nil
}

func (f *Firestore) Account(ctx context.Context, userID int) (int, error) {
	docs, err := f.c.Collection(AnswersCollection).WhereEntity(firestore.PropertyFilter{
		Path:     "user_id",
		Operator: EqualOp,
		Value:    userID,
	}).Documents(ctx).GetAll()
	if err != nil {
		return 0, fmt.Errorf("answer docs: %w", err)
	}

	sum := 0
	for _, d := range docs {
		data := d.Data()
		b := data["bet"].(int64)
		if !data["is_right"].(bool) {
			b *= -1
		}
		sum += int(b)
	}

	return 1000 + sum, nil
}
