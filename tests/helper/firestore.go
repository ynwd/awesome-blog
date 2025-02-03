package helper

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"cloud.google.com/go/firestore"
	"github.com/ynwd/awesome-blog/pkg/database"
)

func setupEnv() {
	os.Setenv("GOOGLE_CLOUD_PROJECT_ID", "softion-playground")
	os.Setenv("GOOGLE_CLOUD_FIRESTORE_DATABASE_ID", "blogdb-yanu-widodo")
	os.Setenv("GOOGLE_CLOUD_FIRESTORE_COLLECTION_POSTS", "posts")
}

func SetupFirestoreDB(t *testing.T) *firestore.Client {
	setupEnv()
	firestoreDB := database.NewFirestore(os.Getenv("GOOGLE_CLOUD_PROJECT_ID"),
		os.Getenv("GOOGLE_CLOUD_FIRESTORE_DATABASE_ID"))
	if err := firestoreDB.Connect(context.Background()); err != nil {
		t.Fatalf("Failed to connect to Firestore: %v", err)
	}
	return firestoreDB.Client()
}

func CleanDatabase() error {
	ctx := context.Background()
	SetTestEnv()
	db := database.NewFirestore("softion-playground", "blogdb-yanu-widodo")
	db.Connect(ctx)
	client := db.Client()
	collections := []string{"users", "posts", "comments", "likes"}
	for _, col := range collections {
		docs, err := client.Collection(col).Documents(ctx).GetAll()
		if err != nil {
			return fmt.Errorf("failed to get documents from %s: %v", col, err)
		}

		for _, doc := range docs {
			_, err := doc.Ref.Delete(ctx)
			if err != nil {
				return fmt.Errorf("failed to delete document from %s: %v", col, err)
			}
		}
		log.Printf("Cleaned collection: %s", col)
	}

	return nil
}

func CleanupFirestore(t *testing.T, client *firestore.Client, collection string) {
	ctx := context.Background()
	docs, err := client.Collection(collection).Documents(ctx).GetAll()
	if err != nil {
		t.Fatalf("Failed to get documents for cleanup: %v", err)
	}

	for _, doc := range docs {
		_, err := doc.Ref.Delete(ctx)
		if err != nil {
			t.Fatalf("Failed to delete document during cleanup: %v", err)
		}
	}
}
