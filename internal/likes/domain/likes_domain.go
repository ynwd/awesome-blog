package domain

import "time"

type Likes struct {
	PostID       string    `json:"post_id" firestore:"post_id"`
	UsernameFrom string    `json:"username_from" firestore:"username_from"`
	CreatedAt    time.Time `json:"created_at,omitempty" firestore:"created_at"`
}
