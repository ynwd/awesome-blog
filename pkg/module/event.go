package module

type EventType string

const (
	LikeEvent    EventType = "LIKE"
	PostEvent    EventType = "POST"
	CommentEvent EventType = "COMMENT"
)

type BaseEvent struct {
	Type      EventType   `json:"type"`
	Payload   interface{} `json:"payload"`
	Timestamp string      `json:"timestamp"`
}
