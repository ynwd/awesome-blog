package repo

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/ynwd/awesome-blog/internal/posts/domain"
	"github.com/ynwd/awesome-blog/tests/helper"
)

func TestPostsFirestore_Create(t *testing.T) {
	client := helper.SetupFirestoreDB(t)
	defer client.Close()

	repo := NewPostsRepository(client)
	ctx := context.Background()

	tests := []struct {
		name    string
		post    domain.Posts
		wantErr bool
	}{
		{
			name: "successful post creation",
			post: domain.Posts{
				Username:    "testuser",
				Title:       "Test Post",
				Description: "Test Description",
				CreatedAt:   time.Now(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, err := repo.Create(ctx, tt.post)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, gotID)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, gotID)

			// Verify data persistence
			doc, err := client.Collection("posts").Doc(gotID).Get(ctx)
			assert.NoError(t, err)

			var savedPost domain.Posts
			err = doc.DataTo(&savedPost)
			assert.NoError(t, err)
			assert.Equal(t, tt.post.Username, savedPost.Username)
			assert.Equal(t, tt.post.Title, savedPost.Title)
			assert.Equal(t, tt.post.Description, savedPost.Description)
			assert.NotZero(t, savedPost.CreatedAt)

			// Cleanup
			_, err = client.Collection("posts").Doc(gotID).Delete(ctx)
			assert.NoError(t, err)
		})
	}
}

func TestPostsFirestore_GetAll(t *testing.T) {
	client := helper.SetupFirestoreDB(t)
	defer client.Close()

	repo := NewPostsRepository(client)
	ctx := context.Background()

	// Create test posts
	testPosts := []domain.Posts{
		{
			Username:    "user1",
			Title:       "Post 1",
			Description: "Description 1",
			CreatedAt:   time.Now(),
		},
		{
			Username:    "user2",
			Title:       "Post 2",
			Description: "Description 2",
			CreatedAt:   time.Now(),
		},
	}

	var createdIDs []string
	for _, post := range testPosts {
		id, err := repo.Create(ctx, post)
		assert.NoError(t, err)
		createdIDs = append(createdIDs, id)
	}

	// Cleanup
	for _, id := range createdIDs {
		_, err := client.Collection("posts").Doc(id).Delete(ctx)
		assert.NoError(t, err)
	}
}
