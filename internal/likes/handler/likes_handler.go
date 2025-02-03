package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ynwd/awesome-blog/internal/likes/domain"
	"github.com/ynwd/awesome-blog/internal/likes/dto"
	"github.com/ynwd/awesome-blog/internal/likes/service"
	"github.com/ynwd/awesome-blog/pkg/module"
	"github.com/ynwd/awesome-blog/pkg/pubsub"
	"github.com/ynwd/awesome-blog/pkg/res"
)

type LikesHandler struct {
	likesService service.LikesService
	pubsub       pubsub.PubSubClient
}

func NewLikesHandler(likesService service.LikesService, pubsubClient pubsub.PubSubClient) *LikesHandler {
	return &LikesHandler{
		likesService: likesService,
		pubsub:       pubsubClient,
	}
}

func (h *LikesHandler) RegisterRoutes(router *gin.Engine) {
	router.POST("/likes", h.CreateLike)
}

func (h *LikesHandler) CreateLike(c *gin.Context) {
	var req dto.CreateLikeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, res.Error(err.Error()))
		return
	}

	like := domain.Likes{
		PostID:       req.PostID,
		UsernameFrom: req.UsernameFrom,
	}

	if err := h.likesService.CreateLike(c.Request.Context(), like); err != nil {
		c.JSON(http.StatusInternalServerError, res.Error(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, res.Success(nil, "Like created successfully"))
}

func (h *LikesHandler) PublishLike(c *gin.Context) {
	var likeEvent domain.Likes
	if err := c.ShouldBindJSON(&likeEvent); err != nil {
		c.JSON(http.StatusBadRequest, res.Error("Invalid request payload"))
		return
	}

	event := module.BaseEvent{
		Type:      module.LikeEvent,
		Payload:   likeEvent,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	if err := h.pubsub.Publish(c.Request.Context(), event); err != nil {
		c.JSON(http.StatusInternalServerError, res.Error("Failed to publish likes event"))
		return
	}

	c.JSON(http.StatusCreated, res.Success(nil, "likes event published successfully"))
}
