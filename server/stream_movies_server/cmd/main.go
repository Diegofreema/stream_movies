package main

import (
	"fmt"

	"github.com/DiegoFreema/stream_movies/controllers"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	router.GET("/movies", controllers.GetMovies())

	if err := router.Run(":8080"); err != nil {
		fmt.Println("Failed to start server")
		panic(err)
	}

}
