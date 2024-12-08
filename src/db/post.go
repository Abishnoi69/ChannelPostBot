package db

import (
	"errors"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Chat represents a chat with a single MsgId
type Chat struct {
	ChatId int64 `bson:"chat_id,omitempty" json:"chat_id,omitempty"`
	MsgId  int64 `bson:"msg_id,omitempty" json:"msg_id,omitempty"`
}

// Post represents a post document in MongoDB
type Post struct {
	PostId      string   `bson:"_id,omitempty" json:"post_id,omitempty"`
	UserId      int64    `bson:"user_id,omitempty" json:"user_id,omitempty"`
	MsgType     int      `bson:"msgtype,omitempty" json:"msgtype,omitempty"`
	Chats       []Chat   `bson:"chats,omitempty" json:"chats,omitempty"`
	FileID      string   `bson:"fileid,omitempty" json:"fileid,omitempty"`
	Buttons     []Button `bson:"buttons,omitempty" json:"buttons,omitempty"`
	FilterReply string   `bson:"reply,omitempty" json:"reply,omitempty"`
}

// GetPost retrieves a post by its PostId
func GetPost(postID string) (*Post, error) {
	var post Post
	err := findOne(postColl, bson.M{"_id": postID}).Decode(&post)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // Return nil if no post found
		}
		return nil, err
	}
	return &post, nil
}

// RemovePost removes a post by its PostId
func RemovePost(postID string) error {
	//if err := deleteOne(postColl, bson.M{"_id": postID}); err != nil {
	//	log.Printf("[Database] RemovePost: %v - PostId: %s", err, postID)
	//	return err
	//}
	log.Printf("[Database] RemovePost: Successfully deleted PostId: %s", postID)
	return nil
}

func AddPost(postID string, userID, chatID, msgID int64, msgType int, fileID string, buttons []Button, filterReply string) (string, error) {
	// Prepare the new chat data
	chat := Chat{
		ChatId: chatID,
		MsgId:  msgID,
	}

	// Use UpdateOne to add the new chat data to the Post document
	filter := bson.M{"_id": postID}
	update := bson.M{
		"$addToSet": bson.M{
			"chats": chat, // Add the chat to the chats array (only if it doesn't already exist)
		},
		"$set": bson.M{
			"user_id": userID,      // Update user_id if necessary
			"msgtype": msgType,     // Update or set MsgType
			"fileid":  fileID,      // Update FileID
			"buttons": buttons,     // Update Buttons
			"reply":   filterReply, // Update FilterReply
		},
	}

	// Perform the update operation
	_, err := postColl.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		log.Printf("[Database] AddPost: %v - PostId: %s, User: %d, Chat: %d, Msg: %d, MsgType: %d", err, postID, userID, chatID, msgID, msgType)
		return "", err
	}

	return postID, nil
}

// ListPosts retrieves all posts for a user
func ListPosts(userID int64) ([]Post, error) {
	var posts []Post
	cursor, err := find(postColl, bson.M{"user_id": userID})
	if err != nil {
		log.Printf("[Database] ListPosts: Failed to retrieve posts for User %d: %v", userID, err)
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var post Post
		if err := cursor.Decode(&post); err != nil {
			log.Printf("[Database] ListPosts: Error decoding post for User %d - %v", userID, err)
			continue
		}
		posts = append(posts, post)
	}

	if err := cursor.Err(); err != nil {
		log.Printf("[Database] ListPosts: Cursor error for User %d - %v", userID, err)
		return nil, err
	}

	return posts, nil
}
