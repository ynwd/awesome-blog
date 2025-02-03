package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ynwd/awesome-blog/internal/posts/domain"
	"github.com/ynwd/awesome-blog/internal/posts/dto"
	"github.com/ynwd/awesome-blog/internal/posts/service"
	"github.com/ynwd/awesome-blog/pkg/module"
	"github.com/ynwd/awesome-blog/pkg/pubsub"
	"github.com/ynwd/awesome-blog/pkg/res"
)

type PostsHandler struct {
	postsService service.PostsService
	pubsub       pubsub.PubSubClient
}

func NewPostsHandler(postsService service.PostsService, pubsubClient pubsub.PubSubClient) *PostsHandler {
	return &PostsHandler{
		postsService: postsService,
		pubsub:       pubsubClient,
	}
}

func (h *PostsHandler) CreatePost(c *gin.Context) {
	var req dto.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, res.Error(err.Error()))
		return
	}

	post := domain.Posts{
		Username:    req.Username,
		Title:       req.Title,
		Description: req.Description,
	}

	postID, err := h.postsService.CreatePost(c.Request.Context(), post)
	if err != nil {
		c.JSON(http.StatusInternalServerError, res.Error(err.Error()))
		return
	}

	response := dto.PostResponse{
		ID:          postID,
		Username:    post.Username,
		Title:       post.Title,
		Description: post.Description,
	}

	c.JSON(http.StatusCreated, res.Success(response, "Post created successfully"))
}

func (h *PostsHandler) PublishPost(c *gin.Context) {
	var postEvent domain.Posts
	if err := c.ShouldBindJSON(&postEvent); err != nil {
		c.JSON(http.StatusBadRequest, res.Error("Invalid request payload"))
		return
	}

	event := module.BaseEvent{
		Type:      module.PostEvent,
		Payload:   postEvent,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	if err := h.pubsub.Publish(c.Request.Context(), event); err != nil {
		c.JSON(http.StatusInternalServerError, res.Error("Failed to publish posts event"))
		return
	}

	c.JSON(http.StatusCreated, res.Success(nil, "posts event published successfully"))
}
