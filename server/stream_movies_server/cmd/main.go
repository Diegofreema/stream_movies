package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	router.GET("/hello", func(ctx *gin.Context) {
		fmt.Println("Hello World")
		ctx.JSON(200, gin.H{
			"message": "Hello World",
		})
	})

	if err := router.Run(":8080"); err != nil {
		fmt.Println("Failed to start server")
		panic(err)
	}

}
