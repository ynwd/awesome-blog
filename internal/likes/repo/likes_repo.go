package repo

import (
	"context"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/ynwd/awesome-blog/internal/likes/domain"
)

type likesFirestore struct {
	client     *firestore.Client
	collection string
}

func NewLikesRepository(client *firestore.Client) LikesRepository {
	return &likesFirestore{
		client:     client,
		collection: os.Getenv("GOOGLE_CLOUD_FIRESTORE_COLLECTION_LIKES"),
	}
}

func (r *likesFirestore) Create(ctx context.Context, like domain.Likes) error {
	_, _, err := r.client.Collection(r.collection).Add(ctx, like)
	return err
}
