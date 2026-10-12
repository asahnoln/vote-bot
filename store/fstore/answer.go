package fstore

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/asahnoln/vote-bot/store"
)

const AnswersCollection = "answers"

func (f *Firestore) SaveAnswer(ctx context.Context, userID int, questionID string, answer string) error {
	// TODO: Case for not existing question
	q, err := f.question(ctx, questionID)
	if err != nil {
		return fmt.Errorf("question: %w", err)
	}

	_, err = f.c.Collection("answers").
		Doc(answerDocID(userID, questionID)).
		Set(ctx,
			map[string]any{
				"question_id": q.ID,
				"user_id":     userID,
				"a":           answer,
				"is_right":    q.Opts[q.A] == answer,
				"answered_at": firestore.ServerTimestamp,
			})
	if err != nil {
		return fmt.Errorf("doc set: %w", err)
	}
	return nil
}

func (f *Firestore) question(ctx context.Context, questionID string) (q store.Question, err error) {
	doc, err := f.c.Collection(QuestionsCollection).Doc(questionID).Get(ctx)
	if err != nil {
		return q, fmt.Errorf("get: %w", err)
	}

	err = doc.DataTo(&q)
	if err != nil {
		return q, fmt.Errorf("data to: %w", err)
	}

	q.ID = doc.Ref.ID

	return q, nil
}

func (f *Firestore) SaveBet(ctx context.Context, userID int, questionID string, bet int) error {
	_, err := f.c.Collection(AnswersCollection).
		Doc(answerDocID(userID, questionID)).
		Set(ctx,
			map[string]any{
				"bet": bet,
			}, firestore.MergeAll)
	if err != nil {
		return fmt.Errorf("doc set: %w", err)
	}
	return nil
}

func answerDocID(userID int, questionID string) string {
	return fmt.Sprintf("%d:%s", userID, questionID)
}
