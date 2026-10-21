package fstore

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/asahnoln/vote-bot/store"
	"google.golang.org/api/iterator"
)

func (f *Firestore) LastClosedQuestion(ctx context.Context) (q store.Question, err error) {
	q, err = f.getQuestionByCurrentIdx(ctx, -1)
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

const ACCOUNT_START_MONEY = 1000

func (f *Firestore) Accounts(ctx context.Context, userIDs []int) (accs map[int]int, err error) {
	iter := f.c.Collection(AnswersCollection).WhereEntity(firestore.PropertyFilter{
		Path:     "user_id",
		Operator: "in",
		Value:    userIDs,
	}).Documents(ctx)

	accs = make(map[int]int)

	for _, u := range userIDs {
		accs[u] = ACCOUNT_START_MONEY
	}

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

		if !a.IsRight {
			a.Bet *= -1
		}

		accs[a.UserID] += int(a.Bet)
	}

	return accs, nil
}
