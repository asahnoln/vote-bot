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
	answersCollection, err := f.CollectionByMode(ctx, AnswersCollection)
	if err != nil {
		return answers, fmt.Errorf("collection by mode: %w", err)
	}

	q := f.c.Collection(answersCollection).WhereEntity(firestore.PropertyFilter{
		Path:     "question_id",
		Operator: EqualOp,
		Value:    questionID,
	})

	err = f.c.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		iter := tx.Documents(q)

	loop:
		for {
			d, err := iter.Next()
			switch {
			case err == iterator.Done:
				break loop
			case err != nil:
				return fmt.Errorf("answers iter: %w", err)
			}

			err = tx.Update(d.Ref, []firestore.Update{
				{Path: "closed", Value: true},
			})
			if err != nil {
				return fmt.Errorf("doc %q update: %w", d.Ref.ID, err)
			}

			a := store.Answer{}
			err = d.DataTo(&a)
			if err != nil {
				return fmt.Errorf("data %#v to answer: %w", d, err)
			}

			answers = append(answers, a)
		}

		return nil
	})
	if err != nil {
		return answers, fmt.Errorf("transaction: %w", err)
	}

	return answers, nil
}
