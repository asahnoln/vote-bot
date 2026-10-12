package fstore_test

import (
	"context"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/asahnoln/vote-bot/store"
	"github.com/asahnoln/vote-bot/store/fstore"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestSaveAnswer(t *testing.T) {
	os.Setenv("FIRESTORE_EMULATOR_HOST", "[::1]:8711")

	ctx := context.Background()

	c, err := firestore.NewClient(ctx, "test")
	if err != nil {
		t.Fatalf("app firestore: got err %v; want nil", err)
	}

	resetDB(ctx, c)

	c.Collection("answers").Doc("52:test-q-67").Delete(ctx)
	c.Collection(fstore.QuestionsCollection).Doc("test-q-67").Set(ctx, store.Question{
		Q:      "52 or 42?",
		Closed: false,
		A:      1,
		Opts:   []string{"wrong", "right"},
		Order:  0,
	})

	f := fstore.New(c)

	err = f.SaveAnswer(ctx, 52, "test-q-67", "right")
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
		"a":           "right",
		"is_right":    true,
		"answered_at": now,
	}
	if diff := cmp.Diff(want, d.Data(), cmpopts.EquateApproxTime(time.Second)); diff != "" {
		t.Errorf("answer mismatch (-want +got):\n%s", diff)
	}
}

func TestSaveBet(t *testing.T) {
	os.Setenv("FIRESTORE_EMULATOR_HOST", "[::1]:8711")

	ctx := context.Background()

	c, err := firestore.NewClient(ctx, "test")
	if err != nil {
		t.Fatalf("app firestore: got err %v; want nil", err)
	}

	resetDB(ctx, c)

	c.Collection("answers").Doc("234:test-q-111").Set(ctx, map[string]any{
		"a": "has answer",
	})

	f := fstore.New(c)

	err = f.SaveBet(ctx, 234, "test-q-111", 500)
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
