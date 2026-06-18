package main

import (
	"fmt"

	"github.com/DiegoFreema/stream_movies/controllers"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	router.GET("/movies", controllers.GetMovies())
	router.GET("/movies/:imdb_id", controllers.GetMovie())
	router.POST("/movies", controllers.AddMovie())

	router.POST("/users", controllers.RegisterUser())
	if err := router.Run(":8080"); err != nil {
		fmt.Println("Failed to start server")
		panic(err)
	}

}
