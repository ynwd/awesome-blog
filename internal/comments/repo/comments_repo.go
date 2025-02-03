package repo

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/ynwd/awesome-blog/internal/comments/domain"
)

type commentsFirestore struct {
	client     *firestore.Client
	collection string
}

func NewCommentsRepository(client *firestore.Client) CommentsRepository {
	return &commentsFirestore{
		client:     client,
		collection: "comments",
	}
}

func (r *commentsFirestore) Create(ctx context.Context, comment domain.Comments) error {
	_, _, err := r.client.Collection(r.collection).Add(ctx, comment)
	return err
}
