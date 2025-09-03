package Database

import (
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"os"
)

func DB() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
	mongoURL := os.Getenv("MONGO_URL")
	_, err = mongo.Connect(options.Client().
		ApplyURI(mongoURL))
	if err != nil {
		panic(err)
	}
}
