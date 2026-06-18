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
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection("users")

func HashPassword(password string) (string, error) {
	HashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", err
	}

	return string(HashPassword), nil
}

func VerifyPassword(userPassword string, providedPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(userPassword), []byte(providedPassword))

	if err != nil {
		return false, err
	}

	return true, nil
}

func RegisterUser() gin.HandlerFunc {
	return func(c *gin.Context) {

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)

		defer cancel()

		var user models.User

		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid user data",
				"success": false,
			})
			return
		}

		validate := validator.New()

		if err := validate.Struct(user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   err.Error(),
				"success": false,
			})
			return
		}

		count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to check email",
				"success": false,
			})
			return
		}

		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "Email already exists",
				"success": false,
			})
			return
		}

		hashedPassword, err := HashPassword(user.Password)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to hash password",
				"success": false,
			})
			return
		}
		user.UserID = bson.NewObjectID().Hex()
		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()
		user.Password = hashedPassword

		result, err := userCollection.InsertOne(ctx, user)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to register user",
				"success": false,
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"message": "User created",
			"result":  result,
		})

	}
}

func LoginUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var userLogin models.UserLogin

		if err := c.ShouldBindJSON(&userLogin); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid user data",
				"success": false,
			})
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)

		defer cancel()

		var foundUser models.User

		err := userCollection.FindOne(ctx, bson.M{"email": userLogin.Email}).Decode(&foundUser)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid email or password",
				"success": false,
			})
			return
		}

		isValid, err := VerifyPassword(foundUser.Password, userLogin.Password)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid email or password",
				"success": false,
			})
			return
		}

		if !isValid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid email or password",
				"success": false,
			})
			return
		}

	}
}
