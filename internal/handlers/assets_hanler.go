package handlers

import "github.com/gin-gonic/gin"

func AssetsHandler(router *gin.Engine) {
	router.Static("./assets", "./public/assets")
}
