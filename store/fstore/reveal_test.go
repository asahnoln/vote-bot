package fstore_test

import (
	"context"
	"os"
	"testing"

	"cloud.google.com/go/firestore"
	"github.com/asahnoln/vote-bot/store"
	"github.com/asahnoln/vote-bot/store/fstore"
)

func TestAccount(t *testing.T) {
	os.Setenv("FIRESTORE_EMULATOR_HOST", "[::1]:8711")

	ctx := context.Background()

	c, err := firestore.NewClient(ctx, "test")
	if err != nil {
		t.Fatalf("app firestore: got err %v; want nil", err)
	}

	resetDB(ctx, c)

	col := c.Collection("answers")
	col.Doc("634:test-q-0").Set(ctx, store.Answer{
		IsRight: true,
		Bet:     200,
	})
	col.Doc("634:test-q-1").Set(ctx, store.Answer{
		IsRight: false,
		Bet:     100,
	})
	col.Doc("634:test-q-2").Set(ctx, store.Answer{
		IsRight: true,
		Bet:     500,
	})

	f := fstore.New(c)

	acc, err := f.Account(ctx, 634)
	if err != nil {
		t.Fatalf("fstore account: got err %v; want nil", err)
	}

	if got, want := acc, 1600; got != want {
		t.Errorf("got acc %v; want %v", got, want)
	}
}
