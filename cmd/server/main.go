package main

import (
	"GinAndrew/internal/handlers"
	"GinAndrew/internal/websocket"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	//gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	// Стастический маршрут для папки assets
	handlers.AssetsHandler(router)

	// Маршрут для favicon
	router.GET("/favicon", handlers.FaviconHandler)

	// WS
	router.GET("/ws", func(c *gin.Context) {
		websocket.HandleConnections(c.Writer, c.Request)
	})

	//Если маршрута не существует
	router.NoRoute(handlers.IndexHandler)

	err := router.Run(":8082")
	{
		if err != nil {
			log.Fatal("Sever isnt running")
		}
	}
}
