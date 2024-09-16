package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func FaviconHandler(c *gin.Context) {
	c.Status(http.StatusNoContent)
}
