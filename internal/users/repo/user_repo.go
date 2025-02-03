package repo

import (
	"context"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/ynwd/awesome-blog/internal/users/domain"
	"google.golang.org/api/iterator"
)

type userRepo struct {
	client     *firestore.Client
	collection string
}

func NewFirestoreUserRepository(client *firestore.Client) UserRepository {
	return &userRepo{
		client:     client,
		collection: os.Getenv("GOOGLE_CLOUD_FIRESTORE_COLLECTION_USERS"),
	}
}

func (r *userRepo) Create(ctx context.Context, user domain.User) error {
	_, _, err := r.client.
		Collection(r.collection).
		Add(ctx, user)

	return err
}

func (r *userRepo) GetByUsernameAndPassword(ctx context.Context, username string, password string) (domain.User, error) {
	iter := r.client.Collection(r.collection).
		Where("username", "==", username).
		Where("password", "==", password).
		Limit(1).
		Documents(ctx)

	doc, err := iter.Next()
	if err != nil {
		return domain.User{}, err
	}

	var user domain.User
	if err := doc.DataTo(&user); err != nil {
		return domain.User{}, err
	}

	user.Id = doc.Ref.ID
	return user, nil
}

func (r *userRepo) IsUsernameExists(ctx context.Context, username string) (bool, error) {
	iter := r.client.Collection(r.collection).
		Where("username", "==", username).
		Limit(1).
		Documents(ctx)

	doc, err := iter.Next()
	if err != nil && err != iterator.Done {
		return false, err
	}
	return doc != nil, nil
}
