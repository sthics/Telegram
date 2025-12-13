package http

import (
	"net/http"
	"strconv"

	"github.com/ambarg/mini-telegram/internal/domain"
	"github.com/ambarg/mini-telegram/internal/repository/redis"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	cacheRepo *redis.CacheRepository
	userRepo  domain.UserRepository
}

func NewUserHandler(cacheRepo *redis.CacheRepository, userRepo domain.UserRepository) *UserHandler {
	return &UserHandler{
		cacheRepo: cacheRepo,
		userRepo:  userRepo,
	}
}

// GetUserPresence godoc
// @Summary      Get user presence
// @Description  Get online status and last seen timestamp for a user
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int64  true  "User ID"
// @Success      200  {object}  map[string]any
// @Failure      400  {object}  map[string]string
// @Router       /users/{id}/presence [get]
func (h *UserHandler) GetUserPresence(c *gin.Context) {
	targetUserID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	online, lastSeen, err := h.cacheRepo.GetPresence(c.Request.Context(), targetUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"online":   online,
		"lastSeen": lastSeen,
	})
}

// SearchUsers godoc
// @Summary      Search users
// @Description  Search users by email or username
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        q    query     string  true  "Search Query"
// @Success      200  {array}   domain.User
// @Failure      400  {object}  map[string]string
// @Router       /users [get]
func (h *UserHandler) SearchUsers(c *gin.Context) {
	query := c.Query("q")
	if len(query) < 3 {
		c.JSON(http.StatusOK, []domain.User{})
		return
	}

	// Default limit=20, offset=0 for now, or read from query params if needed
	users, err := h.userRepo.SearchUsers(c.Request.Context(), query, 20, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

// GetProfile godoc
// @Summary      Get current user profile
// @Description  Get the profile of the authenticated user
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  domain.User
// @Failure      401  {object}  map[string]string
// @Router       /users/me [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, err := h.userRepo.GetByID(c.Request.Context(), userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

type UpdateProfileRequest struct {
	Username  *string `json:"username"`
	AvatarURL *string `json:"avatar_url"`
	Bio       *string `json:"bio"`
}

// UpdateProfile godoc
// @Summary      Update current user profile
// @Description  Update the profile of the authenticated user (username, avatar, bio)
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body UpdateProfileRequest true "Profile Update Request"
// @Success      200  {object}  domain.User
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Router       /users/me [patch]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get existing user
	user, err := h.userRepo.GetByID(c.Request.Context(), userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Apply updates
	if req.Username != nil {
		user.Username = *req.Username
	}
	if req.AvatarURL != nil {
		user.AvatarURL = *req.AvatarURL
	}
	if req.Bio != nil {
		user.Bio = *req.Bio
	}

	// Save
	if err := h.userRepo.Update(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

