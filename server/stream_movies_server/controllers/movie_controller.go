package controllers

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/DiegoFreema/stream_movies/database"
	"github.com/DiegoFreema/stream_movies/models"
	"github.com/DiegoFreema/stream_movies/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var moviesCollection *mongo.Collection = database.OpenCollection(database.MoviesCollection)
var rankingCollection *mongo.Collection = database.OpenCollection(database.RankingCollection)
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

func AdminReviewUpdate() gin.HandlerFunc {
	return func(c *gin.Context) {

		userRole, err := utils.GetUserRoleFromContext(c)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   err.Error(),
				"success": false,
			})
			return
		}

		if userRole != "Admin" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized, only admin can access this route",
				"success": false,
			})
			return
		}
		movieId := c.Param("imdb_id")

		if movieId == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid movie id",
				"success": false,
			})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var req struct {
			AdminReview string `json:"admin_review"`
		}

		var res struct {
			RankingName string `json:"ranking_name"`
			AdminReview string `json:"admin_review"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request data",
				"success": false,
			})
			return
		}

		filter := bson.M{"imdb_id": movieId}
		update := bson.M{"$set": bson.M{"admin_review": req.AdminReview, "ranking": bson.M{
			"ranking_value": 5,
			"ranking_name":  "Terrible",
		}}}

		result, err := moviesCollection.UpdateOne(ctx, filter, update)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to update movie",
				"success": false,
			})
			return
		}

		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Movie not found",
				"success": false,
			})
			return
		}

		res.AdminReview = "Terrible"
		res.RankingName = "Terrible"

		c.JSON(http.StatusOK, gin.H{
			"movie":   res,
			"success": true,
		})

	}
}

func GetReviewRanking(admin_review string) (string, int, error) {
	rankings, err := getRankings()

	if err != nil {
		return "", 0, err
	}

	sentimentDelimited := ""

	for _, ranking := range rankings {
		if ranking.RankingValue != 999 {
			sentimentDelimited = sentimentDelimited + ranking.RankingName + ","
		}
	}

	sentimentDelimited = strings.Trim(sentimentDelimited, ",")

	return sentimentDelimited, 0, nil

}

func getRankings() ([]models.Ranking, error) {
	var rankings []models.Ranking

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	cursor, err := rankingCollection.Find(ctx, bson.M{})

	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &rankings); err != nil {
		return nil, err
	}

	return rankings, nil

}

func GetRecommendedMovies() gin.HandlerFunc {
	return func(c *gin.Context) {

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)

		defer cancel()

		userId, err := utils.GetUserIdFromContext(c)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   err.Error(),
				"success": false,
			})
			return
		}

		favoriteGenres, err := GetUsersFavoriteGenres(userId)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to get favorite genres",
				"success": false,
			})
			return
		}

		var recommendedMovieLimit int64 = 5

		findOptions := options.Find().SetLimit(recommendedMovieLimit).SetSort(bson.D{{Key: "ranking.ranking_value", Value: 1}})

		filter := bson.M{"genre.genre_name": bson.M{"$in": favoriteGenres}}

		cursor, err := moviesCollection.Find(ctx, filter, findOptions)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to get recommended movies",
				"success": false,
			})
			return
		}

		defer cursor.Close(ctx)

		var recommendedMovies []models.Movie

		if err := cursor.All(ctx, &recommendedMovies); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to get recommended movies",
				"success": false,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"movies":         recommendedMovies,
			"success":        true,
			"favoriteGenres": favoriteGenres,
		})
	}
}

func GetUsersFavoriteGenres(userId string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	filter := bson.M{"user_id": userId}
	projection := bson.M{"favorite_genres.genre_name": 1, "_id": 0}

	var result bson.M
	opts := options.FindOne().SetProjection(projection)

	err := userCollection.FindOne(ctx, filter, opts).Decode(&result)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return []string{}, nil
		}
		return []string{}, err
	}

	favGenresArray, ok := result["favorite_genres"].(bson.A)

	if !ok {
		return []string{}, errors.New("invalid favorite genres type")
	}

	var genreNames []string

	for _, genre := range favGenresArray {
		if genreMap, ok := genre.(bson.D); ok {
			for _, elem := range genreMap {
				if elem.Key == "genre_name" {
					if name, ok := elem.Value.(string); ok {
						genreNames = append(genreNames, name)
					}
				}
			}
		}

	}

	return genreNames, nil

}
