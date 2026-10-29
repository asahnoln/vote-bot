package fstore_test

import (
	"context"
	"testing"

	"github.com/asahnoln/vote-bot/store"
	"github.com/asahnoln/vote-bot/store/fstore"
	"github.com/google/go-cmp/cmp"
)

func TestAdmins(t *testing.T) {
	ctx := context.Background()

	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	col := c.Collection("admins")
	col.Doc("101").Set(ctx, map[string]any{
		"user_id": 101,
		"active":  true,
	})
	col.Doc("102").Set(ctx, map[string]any{
		"user_id": 102,
		"active":  true,
	})
	col.Doc("103").Set(ctx, map[string]any{
		"user_id": 103,
		"active":  false,
	})

	f := fstore.New(c)

	admins, err := f.Admins(ctx)
	if err != nil {
		t.Fatalf("fstore admins: got err %v; want nil", err)
	}

	want := []store.User{
		{UserID: 101},
		{UserID: 102},
	}
	if diff := cmp.Diff(want, admins); diff != "" {
		t.Errorf("admins mismatch (-want +got):\n%s", diff)
	}
}

func TestSetMode(t *testing.T) {
	ctx := context.Background()

	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	col := c.Collection("config")
	col.Doc("current_question").Set(ctx, map[string]any{
		"idx": 99,
	})

	col = c.Collection("users")
	col.Doc("1").Set(ctx, map[string]any{
		"user_id": 1,
		"acc":     0,
	})
	col.Doc("2").Set(ctx, map[string]any{
		"user_id": 2,
		"acc":     10,
	})

	f := fstore.New(c)

	err := f.SetMode(ctx, store.TutorialMode)
	if err != nil {
		t.Fatalf("fstore admins: got err %v; want nil", err)
	}

	d, _ := c.Collection("config").Doc("mode").Get(ctx)
	if got, want := d.Data()["mode"], string(store.TutorialMode); got != want {
		t.Errorf("mode: got %v; want %v", got, want)
	}

	d, _ = c.Collection("config").Doc("current_question").Get(ctx)
	if got, want := d.Data()["idx"], int64(0); got != want {
		t.Errorf("current question idx: got %v; want %v", got, want)
	}

	d, _ = c.Collection("config").Doc("state").Get(ctx)
	if got, want := d.Data()["state"], string(store.StateQuestion); got != want {
		t.Errorf("current state: got %v; want %v", got, want)
	}

	d, _ = c.Collection("users").Doc("1").Get(ctx)
	if got, want := d.Data()["acc"], int64(1000); got != want {
		t.Errorf("user 1 acc: got %v; want %v", got, want)
	}

	d, _ = c.Collection("users").Doc("2").Get(ctx)
	if got, want := d.Data()["acc"], int64(1000); got != want {
		t.Errorf("user 2 acc: got %v; want %v", got, want)
	}
}

func TestCollectionByModeNoErrorWhenNotFound(t *testing.T) {
	ctx := context.Background()

	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	f := fstore.New(c)

	col, err := f.CollectionByMode(ctx, "someCollection")
	if err != nil {
		t.Fatalf("fstore CollectionByMode: got err %v; want nil", err)
	}

	if got, want := col, "someCollection"; got != want {
		t.Errorf("got %v; want %v", got, want)
	}
}
