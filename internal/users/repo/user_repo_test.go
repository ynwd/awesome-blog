package repo

import (
	"context"
	"os"
	"testing"

	"cloud.google.com/go/firestore"
	"github.com/stretchr/testify/assert"
	"github.com/ynwd/awesome-blog/internal/users/domain"
	"github.com/ynwd/awesome-blog/pkg/database"
)

// setup environment variables
func setupEnv() {
	// Set the environment variables
	os.Setenv("GOOGLE_CLOUD_PROJECT_ID", "softion-playground")
	os.Setenv("GOOGLE_CLOUD_FIRESTORE_DATABASE_ID", "blogdb-yanu-widodo")
	os.Setenv("GOOGLE_CLOUD_FIRESTORE_COLLECTION_USERS", "users")
}

// setupFirestoreDB creates a Firestore client for testing.
func setupFirestoreDB(t *testing.T) *firestore.Client {
	setupEnv()
	firestoreDB := database.NewFirestore(os.Getenv("GOOGLE_CLOUD_PROJECT_ID"),
		os.Getenv("GOOGLE_CLOUD_FIRESTORE_DATABASE_ID"))
	if err := firestoreDB.Connect(context.Background()); err != nil {
		t.Fatalf("Failed to connect to Firestore: %v", err)
	}
	return firestoreDB.Client()
}

// cleanupFirestore deletes all documents in the given collection.
func cleanupFirestore(t *testing.T, client *firestore.Client, collection string) {
	ctx := context.Background()
	docs, err := client.Collection(collection).Documents(ctx).GetAll()
	if err != nil {
		t.Fatalf("Failed to get documents for cleanup: %v", err)
	}
	for _, doc := range docs {
		_, err := doc.Ref.Delete(ctx)
		if err != nil {
			t.Fatalf("Failed to delete document: %v", err)
		}
	}
}

func TestCreateUser(t *testing.T) {
	db := setupFirestoreDB(t)
	defer db.Close()

	repo := NewFirestoreUserRepository(db)

	// Clean up Firestore before and after the test
	cleanupFirestore(t, db, os.Getenv("GOOGLE_CLOUD_FIRESTORE_COLLECTION_USERS"))
	defer cleanupFirestore(t, db, os.Getenv("GOOGLE_CLOUD_FIRESTORE_COLLECTION_USERS"))

	// Test data
	user := domain.User{
		Username: "testuser",
		Password: "testpass",
	}

	// Test Create
	err := repo.Create(context.Background(), user)
	assert.NoError(t, err)

	// Verify the user was created
	iter := db.Collection("users").Where("username", "==", "testuser").Documents(context.Background())
	doc, err := iter.Next()
	assert.NoError(t, err)

	var retrievedUser domain.User
	err = doc.DataTo(&retrievedUser)
	assert.NoError(t, err)
	assert.Equal(t, "testuser", retrievedUser.Username)
	assert.Equal(t, "testpass", retrievedUser.Password)
}

func TestGetByUsernameAndPassword(t *testing.T) {
	client := setupFirestoreDB(t)
	defer client.Close()

	// Set the collection name for testing
	os.Setenv("GOOGLE_CLOUD_FIRESTORE_COLLECTION_USERS", "users")

	repo := NewFirestoreUserRepository(client)

	// Clean up Firestore before and after the test
	cleanupFirestore(t, client, os.Getenv("GOOGLE_CLOUD_FIRESTORE_COLLECTION_USERS"))
	defer cleanupFirestore(t, client, os.Getenv("GOOGLE_CLOUD_FIRESTORE_COLLECTION_USERS"))

	// Test data
	user := domain.User{
		Username: "testuser",
		Password: "testpass",
	}

	// Create a user
	_, _, err := client.Collection("users").Add(context.Background(), user)
	assert.NoError(t, err)

	// Test GetByUsernameAndPassword
	retrievedUser, err := repo.GetByUsernameAndPassword(context.Background(), "testuser", "testpass")
	assert.NoError(t, err)
	assert.Equal(t, "testuser", retrievedUser.Username)
	assert.Equal(t, "testpass", retrievedUser.Password)
}

func TestIsUsernameExists(t *testing.T) {
	client := setupFirestoreDB(t)
	defer client.Close()

	// Set the collection name for testing
	os.Setenv("GOOGLE_CLOUD_FIRESTORE_COLLECTION_USERS", "users")

	repo := NewFirestoreUserRepository(client)

	// Clean up Firestore before and after the test
	cleanupFirestore(t, client, os.Getenv("GOOGLE_CLOUD_FIRESTORE_COLLECTION_USERS"))
	defer cleanupFirestore(t, client, os.Getenv("GOOGLE_CLOUD_FIRESTORE_COLLECTION_USERS"))

	// Test data
	user := domain.User{
		Username: "testuser",
		Password: "testpass",
	}

	// Test IsUsernameExists before creating the user
	exists, err := repo.IsUsernameExists(context.Background(), "testuser")
	assert.NoError(t, err)
	assert.False(t, exists)

	// Create a user
	_, _, err = client.Collection("users").Add(context.Background(), user)
	assert.NoError(t, err)

	// Test IsUsernameExists after creating the user
	exists, err = repo.IsUsernameExists(context.Background(), "testuser")
	assert.NoError(t, err)
	assert.True(t, exists)
}
