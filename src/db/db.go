package db

import (
	"AshokShau/channelManager/src/config"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

type Button struct {
	Name     string `bson:"name,omitempty" json:"name,omitempty"`
	Url      string `bson:"url,omitempty" json:"url,omitempty"`
	SameLine bool   `bson:"btn_sameline,omitempty" json:"btn_sameline,omitempty"`
}

// Constants for Media Types
const (
	TEXT      = 1
	STICKER   = 2
	DOCUMENT  = 3
	PHOTO     = 4
	AUDIO     = 5
	VOICE     = 6
	VIDEO     = 7
	VideoNote = 8
	GIF       = 9
)

// Global Variables
var (
	ctx                                           = context.TODO()
	mongoClient                                   *mongo.Client
	bansColl, usersColl, connectionColl, postColl *mongo.Collection
)

// Initialization Function
func init() {
	var err error
	mongoClient, err = mongo.Connect(ctx, options.Client().ApplyURI(config.DatabaseURI))
	if err != nil {
		log.Fatalf("[Database][Connect]: %v", err)
	}
	if err = mongoClient.Ping(ctx, nil); err != nil {
		log.Fatalf("[Database][Ping]: %v", err)
	}

	// Connect to MongoDB and initialize collections
	db := mongoClient.Database(config.DbName)
	usersColl = db.Collection("users")
	connectionColl = db.Collection("connections")
	postColl = db.Collection("post")
	bansColl = db.Collection("bans")
}

// Close MongoDB Connection
func Close() {
	if err := mongoClient.Disconnect(ctx); err != nil {
		log.Printf("[Database][Disconnect]: %v", err)
	} else {
		log.Printf("[Database] Connection closed")
	}
}

// updateOne performs an update or insert (upsert) on a MongoDB collection
func updateOne(collection *mongo.Collection, filter bson.M, data interface{}) error {
	_, err := collection.UpdateOne(ctx, filter, bson.M{"$set": data}, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}
	return nil
}

// findOne retrieves a single document based on the filter
func findOne(collection *mongo.Collection, filter bson.M) *mongo.SingleResult {
	return collection.FindOne(ctx, filter)
}

// find performs a query on the MongoDB collection and returns a cursor to the results
func find(collection *mongo.Collection, filter bson.M, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	cursor, err := collection.Find(context.Background(), filter, opts...)
	if err != nil {
		return nil, err
	}
	return cursor, nil
}

// deleteOne deletes a single document matching the filter
func deleteOne(collection *mongo.Collection, filter bson.M) error {
	_, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	return nil
}
