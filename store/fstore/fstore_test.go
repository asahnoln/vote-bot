package fstore_test

import (
	"context"
	"testing"

	firebase "firebase.google.com/go/v4"
	"github.com/asahnoln/vote-bot/store"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/option"
)

func TestQuestion(t *testing.T) {
	opt := option.WithAuthCredentialsFile(option.ServiceAccount, "testdata/firebase-adminsdk.json")
	ctx := context.Background()
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		t.Fatalf("firebase new app: got err %v; want nil", err)
	}

	c, err := app.Firestore(ctx)
	if err != nil {
		t.Fatalf("app firestore: got err %v; want nil", err)
	}

	col := c.Collection("questions")
	_, err = col.Doc("test-q-0").Set(ctx, store.Question{
		Q:      "Should not be used?",
		Closed: true,
		A:      1,
		Opts:   []string{"skip", "avoid"},
		Order:  0,
	})
	if err != nil {
		t.Fatalf("test data add 0: got err %v; want nil", err)
	}
	_, err = col.Doc("test-q-1").Set(ctx, store.Question{
		Q:      "SaaaS or SoooS?",
		Closed: false,
		A:      2,
		Opts:   []string{"jej", "joj", "sas", "sos"},
		Order:  1,
	})
	if err != nil {
		t.Fatalf("test data add 1: got err %v; want nil", err)
	}

	f := fstore.New(c)

	q, err := f.CurrentQuestion(ctx)
	if err != nil {
		t.Fatalf("fstore question: got err %v; want nil", err)
	}

	want := store.Question{
		Q:    "SaaaS or SoooS?",
		Opts: []string{"jej", "joj", "sas", "sos"},
	}
	if diff := cmp.Diff(want, q); diff != "" {
		t.Errorf("question mismatch (-want +got):\n%s", diff)
	}
}
