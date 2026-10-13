package fstore_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/asahnoln/vote-bot/store"
	"github.com/asahnoln/vote-bot/store/fstore"
	"github.com/google/go-cmp/cmp"
)

func TestQuestion(t *testing.T) {
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
		Closed: false,
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

	q, err := f.CurrentQuestion(ctx)
	if err != nil {
		t.Fatalf("fstore question: got err %v; want nil", err)
	}

	want := store.Question{
		ID:    "test-q-1",
		Q:     "SaaaS or SoooS?",
		Opts:  []string{"jej", "joj", "sas", "sos"},
		A:     2,
		Order: 1,
	}
	if diff := cmp.Diff(want, q); diff != "" {
		t.Errorf("question mismatch (-want +got):\n%s", diff)
	}
}

func TestNoQuestion(t *testing.T) {
	ctx := context.Background()

	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	f := fstore.New(c)

	_, err := f.CurrentQuestion(ctx)
	if got, want := err, store.ErrNoQuestionsFound; !errors.Is(got, want) {
		t.Fatalf("fstore no questions: got err %v; want %v", got, want)
	}
}

func TestSomeQuestionHasNoOpts(t *testing.T) {
	ctx := context.Background()

	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	col := c.Collection("questions")
	col.Doc("test-q-0").Set(ctx, store.Question{
		Q:      "NO OPTS???",
		Closed: false,
		A:      1,
		Order:  0,
	})
	f := fstore.New(c)

	_, err := f.CurrentQuestion(ctx)
	if got, want := err, store.ErrQuestionHasNoOpts; !errors.Is(got, want) {
		t.Fatalf("fstore no questions: got err %v; want %v", got, want)
	}
}

func TestSomeQuestionAnswerIndexIsLarger(t *testing.T) {
	ctx := context.Background()

	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	tests := []struct {
		a int
	}{
		{100},
		{-999},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("index %d", tt.a), func(t *testing.T) {
			col := c.Collection("questions")
			col.Doc("test-q-0").Set(ctx, store.Question{
				Q:      "Answer is out of bounds???",
				Closed: false,
				A:      tt.a,
				Opts:   []string{"One answer..."},
				Order:  0,
			})
			f := fstore.New(c)

			_, err := f.CurrentQuestion(ctx)
			if got, want := err, store.ErrQuestionAnswerIndexIsOutOfBounds; !errors.Is(got, want) {
				t.Fatalf("fstore no questions: got err %v; want %v", got, want)
			}
		})
	}
}

func TestCloseCurrentQuestion(t *testing.T) {
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
		Closed: false,
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

	err := f.CloseCurrentQuestion(ctx)
	if err != nil {
		t.Fatalf("fstore close question: got err %v; want nil", err)
	}

	d, _ := c.Collection("questions").Doc("test-q-1").Get(ctx)
	got := store.Question{}
	d.DataTo(&got)
	want := store.Question{
		Q:      "SaaaS or SoooS?",
		Closed: true,
		A:      2,
		Opts:   []string{"jej", "joj", "sas", "sos"},
		Order:  1,
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("close question mismatch (-want +got):\n%s", diff)
	}
}

func TestCloseCurrentQuestionNoQuestionsError(t *testing.T) {
	ctx := context.Background()

	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	f := fstore.New(c)

	err := f.CloseCurrentQuestion(ctx)
	if got, want := err, store.ErrNoQuestionsFound; !errors.Is(got, want) {
		t.Fatalf("fstore no questions: got err %v; want %v", got, want)
	}
}
