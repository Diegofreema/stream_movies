package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetMovies() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "Movie route",
		})
	}
}
