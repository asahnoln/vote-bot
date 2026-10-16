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
	q, err = f.getQuestionByCurrentIdx(ctx, 0)
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

const (
	ConfigCollection     = "config"
	CurrentQuestionDocID = "current_question"
)

const IdxPath = "idx"

func (f *Firestore) CloseCurrentQuestion(ctx context.Context) error {
	results, err := f.c.Collection(QuestionsCollection).NewAggregationQuery().WithCount("all").Get(ctx)
	if err != nil {
		return fmt.Errorf("count all questions: %w", err)
	}

	count := results["all"].(*firestorepb.Value).GetIntegerValue()
	idx, err := f.getCurrentQuestionIdx(ctx)
	if err != nil {
		return fmt.Errorf("current idx: %w", err)
	}

	if idx >= int(count) {
		return store.ErrNoOpenQuestionsLeft
	}

	_, err = f.c.Collection(ConfigCollection).Doc(CurrentQuestionDocID).Update(ctx, []firestore.Update{
		{Path: IdxPath, Value: firestore.Increment(1)},
	})
	if err != nil {
		return fmt.Errorf("update idx: %w", err)
	}

	return nil
}

func (f *Firestore) getQuestionByCurrentIdx(ctx context.Context, n int) (q store.Question, err error) {
	idx, err := f.getCurrentQuestionIdx(ctx)
	if err != nil {
		return q, fmt.Errorf("current question idx: %w", err)
	}

	offset := int(idx) + n
	if offset < 0 {
		return q, fmt.Errorf("%w: idx %d, n %d", store.ErrQuestionsOffsetMustNotBeNegative, idx, n)
	}

	docs, err := f.c.Collection(QuestionsCollection).
		Offset(offset).
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

func (f *Firestore) getCurrentQuestionIdx(ctx context.Context) (int, error) {
	d, err := f.c.Collection(ConfigCollection).Doc(CurrentQuestionDocID).Get(ctx)
	if err != nil {
		return 0, fmt.Errorf("doc: %w", err)
	}

	idx, err := d.DataAt(IdxPath)
	if err != nil {
		return 0, fmt.Errorf("data: %w", err)
	}

	idxInt, ok := idx.(int64)
	if !ok {
		return 0, fmt.Errorf("idx %[1]v is not int64, but %[1]T", idx)
	}

	return int(idxInt), nil
}
