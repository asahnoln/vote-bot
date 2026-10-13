package fstore_test

import (
	"context"
	"testing"

	"github.com/asahnoln/vote-bot/store"
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
		t.Fatalf("fstore users: got err %v; want nil", err)
	}

	want := []store.User{
		{UserID: 390},
		{UserID: 7829},
	}
	if diff := cmp.Diff(want, uIDs); diff != "" {
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
		t.Fatalf("fstore users: got err %v; want nil", err)
	}

	if got, want := len(uIDs), 0; got != want {
		t.Errorf("user count: got %d; want %d", got, want)
	}
}

func TestSaveUser(t *testing.T) {
	ctx := context.Background()

	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	f := fstore.New(c)

	err := f.SaveUser(ctx, 67)
	if err != nil {
		t.Fatalf("fstore save user: got err %v; want nil", err)
	}

	d, err := c.Collection("users").Doc("67").Get(ctx)
	if err != nil {
		t.Fatalf("fstore user get: got err %v; want nil", err)
	}

	if got, want := d.Data()["user_id"].(int64), int64(67); got != want {
		t.Errorf("got doc user id %v; want %v", got, want)
	}
}
