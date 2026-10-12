package fstore

import (
	"context"
	"fmt"
)

const UsersCollection = "users"

func (f *Firestore) Users(ctx context.Context) (userIDs []int, err error) {
	docs, err := f.c.Collection(UsersCollection).Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("users get all: %w", err)
	}

	for _, d := range docs {
		userIDs = append(userIDs, int(d.Data()["user_id"].(int64)))
	}

	return userIDs, nil
}
