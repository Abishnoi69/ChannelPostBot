package helpers

import (
	"AshokShau/channelManager/src/db"
	"errors"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

// Helper function to send media with options
func sendMedia[T any](b *gotgbot.Bot, chatID int64, fileID string, opts *T) (*gotgbot.Message, error) {
	switch v := any(opts).(type) {
	case *gotgbot.SendStickerOpts:
		return b.SendSticker(chatID, gotgbot.InputFileByID(fileID), v)
	case *gotgbot.SendDocumentOpts:
		return b.SendDocument(chatID, gotgbot.InputFileByID(fileID), v)
	case *gotgbot.SendPhotoOpts:
		return b.SendPhoto(chatID, gotgbot.InputFileByID(fileID), v)
	case *gotgbot.SendAudioOpts:
		return b.SendAudio(chatID, gotgbot.InputFileByID(fileID), v)
	case *gotgbot.SendVoiceOpts:
		return b.SendVoice(chatID, gotgbot.InputFileByID(fileID), v)
	case *gotgbot.SendVideoOpts:
		return b.SendVideo(chatID, gotgbot.InputFileByID(fileID), v)
	case *gotgbot.SendVideoNoteOpts:
		return b.SendVideoNote(chatID, gotgbot.InputFileByID(fileID), v)
	case *gotgbot.SendAnimationOpts:
		return b.SendAnimation(chatID, gotgbot.InputFileByID(fileID), v)
	default:
		return nil, errors.New("unknown media type")
	}
}

var PostEnumFuncMap = map[int]func(b *gotgbot.Bot, ctx *ext.Context, chatId int64, msg, fileID string, keyB *gotgbot.InlineKeyboardMarkup, userSetting *db.UserSettings) (*gotgbot.Message, error){
	db.TEXT: func(b *gotgbot.Bot, ctx *ext.Context, chatId int64, msg, _ string, keyB *gotgbot.InlineKeyboardMarkup, userSetting *db.UserSettings) (*gotgbot.Message, error) {
		opts := &gotgbot.SendMessageOpts{
			ParseMode:           gotgbot.ParseModeHTML,
			LinkPreviewOptions:  &gotgbot.LinkPreviewOptions{IsDisabled: userSetting.WebPreview},
			ReplyMarkup:         keyB,
			ReplyParameters:     &gotgbot.ReplyParameters{AllowSendingWithoutReply: true},
			DisableNotification: userSetting.NoNotif,
			ProtectContent:      userSetting.Protect,
		}
		return b.SendMessage(chatId, msg, opts)
	},
	db.STICKER: func(b *gotgbot.Bot, ctx *ext.Context, chatId int64, _, fileID string, keyB *gotgbot.InlineKeyboardMarkup, userSetting *db.UserSettings) (*gotgbot.Message, error) {
		opts := &gotgbot.SendStickerOpts{
			ReplyMarkup:         keyB,
			DisableNotification: userSetting.NoNotif,
			ProtectContent:      userSetting.Protect,
		}
		return sendMedia(b, chatId, fileID, opts)
	},
	db.DOCUMENT: func(b *gotgbot.Bot, ctx *ext.Context, chatId int64, msg, fileID string, keyB *gotgbot.InlineKeyboardMarkup, userSetting *db.UserSettings) (*gotgbot.Message, error) {
		opts := &gotgbot.SendDocumentOpts{
			ParseMode:           gotgbot.ParseModeHTML,
			ReplyMarkup:         keyB,
			Caption:             msg,
			DisableNotification: userSetting.NoNotif,
			ProtectContent:      userSetting.Protect,
		}
		return sendMedia(b, chatId, fileID, opts)
	},
	db.PHOTO: func(b *gotgbot.Bot, ctx *ext.Context, chatId int64, msg, fileID string, keyB *gotgbot.InlineKeyboardMarkup, userSetting *db.UserSettings) (*gotgbot.Message, error) {
		opts := &gotgbot.SendPhotoOpts{
			ParseMode:             gotgbot.ParseModeHTML,
			ReplyMarkup:           keyB,
			Caption:               msg,
			DisableNotification:   userSetting.NoNotif,
			ProtectContent:        userSetting.Protect,
			ShowCaptionAboveMedia: userSetting.CaptionAbove,
			HasSpoiler:            userSetting.Spoiler,
		}
		return sendMedia(b, chatId, fileID, opts)
	},
	db.AUDIO: func(b *gotgbot.Bot, ctx *ext.Context, chatId int64, msg, fileID string, keyB *gotgbot.InlineKeyboardMarkup, userSetting *db.UserSettings) (*gotgbot.Message, error) {
		opts := &gotgbot.SendAudioOpts{
			ParseMode:           gotgbot.ParseModeHTML,
			ReplyMarkup:         keyB,
			Caption:             msg,
			DisableNotification: userSetting.NoNotif,
			ProtectContent:      userSetting.Protect,
		}
		return sendMedia(b, chatId, fileID, opts)
	},
	db.VOICE: func(b *gotgbot.Bot, ctx *ext.Context, chatId int64, msg, fileID string, keyB *gotgbot.InlineKeyboardMarkup, userSetting *db.UserSettings) (*gotgbot.Message, error) {
		opts := &gotgbot.SendVoiceOpts{
			ParseMode:           gotgbot.ParseModeHTML,
			ReplyMarkup:         keyB,
			Caption:             msg,
			DisableNotification: userSetting.NoNotif,
			ProtectContent:      userSetting.Protect,
		}
		return sendMedia(b, chatId, fileID, opts)
	},
	db.VIDEO: func(b *gotgbot.Bot, ctx *ext.Context, chatId int64, msg, fileID string, keyB *gotgbot.InlineKeyboardMarkup, userSetting *db.UserSettings) (*gotgbot.Message, error) {
		opts := &gotgbot.SendVideoOpts{
			ParseMode:             gotgbot.ParseModeHTML,
			ReplyMarkup:           keyB,
			Caption:               msg,
			DisableNotification:   userSetting.NoNotif,
			ProtectContent:        userSetting.Protect,
			HasSpoiler:            userSetting.Spoiler,
			ShowCaptionAboveMedia: userSetting.CaptionAbove,
		}
		return sendMedia(b, chatId, fileID, opts)
	},

	db.VideoNote: func(b *gotgbot.Bot, ctx *ext.Context, chatId int64, _, fileID string, keyB *gotgbot.InlineKeyboardMarkup, userSetting *db.UserSettings) (*gotgbot.Message, error) {
		opts := &gotgbot.SendVideoNoteOpts{
			ReplyMarkup:         keyB,
			DisableNotification: userSetting.NoNotif,
			ProtectContent:      userSetting.Protect,
		}
		return sendMedia(b, chatId, fileID, opts)
	},
}
