package handler

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/ynwd/awesome-blog/internal/likes/domain"
	"github.com/ynwd/awesome-blog/internal/likes/service"
	"github.com/ynwd/awesome-blog/pkg/module"
	"github.com/ynwd/awesome-blog/pkg/utils"
)

type LikeEventHandler struct {
	// likesRepo    repo.LikesRepository
	service service.LikesService
}

func NewLikeEventHandler(service service.LikesService) *LikeEventHandler {
	return &LikeEventHandler{
		service: service,
	}
}

func (h *LikeEventHandler) Handle(ctx context.Context, event module.BaseEvent) error {
	if event.Type != module.LikeEvent {
		return nil
	}
	log.Printf("Raw event received: %+v", event)
	payloadData, timeStamp := utils.EventParser(event)

	var payload domain.Likes
	if err := json.Unmarshal(payloadData, &payload); err != nil {
		log.Printf("Error unmarshaling payload: %v", err)
		return err
	}

	createdAt, _ := time.Parse(time.RFC3339, timeStamp)
	like := domain.Likes{
		PostID:       payload.PostID,
		UsernameFrom: payload.UsernameFrom,
		CreatedAt:    createdAt,
	}

	if err := h.service.CreateLike(ctx, like); err != nil {
		log.Printf("Error saving like: %v", err)
		return err
	}

	log.Printf("Successfully processed like event: %+v", like)
	return nil
}
