package repo

import (
	"context"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/ynwd/awesome-blog/internal/comments/domain"
	postDomain "github.com/ynwd/awesome-blog/internal/posts/domain"
	postRepo "github.com/ynwd/awesome-blog/internal/posts/repo"
	userDomain "github.com/ynwd/awesome-blog/internal/users/domain"
	userRepo "github.com/ynwd/awesome-blog/internal/users/repo"
	"github.com/ynwd/awesome-blog/tests/helper"
)

func cleanupFirestore(t *testing.T, client *firestore.Client, collection string) {
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

var postID string

// setup users and posts
func setupUsersAndPosts(t *testing.T, client *firestore.Client) {

	userRepo := userRepo.NewFirestoreUserRepository(client)
	postRepo := postRepo.NewPostsRepository(client)

	ctx := context.Background()

	err := userRepo.Create(ctx, userDomain.User{
		Username: "user-123",
		Password: "password-123",
	})
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	p, err := postRepo.Create(ctx, postDomain.Posts{
		Username:    "user-123",
		Title:       "Test Post",
		Description: "Test post content",
		CreatedAt:   time.Now(),
	})

	postID = p

	if err != nil {
		t.Fatalf("Failed to create post: %v", err)
	}

}

func cleanDatabase(t *testing.T, client *firestore.Client) {
	cleanupFirestore(t, client, "comments")
	cleanupFirestore(t, client, "users")
	cleanupFirestore(t, client, "posts")
}

func TestCommentsRepository_Create(t *testing.T) {
	client := helper.SetupRepoClient(t)
	setupUsersAndPosts(t, client)
	defer client.Close()

	repo := NewCommentsRepository(client)

	tests := []struct {
		name    string
		comment domain.Comments
		wantErr bool
	}{
		{
			name: "Success create comment",
			comment: domain.Comments{
				Username:  "user-123",
				PostID:    postID,
				Comment:   "Test comment content",
				CreatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "Success create another comment",
			comment: domain.Comments{
				Username:  "user-123",
				PostID:    postID,
				Comment:   "Another test comment",
				CreatedAt: time.Now(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer cleanDatabase(t, client)

			ctx := context.Background()
			err := repo.Create(ctx, tt.comment)

			if (err != nil) != tt.wantErr {
				t.Errorf("commentsFirestore.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify comment was created
			docs, err := client.Collection("comments").Documents(ctx).GetAll()
			if err != nil {
				t.Fatalf("Failed to get comments: %v", err)
			}

			if len(docs) == 0 {
				t.Errorf("Expected more than 0 comment, got %d", len(docs))
			}

			var savedComment domain.Comments
			if err := docs[0].DataTo(&savedComment); err != nil {
				t.Fatalf("Failed to parse comment data: %v", err)
			}

			if savedComment.Username != tt.comment.Username {
				t.Errorf("Expected UserID %s, got %s", tt.comment.Username, savedComment.Username)
			}
			// if savedComment.PostID != tt.comment.PostID {
			// 	t.Errorf("Expected PostID %s, got %s", tt.comment.PostID, savedComment.PostID)
			// }
			// if savedComment.Comment != tt.comment.Comment {
			// 	t.Errorf("Expected Content %s, got %s", tt.comment.Comment, savedComment.Comment)
			// }
		})
	}
}
