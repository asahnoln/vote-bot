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
	q, err = f.currentQuestion(ctx)
	if err != nil {
		return q, fmt.Errorf("current question: %w", err)
	}

	if len(q.Opts) == 0 {
		return q, store.ErrQuestionHasNoOpts
	}

	if q.A < 0 || q.A >= len(q.Opts) {
		return q, store.ErrQuestionAnswerIndexIsOutOfBounds
	}

	return q, nil
}

func (f *Firestore) CloseCurrentQuestion(ctx context.Context) error {
	q, err := f.currentQuestion(ctx)
	if err != nil {
		return fmt.Errorf("current question: %w", err)
	}

	_, err = f.c.Collection(QuestionsCollection).
		Doc(q.ID).Set(ctx, map[string]any{
		"closed": true,
	}, firestore.MergeAll)
	if err != nil {
		return fmt.Errorf("current question set: %w", err)
	}

	return nil
}

func (f *Firestore) getQuestion(ctx context.Context, closed bool, orderDir firestore.Direction) (q store.Question, err error) {
	docs, err := f.c.Collection(QuestionsCollection).
		WhereEntity(firestore.PropertyFilter{
			Path:     ClosedPath,
			Operator: EqualOp,
			Value:    closed,
		}).
		OrderBy(OrderPath, orderDir).
		Limit(1).
		Documents(ctx).
		GetAll()
	if err != nil {
		return q, fmt.Errorf("get: %w", err)
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

func (f *Firestore) currentQuestion(ctx context.Context) (q store.Question, err error) {
	return f.getQuestion(ctx, false, firestore.Asc)
}
