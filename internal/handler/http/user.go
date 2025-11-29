package http

import (
	"net/http"
	"strconv"

	"github.com/ambarg/mini-telegram/internal/repository/redis"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	cacheRepo *redis.CacheRepository
}

func NewUserHandler(cacheRepo *redis.CacheRepository) *UserHandler {
	return &UserHandler{cacheRepo: cacheRepo}
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
