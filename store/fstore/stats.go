package fstore

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/asahnoln/vote-bot/store"
)

// TODO: Use this for stats handler
func (f *Firestore) StatsFastestUsers(ctx context.Context) (users []store.User, err error) {
	docs, err := f.c.Collection(UsersCollection).
		OrderBy("avg_speed", firestore.Asc).
		Limit(3).
		Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("get all: %w", err)
	}

	for _, d := range docs {
		u := store.User{}
		err := d.DataTo(&u)
		if err != nil {
			return nil, fmt.Errorf("data to: %w", err)
		}

		users = append(users, u)
	}

	return users, nil
}
