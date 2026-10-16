package fstore

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/firestore/apiv1/firestorepb"
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
	results, _ := f.c.Collection(QuestionsCollection).NewAggregationQuery().WithCount("all").Get(ctx)
	count := results["all"].(*firestorepb.Value).GetIntegerValue()
	d, _ := f.c.Collection("config").Doc("current_question").Get(ctx)
	if d.Data()["idx"].(int64) >= count {
		return store.ErrNoOpenQuestionsLeft
	}

	f.c.Collection("config").Doc("current_question").Update(ctx, []firestore.Update{
		{Path: "idx", Value: firestore.Increment(1)},
	})

	return nil
}

func (f *Firestore) getQuestionByCurrentIdx(ctx context.Context, n int) (q store.Question, err error) {
	d, _ := f.c.Collection("config").Doc("current_question").Get(ctx)

	docs, err := f.c.Collection(QuestionsCollection).
		Offset(int(d.Data()["idx"].(int64)) + n).
		Limit(1).
		Documents(ctx).
		GetAll()
	if err != nil {
		return q, fmt.Errorf("get: %w", err)
	}

	if len(docs) == 0 {
		return q, store.ErrNoOpenQuestionsLeft
	}

	err = docs[0].DataTo(&q)
	if err != nil {
		return q, fmt.Errorf("data to: %w", err)
	}

	q.ID = docs[0].Ref.ID

	return q, nil
}

func (f *Firestore) currentQuestion(ctx context.Context) (q store.Question, err error) {
	return f.getQuestionByCurrentIdx(ctx, 0)
}
