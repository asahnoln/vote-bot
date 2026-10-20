package fstore_test

import (
	"context"
	"testing"
	"time"

	"github.com/asahnoln/vote-bot/store"
	"github.com/asahnoln/vote-bot/store/fstore"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestSaveAnswer(t *testing.T) {
	ctx := context.Background()
	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	c.Collection("answers").Doc("52:test-q-67").Delete(ctx)
	c.Collection(fstore.QuestionsCollection).Doc("test-q-67").Set(ctx, store.Question{
		Q:     "52 or 42?",
		A:     []int{1, 2},
		Opts:  []string{"wrong", "right1", "right2"},
		Order: 0,
	})

	f := fstore.New(c)

	err := f.SaveAnswer(ctx, 52, "test-q-67", "right1")
	if err != nil {
		t.Fatalf("fstore save answer: got err %q; want nil", err)
	}

	d, err := c.Collection("answers").Doc("52:test-q-67").Get(ctx)
	if err != nil {
		t.Fatalf("get doc err: %v", err)
	}

	now := time.Now()
	want := map[string]any{
		"question_id": "test-q-67",
		"user_id":     int64(52),
		"a":           "right1",
		"is_right":    true,
		"answered_at": now,
	}
	if diff := cmp.Diff(want, d.Data(), cmpopts.EquateApproxTime(time.Second)); diff != "" {
		t.Errorf("answer mismatch (-want +got):\n%s", diff)
	}
}

func TestSaveAnswerNoQuestion(t *testing.T) {
	ctx := context.Background()
	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	f := fstore.New(c)

	err := f.SaveAnswer(ctx, 999, "test-q-9999", "")
	if err == nil {
		t.Fatal("fstore save answer: got err nil; want not nil")
	}
}

func TestSaveBet(t *testing.T) {
	ctx := context.Background()

	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	c.Collection("answers").Doc("234:test-q-111").Set(ctx, map[string]any{
		"a": "has answer",
	})

	f := fstore.New(c)

	err := f.SaveBet(ctx, 234, "test-q-111", 500)
	if err != nil {
		t.Fatalf("fstore save answer: got err %q; want nil", err)
	}

	d, err := c.Collection("answers").Doc("234:test-q-111").Get(ctx)
	if err != nil {
		t.Fatalf("get doc err: %v", err)
	}

	want := map[string]any{
		"a":   "has answer",
		"bet": int64(500),
	}
	if diff := cmp.Diff(want, d.Data(), cmpopts.EquateApproxTime(time.Second)); diff != "" {
		t.Errorf("bet mismatch (-want +got):\n%s", diff)
	}
}
