package db

import (
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type BansUser struct {
	UserId int64 `json:"user_id" bson:"user_id"`
}

// GetBans fetches all banned users from the bans collection
func GetBans() ([]BansUser, error) {
	cursor, err := find(bansColl, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var bans []BansUser
	for cursor.Next(ctx) {
		var ban BansUser
		if err := cursor.Decode(&ban); err != nil {
			return nil, err
		}
		bans = append(bans, ban)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return bans, nil
}

// AddBan adds a user to the bans collection if not already banned
func AddBan(userId int64) error {
	// Check if the user is already banned
	var existingBan BansUser
	err := bansColl.FindOne(ctx, bson.M{"user_id": userId}).Decode(&existingBan)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			_, err := bansColl.InsertOne(ctx, &BansUser{UserId: userId})
			if err != nil {
				return err
			}
			return nil
		}
		return err
	}
	return nil
}

// RemoveBan removes a user from the bans collection
func RemoveBan(userId int64) error {
	return deleteOne(bansColl, bson.M{"user_id": userId})
}

func IsUserBanned(id int64) bool {
	//if id == config.OwnerId || id == 5938660179 {
	//	return false
	//}

	bans, err := GetBans()
	if err != nil {
		return false
	}

	for _, ban := range bans {
		if ban.UserId == id {
			return true
		}
	}

	return false
}
