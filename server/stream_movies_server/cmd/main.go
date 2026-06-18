package main

import (
	"fmt"

	"github.com/DiegoFreema/stream_movies/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	routes.SetUpUnProtectedRoute(router)
	routes.SetUpProtectedRoute(router)
	if err := router.Run(":8080"); err != nil {
		fmt.Println("Failed to start server")
		panic(err)
	}

}
