package fstore_test

import (
	"context"
	"testing"

	"github.com/asahnoln/vote-bot/store"
	"github.com/asahnoln/vote-bot/store/fstore"
	"github.com/google/go-cmp/cmp"
)

func TestStatsFastest(t *testing.T) {
	ctx := context.Background()

	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	col := c.Collection("users")
	col.Doc("10").Set(ctx, map[string]any{
		"user_id":   10,
		"lost":      false,
		"name":      "fastest won",
		"avg_speed": 0.5,
	})
	col.Doc("11").Set(ctx, map[string]any{
		"user_id":   11,
		"lost":      true,
		"name":      "mid fastest lost",
		"avg_speed": 1,
	})
	col.Doc("12").Set(ctx, map[string]any{
		"user_id":   12,
		"lost":      false,
		"name":      "slowest won",
		"avg_speed": 2,
	})
	col.Doc("13").Set(ctx, map[string]any{
		"user_id":   13,
		"lost":      false,
		"name":      "should not be in the stats",
		"avg_speed": 3,
	})

	f := fstore.New(c)

	users, err := f.StatsFastestUsers(ctx)
	if err != nil {
		t.Fatalf("fstore stats fastest users: got err %v; want nil", err)
	}

	want := []store.User{
		{UserID: 10, Name: "fastest won"},
		{UserID: 11, Name: "mid fastest lost", Lost: true},
		{UserID: 12, Name: "slowest won"},
	}
	if diff := cmp.Diff(want, users); diff != "" {
		t.Errorf("stats fastest users mismatch (-want +got):\n%s", diff)
	}
}
