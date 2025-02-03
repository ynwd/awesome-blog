package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/ynwd/awesome-blog/internal/summary/domain"
)

type mockSummaryRepository struct {
	getUserSummaryFunc func(ctx context.Context, username string, startDate, endDate time.Time) (*domain.SummaryData, error)
}

func (m *mockSummaryRepository) GetUserSummary(ctx context.Context, username string, startDate, endDate time.Time) (*domain.SummaryData, error) {
	if m.getUserSummaryFunc != nil {
		return m.getUserSummaryFunc(ctx, username, startDate, endDate)
	}
	return nil, nil
}

func TestSummaryService_GetYearlySummary(t *testing.T) {
	tests := []struct {
		name      string
		username  string
		setupMock func(*mockSummaryRepository)
		want      *domain.SummaryData
		wantErr   error
	}{
		{
			name:      "empty username returns error",
			username:  "",
			setupMock: func(m *mockSummaryRepository) {},
			want:      nil,
			wantErr:   ErrInvalidUsername,
		},
		{
			name:     "repository error returns error",
			username: "testuser",
			setupMock: func(m *mockSummaryRepository) {
				m.getUserSummaryFunc = func(ctx context.Context, username string, startDate, endDate time.Time) (*domain.SummaryData, error) {
					return nil, errors.New("repository error")
				}
			},
			want:    nil,
			wantErr: errors.New("repository error"),
		},
		{
			name:     "success get yearly summary",
			username: "testuser",
			setupMock: func(m *mockSummaryRepository) {
				m.getUserSummaryFunc = func(ctx context.Context, username string, startDate, endDate time.Time) (*domain.SummaryData, error) {
					return &domain.SummaryData{
						Likes:    map[string]int64{"2025-02": 1},
						Comments: map[string]int64{"2025-02": 1},
						Posts:    map[string]int64{"2025-02": 1},
					}, nil
				}
			},
			want: &domain.SummaryData{
				Likes:    map[string]int64{"2025-02": 1},
				Comments: map[string]int64{"2025-02": 1},
				Posts:    map[string]int64{"2025-02": 1},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockSummaryRepository{}
			tt.setupMock(mockRepo)

			service := NewSummaryService(mockRepo)
			got, err := service.GetYearlySummary(context.Background(), tt.username)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
