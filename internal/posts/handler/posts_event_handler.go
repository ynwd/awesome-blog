package handler

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/ynwd/awesome-blog/internal/posts/domain"
	"github.com/ynwd/awesome-blog/internal/posts/service"
	"github.com/ynwd/awesome-blog/pkg/module"
	"github.com/ynwd/awesome-blog/pkg/utils"
)

type PostEventHandler struct {
	// repo    repo.PostsRepository
	service service.PostsService
}

func NewPostEventHandler(service service.PostsService) *PostEventHandler {
	return &PostEventHandler{
		service: service,
	}
}

func (h *PostEventHandler) Handle(ctx context.Context, event module.BaseEvent) error {
	if event.Type != module.PostEvent {
		return nil
	}
	log.Printf("Raw event received: %+v", event)
	payloadData, timeStamp := utils.EventParser(event)

	var payload domain.Posts
	if err := json.Unmarshal(payloadData, &payload); err != nil {
		log.Printf("Error unmarshaling payload: %v", err)
		return err
	}

	createdAt, _ := time.Parse(time.RFC3339, timeStamp)
	post := domain.Posts{
		Username:    payload.Username,
		Title:       payload.Title,
		Description: payload.Description,
		CreatedAt:   createdAt,
	}

	_, err := h.service.CreatePost(ctx, post)
	if err != nil {
		log.Printf("Error processing posts event: %v", err)
		return err
	}

	log.Printf("Successfully processed posts event: %+v", post)
	return nil
}
