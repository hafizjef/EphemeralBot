package main

import (
	m "ephemeral/modules/middlewares"
	"ephemeral/modules/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	gin.DisableConsoleColor()

	router.Use(routes.InjectOnce())

	router.GET("/", routes.InjectRayID, routes.Home)
	router.GET("/login", routes.RequestTwitterLogin)

	router.POST("/login", m.AuthMiddleWare.LoginHandler)

	api := router.Group("/api")
	api.Use(routes.InjectRayID, m.AuthMiddleWare.MiddlewareFunc())
	{
		api.GET("/refresh_token", m.AuthMiddleWare.RefreshHandler)
		api.GET("/self", routes.GetSelf)
		api.GET("/tweets", routes.GetTweet)
		api.DELETE("/tweets", routes.DeleteAllTweets)
		api.GET("/stop", routes.StopDelete)
	}

	router.Run(":80")
}
