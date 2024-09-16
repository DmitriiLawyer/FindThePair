package handlers

import (
	"github.com/gin-gonic/gin"
)

func IndexHandler(c *gin.Context) {

	c.File("./public/index.html")
}
