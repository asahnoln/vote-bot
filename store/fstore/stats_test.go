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
		"user_id":          10,
		"lost":             false,
		"name":             "1st",
		"right_answer_cnt": 9,
		"acc":              1000,
		"avg_speed":        1,
	})
	col.Doc("11").Set(ctx, map[string]any{
		"user_id":          11,
		"lost":             false,
		"name":             "2nd",
		"right_answer_cnt": 9,
		"acc":              1000,
		"avg_speed":        2,
	})
	col.Doc("12").Set(ctx, map[string]any{
		"user_id":          12,
		"lost":             false,
		"name":             "3rd",
		"right_answer_cnt": 9,
		"acc":              999,
		"avg_speed":        0.5,
	})
	col.Doc("13").Set(ctx, map[string]any{
		"user_id":          13,
		"lost":             false,
		"name":             "4th",
		"right_answer_cnt": 9,
		"acc":              999,
		"avg_speed":        1,
	})
	col.Doc("14").Set(ctx, map[string]any{
		"user_id":          14,
		"lost":             false,
		"name":             "5th",
		"right_answer_cnt": 8,
		"acc":              1000,
		"avg_speed":        0.25,
	})
	col.Doc("15").Set(ctx, map[string]any{
		"user_id":          15,
		"lost":             true,
		"name":             "lost should be out",
		"right_answer_cnt": 100,
		"acc":              999999,
		"avg_speed":        0.1,
	})

	f := fstore.New(c)

	users, err := f.Stats(ctx)
	if err != nil {
		t.Fatalf("fstore stats fastest users: got err %v; want nil", err)
	}

	want := []store.User{
		{UserID: 10, Name: "1st", RightAnswerCnt: 9, Acc: 1000, AvgSpeed: 1},
		{UserID: 11, Name: "2nd", RightAnswerCnt: 9, Acc: 1000, AvgSpeed: 2},
		{UserID: 12, Name: "3rd", RightAnswerCnt: 9, Acc: 999, AvgSpeed: 0.5},
		{UserID: 13, Name: "4th", RightAnswerCnt: 9, Acc: 999, AvgSpeed: 1},
		{UserID: 13, Name: "5th", RightAnswerCnt: 8, Acc: 1000, AvgSpeed: 0.25},
	}
	if diff := cmp.Diff(want, users); diff != "" {
		t.Errorf("stats fastest users mismatch (-want +got):\n%s", diff)
	}
}
