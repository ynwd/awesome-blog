package repo

import (
	"context"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/stretchr/testify/assert"
	"github.com/ynwd/awesome-blog/internal/likes/domain"
	"github.com/ynwd/awesome-blog/tests/helper"
)

// func setupTestEnv(_ *testing.T) (context.Context, *firestore.Client) {
// 	os.Setenv("GOOGLE_CLOUD_PROJECT_ID", "softion-playground")
// 	os.Setenv("GOOGLE_CLOUD_FIRESTORE_DATABASE_ID", "blogdb-yanu-widodo")
// 	os.Setenv("GOOGLE_CLOUD_FIRESTORE_COLLECTION_LIKES", "likes")

// 	ctx := context.Background()
// 	db := database.NewFirestore(os.Getenv("GOOGLE_CLOUD_PROJECT_ID"), os.Getenv("GOOGLE_CLOUD_FIRESTORE_DATABASE_ID"))
// 	db.Connect(ctx)
// 	client := db.Client()

// 	return ctx, client
// }

func cleanupCollection(ctx context.Context, client *firestore.Client, collection string) error {
	docs, err := client.Collection(collection).Documents(ctx).GetAll()
	if err != nil {
		return err
	}

	for _, doc := range docs {
		_, err := doc.Ref.Delete(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func TestLikesRepository_Create(t *testing.T) {
	ctx := context.Background()
	client := helper.SetupRepoClient(t)
	defer client.Close()

	// Cleanup before test
	err := cleanupCollection(ctx, client, "likes")
	assert.NoError(t, err)

	repo := NewLikesRepository(client)

	tests := []struct {
		name    string
		like    domain.Likes
		wantErr bool
	}{
		{
			name: "Success create like",
			like: domain.Likes{
				UsernameFrom: "testuser",
				PostID:       "test-post-123",
				CreatedAt:    time.Now(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(ctx, tt.like)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// Verify like was created
			docs, err := client.Collection("likes").
				Where("username_from", "==", tt.like.UsernameFrom).
				Where("post_id", "==", tt.like.PostID).
				Documents(ctx).GetAll()

			assert.NoError(t, err)
			assert.Len(t, docs, 1)

			createdLike := docs[0].Data()
			assert.Equal(t, tt.like.UsernameFrom, createdLike["username_from"])
			assert.Equal(t, tt.like.PostID, createdLike["post_id"])
		})
	}
}
