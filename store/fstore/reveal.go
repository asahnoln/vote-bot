package fstore

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/asahnoln/vote-bot/store"
	"google.golang.org/api/iterator"
)

func (f *Firestore) LastClosedQuestion(ctx context.Context) (q store.Question, err error) {
	q, err = f.getQuestion(ctx, true, firestore.Desc)
	if err != nil {
		return q, fmt.Errorf("get question: %w", err)
	}

	return q, nil
}

func (f *Firestore) Answers(ctx context.Context, questionID string) (answers []store.Answer, err error) {
	iter := f.c.Collection(AnswersCollection).WhereEntity(firestore.PropertyFilter{
		Path:     "question_id",
		Operator: EqualOp,
		Value:    questionID,
	}).Documents(ctx)

loop:
	for {
		d, err := iter.Next()
		switch {
		case err == iterator.Done:
			break loop
		case err != nil:
			return nil, fmt.Errorf("answers iter: %w", err)
		}

		a := store.Answer{}
		err = d.DataTo(&a)
		if err != nil {
			return nil, fmt.Errorf("data %#v to answer: %w", d, err)
		}

		answers = append(answers, a)
	}

	return answers, nil
}

func (f *Firestore) Accounts(ctx context.Context, userIDs []int) (accs map[int]int, err error) {
	return nil, nil
}

const ACCOUNT_START_MONEY = 1000

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
		a := store.Answer{}
		err := d.DataTo(&a)
		if err != nil {
			return 0, fmt.Errorf("data %#v to answer: %w", d, err)
		}

		if !a.IsRight {
			a.Bet *= -1
		}
		sum += int(a.Bet)
	}

	return ACCOUNT_START_MONEY + sum, nil
}
