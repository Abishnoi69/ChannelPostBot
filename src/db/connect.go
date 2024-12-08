package db

import (
	"errors"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Connections represents a user's connections to chats.
type Connections struct {
	UserId  int64   `bson:"_id,omitempty" json:"_id,omitempty"`
	ChatIds []int64 `bson:"chat_ids,omitempty" json:"chat_ids,omitempty"`
}

// GetUserConnectionSetting retrieves a user's connection settings or initializes defaults if not found.
func getUserConnectionSetting(userID int64) *Connections {
	connectionSrc := &Connections{UserId: userID, ChatIds: []int64{}}
	if err := findOne(connectionColl, bson.M{"_id": userID}).Decode(&connectionSrc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return connectionSrc // Return default if no document found
		}
		log.Printf("[Database] GetUserConnectionSetting: %d - %v", userID, err)
	}
	return connectionSrc
}

// Connection fetches the connection details for a user.
func Connection(userID int64) *Connections {
	return getUserConnectionSetting(userID)
}

// ConnectId adds a chat ID to a user's connection list.
func ConnectId(userID, chatID int64) {
	connectionUpdate := Connection(userID)

	// Add chatID to the list if not already present
	for _, id := range connectionUpdate.ChatIds {
		if id == chatID {
			log.Printf("[Database] ConnectId: Chat ID %d already connected for User %d", chatID, userID)
			return
		}
	}

	connectionUpdate.ChatIds = append(connectionUpdate.ChatIds, chatID)

	if err := updateOne(connectionColl, bson.M{"_id": userID}, connectionUpdate); err != nil {
		log.Printf("[Database] ConnectId: %v - %d", err, userID)
	}
}
func DisconnectId(userID, chatID int64) {
	connection := Connection(userID)

	// Filter out the chat ID
	updatedChatIds := make([]int64, 0, len(connection.ChatIds))
	for _, id := range connection.ChatIds {
		if id != chatID {
			updatedChatIds = append(updatedChatIds, id)
		}
	}

	if len(updatedChatIds) == len(connection.ChatIds) {
		log.Printf("[Database] DisconnectId: Chat ID %d not found for User %d", chatID, userID)
		return
	}

	// Prepare the update document
	filter := bson.M{"_id": userID}
	update := bson.M{}
	if len(updatedChatIds) == 0 {
		// Remove the field if the array is empty
		update = bson.M{"$unset": bson.M{"chat_ids": ""}}
	} else {
		// Update the array otherwise
		update = bson.M{"$set": bson.M{"chat_ids": updatedChatIds}}
	}

	// Apply the update
	if _, err := connectionColl.UpdateOne(ctx, filter, update); err != nil {
		log.Printf("[Database] DisconnectId: Failed to update DB - %v", err)
	} else {
		log.Printf("[Database] DisconnectId: Successfully updated DB for User %d", userID)
	}
}

// DisconnectAll removes all chat IDs from the user's connection list.
func DisconnectAll(userID int64) {
	connectionUpdate := Connection(userID)
	connectionUpdate.ChatIds = []int64{}

	if err := updateOne(connectionColl, bson.M{"_id": userID}, connectionUpdate); err != nil {
		log.Printf("[Database] DisconnectAll: %v - %d", err, userID)
	}
}
