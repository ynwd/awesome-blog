package service

import (
	"context"

	"github.com/ynwd/awesome-blog/internal/summary/domain"
)

type SummaryService interface {
	GetYearlySummary(ctx context.Context, username string) (*domain.SummaryData, error)
}
