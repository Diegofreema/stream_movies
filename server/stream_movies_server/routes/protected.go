package routes

import (
	"github.com/DiegoFreema/stream_movies/controllers"
	"github.com/DiegoFreema/stream_movies/middleware"
	"github.com/gin-gonic/gin"
)

func SetUpProtectedRoute(router *gin.Engine) {
	router.Use(middleware.AuthMiddleware())

	router.GET("/movies/:imdb_id", controllers.GetMovie())
	router.POST("/movies", controllers.AddMovie())
}
