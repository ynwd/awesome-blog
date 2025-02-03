package dto

type CreatePostRequest struct {
	Username    string `json:"username" binding:"required"`
	Title       string `json:"title" binding:"required"`
	Description string `json:"description" binding:"required"`
}

type PostResponse struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	Title       string `json:"title"`
	Description string `json:"description"`
}
