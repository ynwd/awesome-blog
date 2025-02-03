package dto

type SummaryRequest struct {
	Username string `json:"username" binding:"required"`
}
