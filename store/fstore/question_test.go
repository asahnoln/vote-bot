package fstore_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/asahnoln/vote-bot/store"
	"github.com/asahnoln/vote-bot/store/fstore"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestQuestion(t *testing.T) {
	ctx := context.Background()

	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	c.Collection("config").Doc("current_question").Set(ctx, map[string]any{
		"idx": 1,
	})
	col := c.Collection("questions")
	col.Doc("test-q-0").Set(ctx, store.Question{
		Q:     "Should not be used?",
		A:     1,
		Opts:  []string{"skip", "avoid"},
		Order: 0,
	})
	col.Doc("test-q-1").Set(ctx, store.Question{
		Q:     "SaaaS or SoooS?",
		A:     2,
		Opts:  []string{"jej", "joj", "sas", "sos"},
		Order: 1,
	})
	col.Doc("test-q-2").Set(ctx, store.Question{
		Q:     "Should skip in future?",
		A:     1,
		Opts:  []string{"SKIP", "FUTURE"},
		Order: 2,
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

func TestNoQuestionIdx(t *testing.T) {
	ctx := context.Background()

	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	f := fstore.New(c)

	_, err := f.CurrentQuestion(ctx)
	if got, want := status.Code(err), codes.NotFound; err == nil || got != want {
		t.Fatalf("fstore no questions idx: got err %v; want %v", got, want)
	}
}

func TestNoQuestionIdxMissing(t *testing.T) {
	ctx := context.Background()

	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	c.Collection("config").Doc("current_question").Set(ctx, map[string]any{
		"somethingElse": "someData",
	})

	f := fstore.New(c)

	_, err := f.CurrentQuestion(ctx)
	if got, want := err.Error(), `no field "idx"`; !strings.Contains(got, want) {
		t.Fatalf("fstore no questions substring: got err %v; want substring %v", got, want)
	}
}

func TestNoQuestionIdxWrongType(t *testing.T) {
	ctx := context.Background()

	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	c.Collection("config").Doc("current_question").Set(ctx, map[string]any{
		"idx": "wrongType",
	})

	f := fstore.New(c)

	_, err := f.CurrentQuestion(ctx)
	if got, want := err.Error(), `is not int64`; !strings.Contains(got, want) {
		t.Fatalf("fstore no questions substring: got err %v; want substring %v", got, want)
	}
}

func TestNoQuestionIdxNegative(t *testing.T) {
	ctx := context.Background()

	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	c.Collection("config").Doc("current_question").Set(ctx, map[string]any{
		"idx": -1,
	})

	f := fstore.New(c)

	_, err := f.CurrentQuestion(ctx)
	if got, want := err, store.ErrQuestionsOffsetMustNotBeNegative; !errors.Is(got, want) {
		t.Fatalf("fstore no questions: got err %v; want %v", got, want)
	}
}

func TestNoQuestion(t *testing.T) {
	ctx := context.Background()

	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	c.Collection("config").Doc("current_question").Set(ctx, map[string]any{
		"idx": 0,
	})

	f := fstore.New(c)

	_, err := f.CurrentQuestion(ctx)
	if got, want := err, store.ErrNoOpenQuestionsLeft; !errors.Is(got, want) {
		t.Fatalf("fstore no questions: got err %v; want %v", got, want)
	}
}

func TestSomeQuestionHasNoOpts(t *testing.T) {
	ctx := context.Background()

	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	c.Collection("config").Doc("current_question").Set(ctx, map[string]any{
		"idx": 0,
	})
	col := c.Collection("questions")
	col.Doc("test-q-0").Set(ctx, store.Question{
		Q:     "NO OPTS???",
		A:     1,
		Order: 0,
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

	c.Collection("config").Doc("current_question").Set(ctx, map[string]any{
		"idx": 0,
	})
	for _, tt := range tests {
		t.Run(fmt.Sprintf("index %d", tt.a), func(t *testing.T) {
			col := c.Collection("questions")
			col.Doc("test-q-0").Set(ctx, store.Question{
				Q:     "Answer is out of bounds???",
				A:     tt.a,
				Opts:  []string{"One answer..."},
				Order: 0,
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

	c.Collection("config").Doc("current_question").Set(ctx, map[string]any{
		"idx": 1,
	})
	col := c.Collection("questions")
	col.Doc("test-q-0").Set(ctx, store.Question{
		Q:     "Should not be used?",
		A:     1,
		Opts:  []string{"skip", "avoid"},
		Order: 0,
	})
	col.Doc("test-q-1").Set(ctx, store.Question{
		Q:     "SaaaS or SoooS?",
		A:     2,
		Opts:  []string{"jej", "joj", "sas", "sos"},
		Order: 1,
	})
	col.Doc("test-q-2").Set(ctx, store.Question{
		Q:     "Should skip in future?",
		A:     1,
		Opts:  []string{"SKIP", "FUTURE"},
		Order: 2,
	})

	f := fstore.New(c)

	err := f.CloseCurrentQuestion(ctx)
	if err != nil {
		t.Fatalf("fstore close question: got err %v; want nil", err)
	}

	d, _ := c.Collection("config").Doc("current_question").Get(ctx)
	if got, want := d.Data()["idx"].(int64), int64(2); got != want {
		t.Errorf("current question idx: got %v; want %v", got, want)
	}
}

func TestCloseCurrentQuestionIdxBiggerThanQuestionCount(t *testing.T) {
	ctx := context.Background()

	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	c.Collection("config").Doc("current_question").Set(ctx, map[string]any{
		"idx": 1,
	})
	c.Collection("questions").Doc("test-q-0").Set(ctx, store.Question{
		Q:     "Should not be used?",
		A:     1,
		Opts:  []string{"skip", "avoid"},
		Order: 0,
	})
	f := fstore.New(c)

	err := f.CloseCurrentQuestion(ctx)
	if got, want := err, store.ErrNoOpenQuestionsLeft; !errors.Is(got, want) {
		t.Fatalf("fstore no questions: got err %v; want %v", got, want)
	}
}

func TestFirstQuestionMissed(t *testing.T) {
	ctx := context.Background()

	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	tests := []struct {
		idx  int
		want bool
	}{
		{1, true},
		{0, false},
	}

	f := fstore.New(c)

	for _, tt := range tests {
		t.Run(fmt.Sprintf("index %d", tt.idx), func(t *testing.T) {
			c.Collection("config").Doc("current_question").Set(ctx, map[string]any{
				"idx": tt.idx,
			})

			ok, err := f.FirstQuestionMissed(ctx)
			if err != nil {
				t.Fatalf("fstore question: got err %v; want nil", err)
			}

			if got, want := ok, tt.want; got != want {
				t.Errorf("fstore question missed: got err %v; want %v", got, want)
			}
		})
	}
}

func TestFirstQuestionMissedError(t *testing.T) {
	ctx := context.Background()

	c := firestoreMemoryEmulatorClient(t, ctx)
	resetDB(ctx, c)

	f := fstore.New(c)

	_, err := f.FirstQuestionMissed(ctx)
	if got, want := status.Code(err), codes.NotFound; err == nil || got != want {
		t.Fatalf("fstore first question missed idx: got err %v; want %v", got, want)
	}
}
