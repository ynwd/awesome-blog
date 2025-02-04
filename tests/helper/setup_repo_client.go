package helper

import (
	"context"
	"log"
	"os"
	"testing"

	"cloud.google.com/go/firestore"
	"github.com/ynwd/awesome-blog/pkg/database"
)

func SetupRepoClient(t *testing.T) *firestore.Client {
	SetTestEnv()
	firestoreDB := database.NewFirestore(os.Getenv("GOOGLE_CLOUD_PROJECT_ID"),
		os.Getenv("GOOGLE_CLOUD_FIRESTORE_DATABASE_ID"))
	if err := firestoreDB.Connect(context.Background()); err != nil {
		t.Fatalf("Failed to connect to Firestore: %v", err)
	}
	client, err := firestoreDB.Client()
	if err != nil {
		log.Fatalf("Failed to get firestore client: %v", err)
		return nil
	}
	return client
}
