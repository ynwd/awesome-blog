package repo

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/ynwd/awesome-blog/internal/summary/domain"
	"google.golang.org/api/iterator"
)

type SummaryRepository interface {
	GetUserSummary(ctx context.Context, username string, startDate, endDate time.Time) (*domain.SummaryData, error)
}

type summaryFirestore struct {
	client *firestore.Client
}

func NewSummaryRepository(client *firestore.Client) SummaryRepository {
	return &summaryFirestore{
		client: client,
	}
}

func (r *summaryFirestore) GetUserSummary(ctx context.Context, username string, startDate, endDate time.Time) (*domain.SummaryData, error) {
	// Initialize summary data
	data := &domain.SummaryData{
		Likes:    make(map[string]int64),
		Comments: make(map[string]int64),
		Posts:    make(map[string]int64),
	}

	// Get likes
	likesIter := r.client.Collection("likes").
		Where("created_at", ">=", startDate).
		Where("created_at", "<=", endDate).
		Where("username_from", "==", username).
		Documents(ctx)

	err := r.processDocuments(likesIter, data.Likes)
	if err != nil {
		return nil, err
	}

	// Get comments
	commentsIter := r.client.Collection("comments").
		Where("created_at", ">=", startDate).
		Where("created_at", "<=", endDate).
		Where("username", "==", username).
		Documents(ctx)

	err = r.processDocuments(commentsIter, data.Comments)
	if err != nil {
		return nil, err
	}

	// Get posts
	postsIter := r.client.Collection("posts").
		Where("created_at", ">=", startDate).
		Where("created_at", "<=", endDate).
		Where("username", "==", username).
		Documents(ctx)

	err = r.processDocuments(postsIter, data.Posts)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (r *summaryFirestore) processDocuments(iter *firestore.DocumentIterator, data map[string]int64) error {
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}

		docData := doc.Data()
		createdAt, ok := docData["created_at"].(time.Time)
		if !ok {
			continue // Skip if created_at is not a valid time
		}

		month := createdAt.Format("2006-01")
		data[month]++
	}
	return nil
}
