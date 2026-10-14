package fstore_test

import (
	"context"
	"os"
	"testing"

	firestore "cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/iterator"
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
	deleteCollection(ctx, c, "questions")
	deleteCollection(ctx, c, "users")
	deleteCollection(ctx, c, "answers")
}

func deleteCollection(ctx context.Context, c *firestore.Client, name string) {
	col := c.Collection(name)
	bulkwriter := c.BulkWriter(ctx)

	for {
		// Get a batch of documents
		iter := col.Documents(ctx)
		numDeleted := 0

		// Iterate through the documents, adding
		// a delete operation for each one to the BulkWriter.
		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}

			bulkwriter.Delete(doc.Ref)
			numDeleted++
		}

		// If there are no documents to delete,
		// the process is over.
		if numDeleted == 0 {
			bulkwriter.End()
			break
		}

		bulkwriter.Flush()
	}
}
