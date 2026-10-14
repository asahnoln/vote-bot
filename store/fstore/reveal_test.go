package fstore_test

import (
	"context"
	"testing"

	"github.com/asahnoln/vote-bot/store"
	"github.com/asahnoln/vote-bot/store/fstore"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestLastClosedQuestion(t *testing.T) {
	ctx := context.Background()

	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	col := c.Collection("questions")
	col.Doc("test-q-0").Set(ctx, store.Question{
		Q:      "Should not be used?",
		Closed: true,
		A:      1,
		Opts:   []string{"skip", "avoid"},
		Order:  0,
	})
	col.Doc("test-q-1").Set(ctx, store.Question{
		Q:      "SaaaS or SoooS?",
		Closed: true,
		A:      2,
		Opts:   []string{"jej", "joj", "sas", "sos"},
		Order:  1,
	})
	col.Doc("test-q-2").Set(ctx, store.Question{
		Q:      "Should skip in future?",
		Closed: false,
		A:      1,
		Opts:   []string{"SKIP", "FUTURE"},
		Order:  2,
	})

	f := fstore.New(c)

	q, err := f.LastClosedQuestion(ctx)
	if err != nil {
		t.Fatalf("fstore question: got err %v; want nil", err)
	}

	want := store.Question{
		ID:     "test-q-1",
		Q:      "SaaaS or SoooS?",
		Opts:   []string{"jej", "joj", "sas", "sos"},
		Closed: true,
		A:      2,
		Order:  1,
	}
	if diff := cmp.Diff(want, q); diff != "" {
		t.Errorf("question mismatch (-want +got):\n%s", diff)
	}
}

func TestAnswer(t *testing.T) {
	ctx := context.Background()

	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	col := c.Collection("answers")
	col.Doc("8292:answer-q-51").Set(ctx, map[string]any{
		"is_right": true,
	})

	f := fstore.New(c)

	a, err := f.Answer(ctx, 8292, "answer-q-51")
	if err != nil {
		t.Fatalf("fstore answer: got err %v; want nil", err)
	}

	want := store.Answer{
		IsRight: true,
	}
	if diff := cmp.Diff(want, a); diff != "" {
		t.Errorf("answer mismatch (-want +got):\n%s", diff)
	}
}

func TestAnswerError(t *testing.T) {
	ctx := context.Background()

	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	f := fstore.New(c)

	_, err := f.Answer(ctx, 99999, "answer-q-99999")
	if got, want := status.Code(err), codes.NotFound; got != want {
		t.Errorf("got err code %v; want %v", got, want)
	}
}

func TestAccount(t *testing.T) {
	ctx := context.Background()

	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	col := c.Collection("answers")
	col.Doc("634:test-q-0").Set(ctx, map[string]any{
		"user_id":  634,
		"is_right": true,
		"bet":      200,
	})
	col.Doc("634:test-q-1").Set(ctx, map[string]any{
		"user_id":  634,
		"is_right": false,
		"bet":      100,
	})
	col.Doc("634:test-q-2").Set(ctx, map[string]any{
		"user_id":  634,
		"is_right": true,
		"bet":      500,
	})
	col.Doc("101:test-q-2").Set(ctx, map[string]any{
		"user_id":  101,
		"is_right": true,
		"bet":      9999,
	})

	f := fstore.New(c)

	acc, err := f.Account(ctx, 634)
	if err != nil {
		t.Fatalf("fstore account: got err %v; want nil", err)
	}

	if got, want := acc, 1600; got != want {
		t.Errorf("got acc %v; want %v", got, want)
	}
}
