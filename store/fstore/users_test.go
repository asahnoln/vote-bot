package fstore_test

import (
	"context"
	"errors"
	"testing"

	"github.com/asahnoln/vote-bot/store"
	"github.com/asahnoln/vote-bot/store/fstore"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAllUsers(t *testing.T) {
	ctx := context.Background()

	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	col := c.Collection("users")
	col.Doc("7829").Set(ctx, map[string]any{
		"user_id":         7829,
		"last_message_id": 101,
		"lost":            false,
	})
	col.Doc("404").Set(ctx, map[string]any{
		"user_id":         404,
		"last_message_id": 404,
		"lost":            true,
	})

	f := fstore.New(c)

	uIDs, err := f.AllUsers(ctx)
	if err != nil {
		t.Fatalf("fstore users: got err %v; want nil", err)
	}

	want := []store.User{
		{UserID: 404, LastMessageID: 404, Lost: true},
		{UserID: 7829, LastMessageID: 101},
	}
	if diff := cmp.Diff(want, uIDs); diff != "" {
		t.Errorf("all user IDs mismatch (-want +got):\n%s", diff)
	}
}

func TestUsers(t *testing.T) {
	ctx := context.Background()

	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	col := c.Collection("users")
	col.Doc("7829").Set(ctx, map[string]any{
		"user_id":         7829,
		"last_message_id": 101,
		"lost":            false,
	})
	col.Doc("390").Set(ctx, map[string]any{
		"user_id":         390,
		"last_message_id": 201,
		"lost":            false,
	})
	col.Doc("404").Set(ctx, map[string]any{
		"user_id":         404,
		"last_message_id": 404,
		"lost":            true,
	})

	f := fstore.New(c)

	uIDs, err := f.Users(ctx)
	if err != nil {
		t.Fatalf("fstore users: got err %v; want nil", err)
	}

	want := []store.User{
		{UserID: 390, LastMessageID: 201},
		{UserID: 7829, LastMessageID: 101},
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

	err := f.SaveUser(ctx, 67, "Cool name")
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
	if got, want := d.Data()["lost"].(bool), false; got != want {
		t.Errorf("got doc lost %v; want %v", got, want)
	}
	if got, want := d.Data()["name"].(string), "Cool name"; got != want {
		t.Errorf("got doc name %q; want %q", got, want)
	}
}

func TestSaveUserAlreadySavedError(t *testing.T) {
	ctx := context.Background()

	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	c.Collection("users").Doc("52").Set(ctx, map[string]any{
		"user_id": 52,
	})

	f := fstore.New(c)

	err := f.SaveUser(ctx, 52, "wow")
	if got, want := err, store.ErrUserAlreadySaved; !errors.Is(got, want) {
		t.Fatalf("fstore save user: got err %v; want %v", got, want)
	}
}

func TestSaveLastMessageID(t *testing.T) {
	ctx := context.Background()

	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	c.Collection("users").Doc("101101").Set(ctx, map[string]any{
		"user_id": 101101,
	})

	f := fstore.New(c)

	err := f.SaveLastMessageID(ctx, 101101, 55)
	if err != nil {
		t.Fatalf("fstore save user: got err %v; want nil", err)
	}

	d, err := c.Collection("users").Doc("101101").Get(ctx)
	if err != nil {
		t.Fatalf("fstore user get: got err %v; want nil", err)
	}

	got := store.User{}
	d.DataTo(&got)
	want := store.User{
		UserID:        101101,
		LastMessageID: 55,
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("user mismatch (-want +got):\n%s", diff)
	}
}

func TestSaveLostUser(t *testing.T) {
	ctx := context.Background()

	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	c.Collection("users").Doc("2222").Set(ctx, map[string]any{
		"user_id": 2222,
	})

	f := fstore.New(c)

	err := f.SaveLostUser(ctx, 2222)
	if err != nil {
		t.Fatalf("fstore save lost user: got err %v; want nil", err)
	}

	d, err := c.Collection("users").Doc("2222").Get(ctx)
	if err != nil {
		t.Fatalf("fstore user get: got err %v; want nil", err)
	}

	got := store.User{}
	d.DataTo(&got)
	want := store.User{
		UserID: 2222,
		Lost:   true,
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("lost user mismatch (-want +got):\n%s", diff)
	}
}

func TestSaveLostUserError(t *testing.T) {
	ctx := context.Background()

	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	f := fstore.New(c)

	err := f.SaveLostUser(ctx, 9999)
	if got, want := status.Code(err), codes.NotFound; err == nil || got != want {
		t.Fatalf("fstore save lost user: got err %v; want %v", got, want)
	}
}
