package fstore_test

import (
	"context"
	"testing"

	"github.com/asahnoln/vote-bot/store"
	"github.com/asahnoln/vote-bot/store/fstore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestState(t *testing.T) {
	ctx := context.Background()

	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	col := c.Collection("config")
	col.Doc("state").Set(ctx, map[string]any{
		"state": store.StateCloseQuestion,
	})

	f := fstore.New(c)

	state, err := f.State(ctx)
	if err != nil {
		t.Fatalf("got err %v; want nil", err)
	}

	if got, want := state, store.StateCloseQuestion; got != want {
		t.Errorf("got state %v; want %v", got, want)
	}
}

func TestStateError(t *testing.T) {
	ctx := context.Background()

	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	f := fstore.New(c)

	_, err := f.State(ctx)
	if got, want := status.Code(err), codes.NotFound; err == nil || got != want {
		t.Fatalf("fstore state: got err %v; want %v", got, want)
	}
}

func TestStateWrongFormat(t *testing.T) {
	ctx := context.Background()

	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	col := c.Collection("config")
	col.Doc("state").Set(ctx, map[string]any{
		"state": 100,
	})

	f := fstore.New(c)

	_, err := f.State(ctx)
	if got, want := status.Code(err), codes.Unknown; err == nil || got != want {
		t.Fatalf("fstore state: got err %v; want %v", got, want)
	}
}

func TestSaveState(t *testing.T) {
	ctx := context.Background()

	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	col := c.Collection("config")
	col.Doc("state").Set(ctx, map[string]any{
		"state": store.StateCloseQuestion,
	})

	f := fstore.New(c)

	err := f.SaveState(ctx, store.StateRevealAnswer)
	if err != nil {
		t.Fatalf("got err %v; want nil", err)
	}

	d, _ := c.Collection("config").Doc("state").Get(ctx)
	if got, want := d.Data()["state"], string(store.StateRevealAnswer); got != want {
		t.Errorf("state: got %v; want %v", got, want)
	}
}
