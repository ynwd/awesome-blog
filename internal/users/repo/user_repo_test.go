package repo

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ynwd/awesome-blog/internal/users/domain"
	"github.com/ynwd/awesome-blog/tests/helper"
)

func TestCreateUser(t *testing.T) {
	db := helper.SetupRepoClient(t)
	defer db.Close()

	repo := NewFirestoreUserRepository(db)

	// Clean up Firestore before and after the test
	err := helper.CleanDatabase()
	assert.NoError(t, err)

	defer helper.CleanDatabase()

	// Test data
	user := domain.User{
		Username: "testuser",
		Password: "testpass",
	}

	// Test Create
	err = repo.Create(context.Background(), user)
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
	client := helper.SetupRepoClient(t)
	defer client.Close()

	// Set the collection name for testing
	os.Setenv("GOOGLE_CLOUD_FIRESTORE_COLLECTION_USERS", "users")

	repo := NewFirestoreUserRepository(client)

	// Clean up Firestore before and after the test
	err := helper.CleanDatabase()
	assert.NoError(t, err)

	defer helper.CleanDatabase()

	// Test data
	user := domain.User{
		Username: "testuser",
		Password: "testpass",
	}

	// Create a user
	_, _, err = client.Collection("users").Add(context.Background(), user)
	assert.NoError(t, err)

	// Test GetByUsernameAndPassword
	retrievedUser, err := repo.GetByUsernameAndPassword(context.Background(), "testuser", "testpass")
	assert.NoError(t, err)
	assert.Equal(t, "testuser", retrievedUser.Username)
	assert.Equal(t, "testpass", retrievedUser.Password)
}

func TestIsUsernameExists(t *testing.T) {
	client := helper.SetupRepoClient(t)
	defer client.Close()

	// Set the collection name for testing
	os.Setenv("GOOGLE_CLOUD_FIRESTORE_COLLECTION_USERS", "users")

	repo := NewFirestoreUserRepository(client)

	// Clean up Firestore before and after the test
	err := helper.CleanDatabase()
	assert.NoError(t, err)

	defer helper.CleanDatabase()

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
