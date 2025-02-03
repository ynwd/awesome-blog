package handler

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/ynwd/awesome-blog/internal/comments/domain"
	"github.com/ynwd/awesome-blog/internal/comments/service"
	"github.com/ynwd/awesome-blog/pkg/module"
	"github.com/ynwd/awesome-blog/pkg/utils"
)

type CommentsEventHandler struct {
	service service.CommentsService
}

func NewCommentsEventHandler(service service.CommentsService) *CommentsEventHandler {
	return &CommentsEventHandler{
		service: service,
	}
}

func (h *CommentsEventHandler) Handle(ctx context.Context, event module.BaseEvent) error {
	if event.Type != module.CommentEvent {
		return nil
	}
	log.Printf("Raw event received: %+v", event)
	payloadData, timeStamp := utils.EventParser(event)

	var payload domain.Comments
	if err := json.Unmarshal(payloadData, &payload); err != nil {
		log.Printf("Error unmarshaling payload: %v", err)
		return err
	}

	createdAt, _ := time.Parse(time.RFC3339, timeStamp)
	comments := domain.Comments{
		PostID:    payload.PostID,
		Username:  payload.Username,
		Comment:   payload.Comment,
		CreatedAt: createdAt,
	}

	err := h.service.CreateComment(ctx, comments)
	if err != nil {
		log.Printf("Error processing comment event: %v", err)
		return err
	}

	log.Printf("Successfully processed comment event: %+v", comments)
	return nil
}
