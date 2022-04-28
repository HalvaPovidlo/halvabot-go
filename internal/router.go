package internal

import "github.com/gin-gonic/gin"

type Handler struct {
	super *gin.RouterGroup
}

func NewHandler(superGroup *gin.RouterGroup) *Handler {
	return &Handler{
		super: superGroup,
	}
}

func (h *Handler) Router() *gin.RouterGroup {
	discord := h.super.Group("/discord")
	return discord
}
