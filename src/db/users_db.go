package db

import (
	"errors"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserSettings struct {
	UserId       int64 `bson:"_id,omitempty" json:"_id,omitempty"`
	NoNotif      bool  `bson:"nonotif,omitempty" json:"nonotif,omitempty"`
	Protect      bool  `bson:"protect,omitempty" json:"protect,omitempty"`
	Spoiler      bool  `bson:"spoiler,omitempty" json:"spoiler,omitempty"`
	WebPreview   bool  `bson:"webpreview,omitempty" json:"webpreview,omitempty"`
	CaptionAbove bool  `bson:"captionabove,omitempty" json:"captionabove,omitempty"`
	ForwardTag   bool  `bson:"forwardtag,omitempty" json:"forwardtag,omitempty"`
}

// GetUserSettings retrieves a user's settings or initializes defaults if not found.
func GetUserSettings(userID int64) *UserSettings {
	settings := &UserSettings{
		UserId:       userID,
		NoNotif:      false,
		Protect:      false,
		Spoiler:      false,
		WebPreview:   false,
		CaptionAbove: false,
		ForwardTag:   false,
	}
	if err := findOne(usersColl, bson.M{"_id": userID}).Decode(settings); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			_ = updateOne(usersColl, bson.M{"_id": userID}, settings)
			return settings // Return default if no document found
		}
		log.Printf("[Database] GetUserSettings: %d - %v", userID, err)
	}
	return settings
}

// UpdateNoNotif updates the "NoNotif" setting.
func UpdateNoNotif(userID int64, value bool) {
	updateUserSetting(userID, "nonotif", value)
}

// UpdateProtect updates the "Protect" setting.
func UpdateProtect(userID int64, value bool) {
	updateUserSetting(userID, "protect", value)
}

// UpdateSpoiler updates the "Spoiler" setting.
func UpdateSpoiler(userID int64, value bool) {
	updateUserSetting(userID, "spoiler", value)
}

// UpdateWebPreview updates the "WebPreview" setting.
func UpdateWebPreview(userID int64, value bool) {
	updateUserSetting(userID, "webpreview", value)
}

// UpdateCaptionAbove updates the "CaptionAbove" setting.
func UpdateCaptionAbove(userID int64, value bool) {
	updateUserSetting(userID, "captionabove", value)
}

// UpdateForwardTag updates the "ForwardTag" setting.
func UpdateForwardTag(userID int64, value bool) {
	updateUserSetting(userID, "forwardtag", value)
}

// updateUserSetting updates a specific field for a user's settings.
func updateUserSetting(userID int64, field string, value bool) {
	update := bson.M{"$set": bson.M{field: value}}
	_, err := usersColl.UpdateOne(ctx, bson.M{"_id": userID}, update)
	if err != nil {
		log.Printf("[Database] updateUserSetting: %s - %v - %d", field, err, userID)
	}
}

// ResetUserSettings resets all settings for a user to their default values.
func ResetUserSettings(userID int64) {
	defaultSettings := UserSettings{
		UserId:       userID,
		NoNotif:      false,
		Protect:      false,
		Spoiler:      false,
		WebPreview:   false,
		CaptionAbove: false,
		ForwardTag:   false,
	}
	if err := updateOne(usersColl, bson.M{"_id": userID}, defaultSettings); err != nil {
		log.Printf("[Database] ResetUserSettings: %v - %d", err, userID)
	}
}

func GetAllUsers() ([]int64, error) {
	var userIDs []int64
	projection := options.Find().SetProjection(bson.M{"_id": 1})
	cursor, err := usersColl.Find(ctx, bson.M{}, projection)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var result struct {
			UserId int64 `bson:"_id"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, result.UserId)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return userIDs, nil
}
