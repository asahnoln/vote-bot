package fstore

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/asahnoln/vote-bot/store"
)

type Firestore struct {
	c *firestore.Client
}

func New(c *firestore.Client) *Firestore {
	return &Firestore{c}
}

const QuestionCollection = "questions"

const (
	ClosedPath = "closed"
	OrderPath  = "order"
)

const EqualOp = "=="

func (f *Firestore) CurrentQuestion(ctx context.Context) (q store.Question, err error) {
	docs, err := f.c.Collection(QuestionCollection).
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

	return q, nil
}
