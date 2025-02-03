package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ynwd/awesome-blog/internal/summary/dto"
	"github.com/ynwd/awesome-blog/internal/summary/service"
	"github.com/ynwd/awesome-blog/pkg/res"
)

type SummaryHandler struct {
	summaryService service.SummaryService
}

func NewSummaryHandler(summaryService service.SummaryService) *SummaryHandler {
	return &SummaryHandler{
		summaryService: summaryService,
	}
}

func (h *SummaryHandler) GetYearlySummary(c *gin.Context) {
	var req dto.SummaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, res.Error(err.Error()))
		return
	}

	summary, err := h.summaryService.GetYearlySummary(c.Request.Context(), req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, res.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, res.Success(summary, "Yearly summary retrieved successfully"))
}
