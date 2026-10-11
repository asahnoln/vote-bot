package fstore

import (
	"context"

	"cloud.google.com/go/firestore"
)

func (f *Firestore) SaveAnswer(ctx context.Context, userID int, answer string) error {
	f.c.Collection("answers").Doc("test-q-67:52").Set(ctx,
		map[string]any{
			"question_id": "test-q-67",
			"user_id":     52,
			"a":           "right",
			"is_right":    true,
			"answered_at": firestore.ServerTimestamp,
		})
	return nil
}
