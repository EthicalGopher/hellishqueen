package Database

import (
	"context"
	"errors"
	"fmt"
	"hellish/crypto"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	ServerId        string  `bson:"server_id"`
	ServerData      string  `bson:"server_data"`
	ApiList         ApiList `bson:"apilist"`
	ActivateChannel string  `bson:"activate_channel"`
	SystemMessage   string  `bson:"system_message"`
}
type ApiList struct {
	Apikeys []string `bson:"apikeys"`
}

var collection *mongo.Collection
var client *mongo.Client

func ConnectDB() error {
	mongoURL := os.Getenv("MONGO_URL")
	if mongoURL == "" {
		return fmt.Errorf("MONGO_URL environment variable not set")
	}
	clientOptions := options.Client().ApplyURI(mongoURL)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("could not connect to MongoDB: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("could not ping MongoDB: %w", err)
	}

	collection = client.Database("Hellish").Collection("users")
	log.Println("Successfully connected to MongoDB!")
	return nil
}

// DisconnectDB should be called when your application is shutting down.
func DisconnectDB() {
	if client != nil {
		log.Println("Disconnecting from MongoDB...")
		if err := client.Disconnect(context.Background()); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		}
	}
}

// FindChannel finds the activated channel for a given server ID.
func FindChannel(serverID string) (string, error) {
	if collection == nil {
		return "", fmt.Errorf("database not initialized, call ConnectDB first")
	}

	// Create a new context for this specific operation.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result User

	err := collection.FindOne(ctx, bson.M{"server_id": serverID}).Decode(&result)
	if err != nil {
		// Handle the case where no document was found.
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", fmt.Errorf("no configuration found for server ID: %s", serverID)
		}
		// Handle other potential errors.
		return "", fmt.Errorf("error finding document: %w", err)
	}

	if result.ActivateChannel == "" {
		return "", fmt.Errorf("found server config, but activate_channel is not set")
	}

	return result.ActivateChannel, nil
}
func InsertChannel(serverId string, channelId string) error {
	if collection == nil {
		return fmt.Errorf("database not initialized, call ConnectDB first")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.D{{"server_id", serverId}}
	update := bson.M{"$set": bson.M{"activate_channel": channelId}}
	opts := options.Update().SetUpsert(true)
	result, err1 := collection.UpdateOne(ctx, filter, update, opts)
	fmt.Println(result)
	if err1 != nil {
		return err1
	}
	return nil

}

// AddAPIKey adds a new API key to a server's list using an upsert operation.
func AddAPIKey(serverId string, apiKey string) error {
	if collection == nil {
		return fmt.Errorf("database not initialized")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	encryptedKey, err := crypto.Encrypt(apiKey)
	if err != nil {
		return err
	}
	filter := bson.M{"server_id": serverId}
	// Use $addToSet to add the new key to the apikeys array only if it's not already present.
	update := bson.M{
		"$addToSet": bson.M{"apilist.apikeys": encryptedKey},
		"$setOnInsert": bson.M{ // Set default fields only if the document is new
			"server_id":        serverId,
			"activate_channel": "",
			"server_data":      "",
			"system_message":   "",
		},
	}
	opts := options.Update().SetUpsert(true)

	_, err = collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to add API key: %w", err)
	}
	return nil
}

// ViewAPIKeys retrieves all API keys for a given server.
func ViewAPIKeys(serverId string) ([]string, error) {
	if collection == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result User
	filter := bson.M{"server_id": serverId}
	err := collection.FindOne(ctx, filter).Decode(&result)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("error finding server config: %w", err)
	}

	// Ensure we don't return a nil slice
	if result.ApiList.Apikeys == nil {
		return []string{}, nil
	}

	return result.ApiList.Apikeys, nil
}

// RemoveAPIKey removes a specific API key from a server's list.
func RemoveAPIKey(serverId string, apiKey string) error {
	if collection == nil {
		return fmt.Errorf("database not initialized")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"server_id": serverId}
	encrptedKey, err := crypto.Encrypt(apiKey)
	if err != nil {
		return err
	}
	update := bson.M{"$pull": bson.M{"apilist.apikeys": encrptedKey}}

	res, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to remove API key: %w", err)
	}

	if res.ModifiedCount == 0 {
		return fmt.Errorf("API key not found on this server")
	}

	return nil
}

// ClearAPIKeys removes all API keys for a server by setting the array to empty.
func ClearAPIKeys(serverId string) error {
	if collection == nil {
		return fmt.Errorf("database not initialized")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"server_id": serverId}
	update := bson.M{"$set": bson.M{"apilist.apikeys": []string{}}}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to clear API keys: %w", err)
	}
	return nil
}

func InsertSystemMessage(serverId string, message string) error {
	if collection == nil {
		return fmt.Errorf("database not initialized")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"server_id": serverId}
	// Use $set to update the message, and $setOnInsert to create default fields if the document is new.
	// This is a single, atomic, and efficient database operation.
	update := bson.M{
		"$set": bson.M{"system_message": message},
		"$setOnInsert": bson.M{
			"server_id":        serverId,
			"activate_channel": "",
			"server_data":      "",
			"apilist":          ApiList{Apikeys: []string{}},
		},
	}
	opts := options.Update().SetUpsert(true)

	_, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to upsert system message: %w", err)
	}
	return nil
}

func ViewSystemMessage(serverId string) (string, error) {
	if collection == nil {
		return "", fmt.Errorf("database not initialized")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result User
	filter := bson.M{"server_id": serverId}
	err := collection.FindOne(ctx, filter).Decode(&result)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", nil
		}
		return "", fmt.Errorf("error finding system message: %w", err)
	}

	return result.SystemMessage, nil
}
