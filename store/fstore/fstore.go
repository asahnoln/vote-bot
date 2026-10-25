package fstore

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/asahnoln/vote-bot/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Firestore struct {
	c *firestore.Client
}

const EqualOp = "=="

const (
	ConfigCollection     = "config"
	CurrentQuestionDocID = "current_question"
	ModeDocID            = "mode"
)

const (
	IdxPath  = "idx"
	ModePath = "mode"
)

func New(c *firestore.Client) *Firestore {
	return &Firestore{c}
}

func (f *Firestore) mode(ctx context.Context) (store.Mode, error) {
	d, err := f.c.Collection(ConfigCollection).Doc(ModeDocID).Get(ctx)
	if err != nil {
		if status.Code(err) != codes.NotFound {
			return "", fmt.Errorf("get: %w", err)
		}
	}

	data := struct {
		Mode store.Mode
	}{}
	err = d.DataTo(&data)
	if err != nil {
		if status.Code(err) != codes.NotFound {
			return "", fmt.Errorf("data to mode: %w", err)
		}
	}

	return data.Mode, nil
}
