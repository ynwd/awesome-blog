package dto

type CreateLikeRequest struct {
	PostID       string `json:"post_id" binding:"required"`
	UsernameFrom string `json:"username_from" binding:"required"`
}

type DeleteLikeRequest struct {
	PostID       string `json:"post_id" binding:"required"`
	UsernameFrom string `json:"username_from" binding:"required"`
}
