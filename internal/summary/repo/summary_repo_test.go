package repo

import (
	"context"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/stretchr/testify/assert"
	"github.com/ynwd/awesome-blog/internal/summary/domain"
	"github.com/ynwd/awesome-blog/tests/helper"
)

func createTestData(ctx context.Context, client *firestore.Client, testDate time.Time) error {
	testData := []struct {
		collection string
		id         string
		data       map[string]interface{}
	}{
		{
			collection: "likes",
			id:         "like1",
			data: map[string]interface{}{
				"username_from": "testuser",
				"post_id":       "post1",
				"created_at":    testDate,
			},
		},
		{
			collection: "comments",
			id:         "comment1",
			data: map[string]interface{}{
				"username":   "testuser",
				"post_id":    "post1",
				"created_at": testDate,
			},
		},
		{
			collection: "posts",
			id:         "post1",
			data: map[string]interface{}{
				"username":   "testuser",
				"created_at": testDate,
			},
		},
	}

	for _, td := range testData {
		_, err := client.Collection(td.collection).Doc(td.id).Set(ctx, td.data)
		if err != nil {
			return err
		}
	}
	return nil
}

func cleanDb(t *testing.T, client *firestore.Client) {
	helper.CleanupFirestore(t, client, "likes")
	helper.CleanupFirestore(t, client, "comments")
	helper.CleanupFirestore(t, client, "posts")
	helper.CleanupFirestore(t, client, "users")
}

func TestSummaryRepository_GetUserSummary(t *testing.T) {
	ctx := context.Background()
	client := helper.SetupRepoClient(t)
	cleanDb(t, client)
	defer cleanDb(t, client)

	testDate := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)
	err := createTestData(ctx, client, testDate)
	assert.NoError(t, err)

	repo := NewSummaryRepository(client)

	tests := []struct {
		name      string
		username  string
		startDate time.Time
		endDate   time.Time
		want      *domain.SummaryData
		wantErr   bool
	}{
		{
			name:      "get summary for existing user",
			username:  "testuser",
			startDate: testDate.AddDate(0, 0, -1),
			endDate:   testDate.AddDate(0, 0, 1),
			want: &domain.SummaryData{
				Likes:    map[string]int64{"2025-02": 1},
				Comments: map[string]int64{"2025-02": 1},
				Posts:    map[string]int64{"2025-02": 1},
			},
			wantErr: false,
		},
		{
			name:      "get summary for non-existing user",
			username:  "nonexistent",
			startDate: time.Now().AddDate(0, 0, -7),
			endDate:   time.Now().AddDate(0, 0, 1),
			want: &domain.SummaryData{
				Likes:    map[string]int64{},
				Comments: map[string]int64{},
				Posts:    map[string]int64{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetUserSummary(ctx, tt.username, tt.startDate, tt.endDate)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
