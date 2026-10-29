package fstore

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/asahnoln/vote-bot/store"
	"google.golang.org/api/iterator"
)

const AdminsCollection = "admins"

func (f *Firestore) Admins(ctx context.Context) (admins []store.User, err error) {
	docs, err := f.c.Collection(AdminsCollection).
		WhereEntity(firestore.PropertyFilter{Path: "active", Operator: EqualOp, Value: true}).
		Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("admins get all: %w", err)
	}

	for _, d := range docs {
		u := store.User{}
		err := d.DataTo(&u)
		if err != nil {
			return nil, fmt.Errorf("doc data to user: %w", err)
		}

		admins = append(admins, u)
	}

	return admins, nil
}

func (f *Firestore) SetMode(ctx context.Context, mode store.Mode) (err error) {
	_, err = f.c.Collection(ConfigCollection).Doc(ModeDocID).Set(ctx, map[string]any{
		ModePath: mode,
	})
	if err != nil {
		return fmt.Errorf("config mode set: %w", err)
	}

	_, err = f.c.Collection(ConfigCollection).Doc(CurrentQuestionDocID).Set(ctx, map[string]any{
		"idx": 0,
	})
	if err != nil {
		return fmt.Errorf("config current question set: %w", err)
	}

	err = f.restoreAllAccs(ctx)
	if err != nil {
		return err
	}

	return
}

func (f *Firestore) restoreAllAccs(ctx context.Context) (err error) {
	iter := f.c.Collection(UsersCollection).WhereEntity(firestore.PropertyFilter{Path: "acc", Operator: "!=", Value: store.AccountStartMoney}).Documents(ctx)
	if err != nil {
		return fmt.Errorf("all users: %w", err)
	}

	for {
		d, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("users iter: %w", err)
		}

		_, err = d.Ref.Update(ctx, []firestore.Update{
			{Path: "acc", Value: store.AccountStartMoney},
		})
		if err != nil {
			return fmt.Errorf("update %v: %w", d.Ref.ID, err)
		}
	}
	return nil
}

func (f *Firestore) CollectionByMode(ctx context.Context, col string) (string, error) {
	mode, err := f.Mode(ctx)
	if err != nil {
		return col, fmt.Errorf("mode: %w", err)
	}

	if mode == store.TutorialMode {
		col = string(mode) + "_" + col
	}
	return col, nil
}
