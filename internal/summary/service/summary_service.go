package service

import (
	"context"
	"errors"
	"time"

	"github.com/ynwd/awesome-blog/internal/summary/domain"
	"github.com/ynwd/awesome-blog/internal/summary/repo"
)

var (
	ErrInvalidUsername = errors.New("invalid username: cannot be empty")
)

type summaryService struct {
	summaryRepo repo.SummaryRepository
}

func NewSummaryService(summaryRepo repo.SummaryRepository) SummaryService {
	return &summaryService{
		summaryRepo: summaryRepo,
	}
}

func (s *summaryService) GetYearlySummary(ctx context.Context, username string) (*domain.SummaryData, error) {
	if username == "" {
		return nil, ErrInvalidUsername
	}

	currentYear := time.Now().Year()
	startDate := time.Date(currentYear, time.January, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(currentYear, time.December, 31, 23, 59, 59, 999999999, time.UTC)

	return s.summaryRepo.GetUserSummary(ctx, username, startDate, endDate)
}
