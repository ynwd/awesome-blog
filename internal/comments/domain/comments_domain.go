package domain

import "time"

type Comments struct {
	Username  string    `json:"username" firestore:"username"`
	PostID    string    `json:"post_id" firestore:"post_id"`
	Comment   string    `json:"comment" firestore:"comment"`
	CreatedAt time.Time `json:"created_at,omitempty" firestore:"created_at"`
}
