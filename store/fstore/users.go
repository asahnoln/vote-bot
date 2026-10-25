package fstore

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"cloud.google.com/go/firestore"
	"github.com/asahnoln/vote-bot/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const UsersCollection = "users"

func (f *Firestore) AllUsers(ctx context.Context) (users []store.User, err error) {
	return f.users(ctx)
}

func (f *Firestore) Users(ctx context.Context) (users []store.User, err error) {
	return f.users(ctx, firestore.PropertyFilter{
		Path:     "lost",
		Operator: EqualOp,
		Value:    false,
	})
}

func (f *Firestore) users(ctx context.Context, efs ...firestore.EntityFilter) (users []store.User, err error) {
	col := f.c.Collection(UsersCollection)
	q := col.Query
	for _, ef := range efs {
		q = col.WhereEntity(ef)
	}

	docs, err := q.
		Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("users get all: %w", err)
	}

	for _, d := range docs {
		u := store.User{}
		err := d.DataTo(&u)
		if err != nil {
			return nil, fmt.Errorf("doc data to user: %w", err)
		}

		users = append(users, u)
	}

	return users, nil
}

const AccountStartMoney = 1000

func (f *Firestore) SaveUser(ctx context.Context, userID int, name string) error {
	_, err := f.c.Collection(UsersCollection).
		Doc(strconv.Itoa(userID)).
		Create(ctx, store.User{
			UserID: userID,
			Name:   name,
			Acc:    AccountStartMoney,
		})
	if err != nil {
		if status.Code(err) == codes.AlreadyExists {
			err = errors.Join(store.ErrUserAlreadySaved, err)
		}

		return fmt.Errorf("user %d doc: %w", userID, err)
	}

	return nil
}

func (f *Firestore) SaveLastMessageID(ctx context.Context, userID int, msgID int) error {
	_, err := f.c.Collection(UsersCollection).
		Doc(strconv.Itoa(userID)).
		Set(ctx, map[string]any{
			"last_message_id": msgID,
		}, firestore.MergeAll)
	if err != nil {
		return fmt.Errorf("user %d doc with msgID %d: %w", userID, msgID, err)
	}

	return nil
}

func (f *Firestore) SaveLostUser(ctx context.Context, userID int) error {
	_, err := f.c.Collection(UsersCollection).
		Doc(strconv.Itoa(userID)).
		Update(ctx, []firestore.Update{
			{Path: "lost", Value: true},
		})
	if err != nil {
		return fmt.Errorf("users update: %w", err)
	}

	return nil
}

func (f *Firestore) RestoreAcc(ctx context.Context, userID int) error {
	_, err := f.c.Collection(UsersCollection).Doc(strconv.Itoa(userID)).Set(ctx, map[string]any{
		"acc": AccountStartMoney,
	}, firestore.MergeAll)
	if err != nil {
		return fmt.Errorf("user set acc: %w", err)
	}

	return nil
}
