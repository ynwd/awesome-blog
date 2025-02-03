package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ynwd/awesome-blog/internal/comments/domain"
	"github.com/ynwd/awesome-blog/internal/comments/dto"
	"github.com/ynwd/awesome-blog/internal/comments/service"
	"github.com/ynwd/awesome-blog/pkg/module"
	"github.com/ynwd/awesome-blog/pkg/pubsub"
	"github.com/ynwd/awesome-blog/pkg/res"
)

type CommentsHandler struct {
	commentsService service.CommentsService
	pubsub          pubsub.PubSubClient
}

func NewCommentsHandler(commentsService service.CommentsService, pubsub pubsub.PubSubClient) *CommentsHandler {
	return &CommentsHandler{
		commentsService: commentsService,
		pubsub:          pubsub,
	}
}

// CreateComment creates a new comment
func (h *CommentsHandler) CreateComment(c *gin.Context) {
	var req dto.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, res.Error(err.Error()))
		return
	}

	comment := domain.Comments{
		Username: req.Username,
		PostID:   req.PostID,
		Comment:  req.Comment,
	}

	if err := h.commentsService.CreateComment(c.Request.Context(), comment); err != nil {
		c.JSON(http.StatusInternalServerError, res.Error(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, res.Success(comment, "Comment created successfully"))
}

// PublishComment publishes a comment event to the pubsub
func (h *CommentsHandler) PublishComment(c *gin.Context) {
	var commentEvent domain.Comments
	if err := c.ShouldBindJSON(&commentEvent); err != nil {
		c.JSON(http.StatusBadRequest, res.Error("Invalid request payload"))
		return
	}
	event := module.BaseEvent{
		Type:      module.CommentEvent,
		Payload:   commentEvent,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	if err := h.pubsub.Publish(c.Request.Context(), event); err != nil {
		c.JSON(http.StatusInternalServerError, res.Error("Failed to publish comments event"))
		return
	}

	c.JSON(http.StatusCreated, res.Success(nil, "comments event published successfully"))
}
