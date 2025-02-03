package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/ynwd/awesome-blog/internal/summary/domain"
	"github.com/ynwd/awesome-blog/internal/summary/dto"
	"github.com/ynwd/awesome-blog/pkg/res"
)

type mockSummaryService struct {
	getYearlySummaryFunc func(ctx context.Context, username string) (*domain.SummaryData, error)
}

func (m *mockSummaryService) GetYearlySummary(ctx context.Context, username string) (*domain.SummaryData, error) {
	if m.getYearlySummaryFunc != nil {
		return m.getYearlySummaryFunc(ctx, username)
	}
	return nil, nil
}

func TestSummaryHandler_GetYearlySummary(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		payload    interface{}
		setupMock  func(*mockSummaryService)
		wantStatus int
		wantRes    res.Response
	}{
		{
			name:       "invalid json payload",
			payload:    "invalid",
			setupMock:  func(m *mockSummaryService) {},
			wantStatus: http.StatusBadRequest,
			wantRes: res.Response{
				Status:  "error",
				Message: "json: cannot unmarshal string into Go value of type dto.SummaryRequest",
			},
		},
		{
			name: "service error",
			payload: dto.SummaryRequest{
				Username: "testuser",
			},
			setupMock: func(m *mockSummaryService) {
				m.getYearlySummaryFunc = func(ctx context.Context, username string) (*domain.SummaryData, error) {
					return nil, errors.New("service error")
				}
			},
			wantStatus: http.StatusInternalServerError,
			wantRes: res.Response{
				Status:  "error",
				Message: "service error",
			},
		},
		{
			name: "success get summary",
			payload: dto.SummaryRequest{
				Username: "testuser",
			},
			setupMock: func(m *mockSummaryService) {
				m.getYearlySummaryFunc = func(ctx context.Context, username string) (*domain.SummaryData, error) {
					return &domain.SummaryData{
						Likes:    map[string]int64{"2025-02": 1},
						Comments: map[string]int64{"2025-02": 1},
						Posts:    map[string]int64{"2025-02": 1},
					}, nil
				}
			},
			wantStatus: http.StatusOK,
			wantRes: res.Response{
				Status:  "success",
				Message: "Yearly summary retrieved successfully",
				Data: map[string]interface{}{
					"likes":    map[string]interface{}{"2025-02": float64(1)},
					"comments": map[string]interface{}{"2025-02": float64(1)},
					"posts":    map[string]interface{}{"2025-02": float64(1)},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			mockService := &mockSummaryService{}
			tt.setupMock(mockService)

			handler := NewSummaryHandler(mockService)
			router.POST("/summary", handler.GetYearlySummary)

			payloadBytes, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/summary", bytes.NewBuffer(payloadBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			var got res.Response
			err := json.Unmarshal(w.Body.Bytes(), &got)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantRes, got)
		})
	}
}
