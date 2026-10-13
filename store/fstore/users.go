package fstore

import (
	"context"
	"fmt"
	"strconv"
)

const UsersCollection = "users"

func (f *Firestore) Users(ctx context.Context) (userIDs []int, err error) {
	docs, err := f.c.Collection(UsersCollection).Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("users get all: %w", err)
	}

	// TODO: Handle errors from type casting
	for _, d := range docs {
		userIDs = append(userIDs, int(d.Data()["user_id"].(int64)))
	}

	return userIDs, nil
}

func (f *Firestore) SaveUser(ctx context.Context, userID int) error {
	_, err := f.c.Collection(UsersCollection).
		Doc(strconv.Itoa(userID)).
		Set(ctx, map[string]any{
			"user_id": userID,
		})
	if err != nil {
		return fmt.Errorf("user %d doc: %w", userID, err)
	}

	return nil
}
