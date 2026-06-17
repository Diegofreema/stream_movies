package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func loadEnvFile() {
	paths := []string{
		".env",
		filepath.Join("..", ".env"),
	}

	for _, path := range paths {
		if err := godotenv.Load(path); err == nil {
			return
		}
	}

	log.Println("Error loading .env file: tried .env and ../.env")
}

func DBInstance() *mongo.Client {
	loadEnvFile()

	MongoDb := os.Getenv("MONGODB_URI")

	if MongoDb == "" {
		log.Fatal("Mongo db uri not set")

	}

	fmt.Println(MongoDb)

	clientOptions := options.Client().ApplyURI(MongoDb)

	client, err := mongo.Connect(clientOptions)

	if err != nil {
		return nil
	}

	return client
}

var Client *mongo.Client = DBInstance()

func OpenCollection(collectionName string) *mongo.Collection {
	loadEnvFile()

	databaseName := os.Getenv("DATABASE_NAME")

	fmt.Println(databaseName)

	collection := Client.Database(databaseName).Collection(collectionName)

	if collection == nil {
		return nil
	}

	return collection
}
