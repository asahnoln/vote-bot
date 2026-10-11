package fstore

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/asahnoln/vote-bot/store"
)

const QuestionsCollection = "questions"

const (
	ClosedPath = "closed"
	OrderPath  = "order"
)

func (f *Firestore) CurrentQuestion(ctx context.Context) (q store.Question, err error) {
	docs, err := f.c.Collection(QuestionsCollection).
		WhereEntity(firestore.PropertyFilter{
			Path:     ClosedPath,
			Operator: EqualOp,
			Value:    false,
		}).
		OrderBy(OrderPath, firestore.Asc).
		Limit(1).
		Documents(ctx).
		GetAll()
	if err != nil {
		return q, fmt.Errorf("questions iterator: %w", err)
	}

	if len(docs) == 0 {
		return q, store.ErrNoQuestionsFound
	}

	err = docs[0].DataTo(&q)
	if err != nil {
		return q, fmt.Errorf("data to: %w", err)
	}

	q.ID = docs[0].Ref.ID

	return q, nil
}
