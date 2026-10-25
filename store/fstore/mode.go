package fstore

import (
	"context"
	"fmt"

	"github.com/asahnoln/vote-bot/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const AdminsCollection = "admins"

func (f *Firestore) Admins(ctx context.Context) (admins []store.User, err error) {
	docs, err := f.c.Collection(AdminsCollection).
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
	_, err = f.c.Collection(ConfigCollection).Doc("mode").Set(ctx, map[string]any{
		"mode": mode,
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

	return
}

func (f *Firestore) CollectionByMode(ctx context.Context, col string) (string, error) {
	d, err := f.c.Collection(ConfigCollection).Doc("mode").Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return col, nil
		}

		return "", fmt.Errorf("config mode: %w", err)
	}

	if o, ok := d.Data()["mode"].(string); ok && o == string(store.TutorialMode) {
		col = o + "_" + col
	}
	return col, nil
}
