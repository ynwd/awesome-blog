package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ynwd/awesome-blog/internal/users/domain"
	"github.com/ynwd/awesome-blog/internal/users/dto"
	"github.com/ynwd/awesome-blog/internal/users/service"
	"github.com/ynwd/awesome-blog/pkg/res"
	"github.com/ynwd/awesome-blog/pkg/utils"
)

type UserHandler struct {
	userService service.UserService
	jwtToken    utils.JWT
}

func NewUserHandler(userService service.UserService, jwtToken utils.JWT) *UserHandler {
	return &UserHandler{
		userService: userService,
		jwtToken:    jwtToken,
	}
}

func (h *UserHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, res.Response{
			Status:  "error",
			Message: "Invalid request format",
		})
		return
	}

	user := domain.User{
		Username: req.Username,
		Password: req.Password,
	}

	if err := h.userService.CreateUser(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, res.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, res.Response{
		Status:  "success",
		Message: "User registered successfully",
	})
}

func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, res.Response{
			Status:  "error",
			Message: "Invalid request format",
		})
		return
	}

	user, err := h.userService.AuthenticateUser(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, res.Response{
			Status:  "error",
			Message: "Invalid credentials",
		})
		return
	}

	// Get client IP
	clientIP := c.ClientIP()

	// Generate token with IP
	token, err := h.jwtToken.GenerateToken(user.Username, clientIP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, res.Response{
			Status:  "error",
			Message: "Failed to generate token",
		})
		return
	}

	c.JSON(http.StatusOK, res.Response{
		Message: "Login successful",
		Status:  "success",
		Data:    token,
	})
}
