package routes

import (
	"github.com/DiegoFreema/stream_movies/controllers"
	"github.com/gin-gonic/gin"
)

func SetUpUnProtectedRoute(router *gin.Engine) {

	router.GET("/movies", controllers.GetMovies())

	router.POST("/auth/register", controllers.RegisterUser())
	router.POST("/auth/login", controllers.LoginUser())
}
