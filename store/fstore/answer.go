package fstore

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
)

func (f *Firestore) SaveAnswer(ctx context.Context, userID int, questionID string, answer string) error {
	q, err := f.CurrentQuestion(ctx)
	if err != nil {
		return fmt.Errorf("current question: %w", err)
	}

	_, err = f.c.Collection("answers").
		Doc(fmt.Sprintf("%s:%d", q.ID, userID)).
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

func (f *Firestore) SaveBet(context.Context, int, string, int) error {
	return nil
}
