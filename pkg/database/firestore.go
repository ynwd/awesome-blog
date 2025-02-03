package database

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
)

type FirestoreDB struct {
	ProjectID  string
	DatabaseID string
	client     *firestore.Client
}

func NewFirestore(projectID, databaseID string) *FirestoreDB {
	return &FirestoreDB{
		ProjectID:  projectID,
		DatabaseID: databaseID,
	}
}

func (db *FirestoreDB) Connect(ctx context.Context) error {
	client, err := firestore.NewClientWithDatabase(
		ctx,
		db.ProjectID,
		db.DatabaseID,
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

func (db *FirestoreDB) Client() *firestore.Client {
	return db.client
}
