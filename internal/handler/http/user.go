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
// @Description  Search users by email
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
