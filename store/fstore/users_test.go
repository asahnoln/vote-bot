package fstore_test

import (
	"context"
	"testing"

	"github.com/asahnoln/vote-bot/store/fstore"
	"github.com/google/go-cmp/cmp"
)

func TestUsers(t *testing.T) {
	ctx := context.Background()

	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	col := c.Collection("users")
	col.Doc("7829").Set(ctx, map[string]any{
		"user_id": 7829,
	})
	col.Doc("390").Set(ctx, map[string]any{
		"user_id": 390,
	})

	f := fstore.New(c)

	uIDs, err := f.Users(ctx)
	if err != nil {
		t.Fatalf("fstore question: got err %v; want nil", err)
	}

	if diff := cmp.Diff([]int{390, 7829}, uIDs); diff != "" {
		t.Errorf("user IDs mismatch (-want +got):\n%s", diff)
	}
}

func TestUsersNoData(t *testing.T) {
	ctx := context.Background()

	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	f := fstore.New(c)

	uIDs, err := f.Users(ctx)
	if err != nil {
		t.Fatalf("fstore question: got err %v; want nil", err)
	}

	if got, want := len(uIDs), 0; got != want {
		t.Errorf("user count: got %d; want %d", got, want)
	}
}
