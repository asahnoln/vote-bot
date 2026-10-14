package fstore_test

import (
	"context"
	"os"
	"testing"

	firestore "cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

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

func firestoreMemoryEmulatorClient(t *testing.T, ctx context.Context) *firestore.Client {
	t.Helper()
	os.Setenv("FIRESTORE_EMULATOR_HOST", "[::1]:8711")

	c, err := firestore.NewClient(ctx, "test")
	if err != nil {
		t.Fatalf("app firestore: got err %v; want nil", err)
	}

	return c
}

func resetDB(ctx context.Context, c *firestore.Client) {
	col := c.Collection("questions")
	// TODO: Make easier by loop
	col.Doc("test-q-0").Delete(ctx)
	col.Doc("test-q-1").Delete(ctx)
	col.Doc("test-q-2").Delete(ctx)
	col.Doc("test-q-67").Delete(ctx)

	col = c.Collection("users")
	col.Doc("7829").Delete(ctx)
	col.Doc("390").Delete(ctx)
	col.Doc("67").Delete(ctx)
	col.Doc("101101").Delete(ctx)

	col = c.Collection("answers")
	c.Doc("8292:answer-q-51").Delete(ctx)
	c.Doc("156:answer-q-51").Delete(ctx)
	c.Doc("928:skip-this-q-999").Delete(ctx)
}
