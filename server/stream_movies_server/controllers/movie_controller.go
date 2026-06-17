package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/DiegoFreema/stream_movies/database"
	"github.com/DiegoFreema/stream_movies/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var moviesCollection *mongo.Collection = database.OpenCollection("movies")
var validate = validator.New()

func GetMovies() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)

		defer cancel()

		var movies []models.Movie
		cursor, err := moviesCollection.Find(ctx, bson.M{})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to fetch movies",
				"success": false,
			})
			return
		}

		if err = cursor.All(ctx, &movies); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to fetch movies",
				"success": false,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"movies":  movies,
			"success": true,
		})
	}
}

func GetMovie() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		movieId := c.Param("imdb_id")

		if movieId == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid movie id",
				"success": false,
			})
			return
		}

		var movie models.Movie

		err := moviesCollection.FindOne(ctx, bson.M{"imdb_id": movieId}).Decode(&movie)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Movie not found",
				"success": false,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"movie":   movie,
			"success": true,
		})

	}
}

func AddMovie() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var movie models.Movie

		if err := c.ShouldBindJSON(&movie); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid movie data",
				"success": false,
			})
			return
		}

		if err := validate.Struct(movie); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   err.Error(),
				"success": false,
			})
			return
		}

		result, err := moviesCollection.InsertOne(ctx, movie)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to add movie",
				"success": false,
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"movie":   result,
			"success": true,
		})

	}
}
