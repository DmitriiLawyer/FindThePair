package handlers

import "github.com/gin-gonic/gin"

func TestHandler(c *gin.Context) {
	c.File("./index2.html")
}
