package utils

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/DiegoFreema/stream_movies/database"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type SignedDetails struct {
	Email     string
	Role      string
	FirstName string
	LastName  string
	UserId    string
	jwt.RegisteredClaims
}

var userCollection *mongo.Collection = database.OpenCollection(database.UsersCollection)

var SECRET_KEY string = os.Getenv("SECRET_KEY")
var REFRESH_SECRET_KEY string = os.Getenv("REFRESH_SECRET_KEY")

func GenerateAllTokens(email, firstName, lastName, role, userId string) (string, string, error) {
	claims := &SignedDetails{
		Email:     email,
		Role:      role,
		FirstName: firstName,
		LastName:  lastName,
		UserId:    userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Local().Add(time.Hour * time.Duration(24))),
			Issuer:    "magic_stream",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(SECRET_KEY))

	if err != nil {
		return "", "", err
	}

	refreshClaims := &SignedDetails{
		Email:     email,
		Role:      role,
		FirstName: firstName,
		LastName:  lastName,
		UserId:    userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Local().Add(time.Hour * time.Duration(24) * 7)),
			Issuer:    "magic_stream",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)

	refreshedSignedToken, err := refreshToken.SignedString([]byte(REFRESH_SECRET_KEY))

	if err != nil {
		return "", "", err
	}

	return signedToken, refreshedSignedToken, nil

}

func UpdateAllTokens(userId, token, refreshToken string) (err error) {

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	defer cancel()

	updatedAt, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

	updateData := bson.M{
		"$set": bson.M{
			"token":         token,
			"refresh_token": refreshToken,
			"updated_at":    updatedAt,
		},
	}

	_, err = userCollection.UpdateOne(ctx, bson.M{
		"user_id": userId,
	}, updateData)

	if err != nil {
		return err
	}

	return nil
}

func GetAccessToken(c *gin.Context) (string, error) {

	authHeader := c.Request.Header.Get("Authorization")

	if authHeader == "" {
		return "", errors.New("Authorization header is required")
	}

	tokenString := authHeader[len("Bearer "):]

	if tokenString == "" {
		return "", errors.New("Token is required")
	}

	return tokenString, nil
}

func ValidateToken(signedToken string) (*SignedDetails, error) {
	claims := &SignedDetails{}

	token, err := jwt.ParseWithClaims(signedToken, claims, func(t *jwt.Token) (any, error) {
		return []byte(SECRET_KEY), nil
	})

	if err != nil {
		return nil, err
	}

	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, err
	}

	if claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("token has expired")
	}

	return claims, nil
}

func GetUserIdFromContext(c *gin.Context) (string, error) {
	userId, exists := c.Get("user_id")

	if !exists {
		return "", errors.New("userId not found in context")
	}

	id, ok := userId.(string)

	if !ok {
		return "", errors.New("invalid userId type")
	}
	return id, nil
}
func GetUserRoleFromContext(c *gin.Context) (string, error) {
	role, exists := c.Get("role")

	if !exists {
		return "", errors.New("Role could not be found in context")
	}

	userRole, ok := role.(string)

	if !ok {
		return "", errors.New("Unauthorized, only admin can access this route")
	}

	return userRole, nil
}
