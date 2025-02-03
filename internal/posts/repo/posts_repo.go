package repo

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/ynwd/awesome-blog/internal/posts/domain"
)

type postsFirestore struct {
	client     *firestore.Client
	collection string
}

func NewPostsRepository(client *firestore.Client) PostsRepository {
	return &postsFirestore{
		client:     client,
		collection: "posts",
	}
}

func (r *postsFirestore) Create(ctx context.Context, post domain.Posts) (string, error) {
	doc, _, err := r.client.Collection(r.collection).Add(ctx, post)
	if err != nil {
		return "", err
	}
	return doc.ID, nil
}
