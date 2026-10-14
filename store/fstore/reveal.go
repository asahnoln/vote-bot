package fstore

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/asahnoln/vote-bot/store"
)

func (f *Firestore) LastClosedQuestion(ctx context.Context) (q store.Question, err error) {
	q, err = f.getQuestion(ctx, true, firestore.Desc)
	if err != nil {
		return q, fmt.Errorf("get question: %w", err)
	}

	return q, nil
}

func (f *Firestore) Answer(ctx context.Context, userID int, questionID string) (a store.Answer, err error) {
	d, err := f.c.Collection(AnswersCollection).Doc(answerDocID(userID, questionID)).Get(ctx)
	if err != nil {
		return a, fmt.Errorf("answer doc: %w", err)
	}

	err = d.DataTo(&a)
	if err != nil {
		return a, fmt.Errorf("data to: %w", err)
	}

	return a, nil
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
