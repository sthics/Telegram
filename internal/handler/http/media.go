package http

import (
	"net/http"

	"github.com/ambarg/mini-telegram/internal/auth"
	"github.com/ambarg/mini-telegram/internal/service/media"
	"github.com/gin-gonic/gin"
)

type MediaHandler struct {
	service *media.Service
}

func NewMediaHandler(service *media.Service) *MediaHandler {
	return &MediaHandler{service: service}
}

type UploadRequest struct {
	Filename    string `json:"filename" binding:"required"`
	ContentType string `json:"contentType" binding:"required"`
}

// GetUploadURL godoc
// @Summary      Get presigned upload URL
// @Description  Get a URL to upload a file directly to object storage
// @Tags         media
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body UploadRequest true "Upload Request"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Router       /uploads/presigned [post]
func (h *MediaHandler) GetUploadURL(c *gin.Context) {
	userID, _ := auth.GetUserID(c)

	var req UploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	url, objectKey, err := h.service.GetUploadURL(c.Request.Context(), userID, req.Filename, req.ContentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"uploadUrl": url,
		"objectKey": objectKey,
	})
}
