package fstore_test

import (
	"context"
	"errors"
	"os"
	"testing"

	firestore "cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"github.com/asahnoln/vote-bot/store"
	"github.com/asahnoln/vote-bot/store/fstore"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/option"
)

func TestQuestion(t *testing.T) {
	os.Setenv("FIRESTORE_EMULATOR_HOST", "[::1]:8711")

	ctx := context.Background()

	c, err := firestore.NewClient(ctx, "test")
	if err != nil {
		t.Fatalf("app firestore: got err %v; want nil", err)
	}

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
	os.Setenv("FIRESTORE_EMULATOR_HOST", "[::1]:8711")

	ctx := context.Background()

	c, err := firestore.NewClient(ctx, "test")
	if err != nil {
		t.Fatalf("app firestore: got err %v; want nil", err)
	}

	col := c.Collection("questions")
	col.Doc("test-q-0").Delete(ctx)
	col.Doc("test-q-1").Delete(ctx)
	col.Doc("test-q-2").Delete(ctx)

	f := fstore.New(c)

	_, err = f.CurrentQuestion(ctx)
	if got, want := err, store.ErrNoQuestionsFound; !errors.Is(got, want) {
		t.Fatalf("fstore no questions: got err %v; want %v", got, want)
	}
}

func firestoreIntegrationClient(t testing.T, ctx context.Context) *firestore.Client {
	t.Helper()

	opt := option.WithAuthCredentialsFile(option.ServiceAccount, "testdata/firebase-adminsdk.json")
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		t.Fatalf("firebase new app: got err %v; want nil", err)
	}

	c, err := app.Firestore(ctx)
	if err != nil {
		t.Fatalf("app firestore: got err %v; want nil", err)
	}

	return c
}
