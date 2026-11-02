package fstore_test

import (
	"context"
	"testing"

	"github.com/asahnoln/vote-bot/store"
	"github.com/asahnoln/vote-bot/store/fstore"
	"github.com/google/go-cmp/cmp"
)

func TestLastClosedQuestion(t *testing.T) {
	ctx := context.Background()

	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	c.Collection("config").Doc("current_question").Set(ctx, map[string]any{
		"idx": 2,
	})
	col := c.Collection("questions")
	col.Doc("test-q-0").Set(ctx, store.Question{
		Q:     "Should not be used?",
		Opts:  map[string]bool{"skip": false, "avoid": true},
		Order: 0,
	})
	col.Doc("test-q-1").Set(ctx, store.Question{
		Q:     "SaaaS or SoooS?",
		Opts:  map[string]bool{"jej": false, "joj": false, "sas": true, "sos": true},
		Order: 1,
	})
	col.Doc("test-q-2").Set(ctx, store.Question{
		Q:     "Should skip in future?",
		Opts:  map[string]bool{"SKIP": false, "FUTURE": true},
		Order: 2,
	})

	f := fstore.New(c)

	q, err := f.LastClosedQuestion(ctx)
	if err != nil {
		t.Fatalf("fstore question: got err %v; want nil", err)
	}

	want := store.Question{
		ID:    "test-q-1",
		Q:     "SaaaS or SoooS?",
		Opts:  map[string]bool{"jej": false, "joj": false, "sas": true, "sos": true},
		Order: 1,
	}
	if diff := cmp.Diff(want, q); diff != "" {
		t.Errorf("question mismatch (-want +got):\n%s", diff)
	}
}

func TestAnswers(t *testing.T) {
	ctx := context.Background()

	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	col := c.Collection("answers")
	col.Doc("8292:answer-q-51").Set(ctx, map[string]any{
		"user_id":     8292,
		"question_id": "answer-q-51",
		"bet":         100,
		"is_right":    true,
		"acc":         100,
	})
	col.Doc("156:answer-q-51").Set(ctx, map[string]any{
		"user_id":     156,
		"question_id": "answer-q-51",
		"all_in":      true,
		"is_right":    false,
		"acc":         200,
	})
	// Filter this out because wrong question ID
	col.Doc("928:skip-this-q-999").Set(ctx, map[string]any{
		"user_id":     928,
		"question_id": "skip-this-q-999",
		"is_right":    true,
		"acc":         -100,
	})

	f := fstore.New(c)

	answers, err := f.Answers(ctx, "answer-q-51")
	if err != nil {
		t.Fatalf("fstore answers: got err %v; want nil", err)
	}

	want := []store.Answer{
		{UserID: 156, IsRight: false, AllIn: true, Acc: 200},
		{UserID: 8292, Bet: 100, IsRight: true, Acc: 100},
	}
	if diff := cmp.Diff(want, answers); diff != "" {
		t.Errorf("answers mismatch (-want +got):\n%s", diff)
	}
}
