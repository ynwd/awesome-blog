package domain

import "time"

type Posts struct {
	Username    string    `json:"username" firestore:"username"`
	Title       string    `json:"title"  firestore:"title"`
	Description string    `json:"description" firestore:"description"`
	CreatedAt   time.Time `json:"created_at,omitempty" firestore:"created_at"`
}
