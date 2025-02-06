package database

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
)

// FirestoreDB is a struct that holds the firestore client
type FirestoreDB struct {
	ProjectID  string
	DatabaseID string
	client     *firestore.Client
}

// NewFirestore creates a new FirestoreDB struct
func NewFirestore(projectID, databaseID string) *FirestoreDB {
	return &FirestoreDB{
		ProjectID:  projectID,
		DatabaseID: databaseID,
	}
}

// Connect creates a new firestore client and assigns it to the FirestoreDB struct
func (db *FirestoreDB) Connect(ctx context.Context) error {
	opt := option.WithCredentialsFile("serviceAccountKey.json")

	client, err := firestore.NewClientWithDatabase(
		ctx,
		db.ProjectID,
		db.DatabaseID,
		opt,
	)
	if err != nil {
		return fmt.Errorf("failed to create firestore client: %w", err)
	}

	db.client = client
	return nil
}

func (db *FirestoreDB) Close() error {
	if db.client != nil {
		return db.client.Close()
	}
	return nil
}

// Client returns the firestore client
func (db *FirestoreDB) Client() (*firestore.Client, error) {
	if db.client == nil {
		return nil, fmt.Errorf("firestore client is not initialized")
	}
	return db.client, nil
}
