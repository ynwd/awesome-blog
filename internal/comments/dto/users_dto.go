package dto

type CreateCommentRequest struct {
	Username string `json:"username" binding:"required"`
	PostID   string `json:"post_id" binding:"required"`
	Comment  string `json:"comment" binding:"required"`
}
