package modules

import (
	"AshokShau/channelManager/src/db"
	"AshokShau/channelManager/src/modules/utils/helpers"
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"log"
	"strings"
	"time"
)

func repost(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	if msg.Chat.Type != "private" {
		return nil
	}

	chatIds := isConnected(b, ctx, msg.From.Id)
	if chatIds == nil {
		return nil
	}

	reply := msg.ReplyToMessage
	if reply == nil {
		_, _ = msg.Reply(b, "Please reply to a message to re-post.", helpers.Shtml())
		return nil
	}

	args := ctx.Args()[1:]
	if len(args) < 1 {
		_, err := msg.Reply(b, "Please provide a post ID to re-post.\nUsage: <code>!repost post_id</code>", helpers.Shtml())
		return err
	}

	post, err := db.GetPost(args[0])
	if err != nil || post == nil {
		_, _ = msg.Reply(b, "Post not found or an error occurred while retrieving the post.", helpers.Shtml())
		return err
	}

	message, err := msg.Reply(b, "📤 Reposting post to connected chats...\nThis may take some time.", helpers.Shtml())
	if err != nil {
		return err
	}

	// Delete the old post
	deletedCount := 0
	for _, chat := range post.Chats {
		_, err = b.DeleteMessage(chat.ChatId, chat.MsgId, nil)
		if err != nil {
			continue
		}
		deletedCount++
		time.Sleep(200 * time.Millisecond)
	}

	_ = db.RemovePost(args[0])

	// Send the new post
	postId := helpers.GenerateUniqueString()
	postText, dataType, fileId, buttons, errorMsg := helpers.GetMsgType(msg)
	if dataType == -1 {
		_, _, _ = message.EditText(b, errorMsg, nil)
		return nil
	}

	keyboard := gotgbot.InlineKeyboardMarkup{InlineKeyboard: helpers.BuildKeyboard(buttons)}
	if keyboard.InlineKeyboard == nil {
		keyboard.InlineKeyboard = make([][]gotgbot.InlineKeyboardButton, 0)
	}

	var successChats []int64
	var failedChats []int64
	userSettins := db.GetUserSettings(msg.From.Id)

	for _, chatId := range chatIds {
		message, err := b.CopyMessage(chatId, reply.Chat.Id, reply.MessageId, &gotgbot.CopyMessageOpts{
			ParseMode:             "HTML",
			ReplyMarkup:           &keyboard,
			Caption:               &postText,
			ProtectContent:        userSettins.Protect,
			ShowCaptionAboveMedia: userSettins.CaptionAbove,
			DisableNotification:   userSettins.NoNotif,
		})

		if err != nil {
			failedChats = append(failedChats, chatId)
			continue
		}

		successChats = append(successChats, chatId)
		_, _ = db.AddPost(postId, msg.From.Id, chatId, message.MessageId, dataType, fileId, buttons, postText)
		time.Sleep(100 * time.Millisecond)
	}

	// Prepare summary
	text := fmt.Sprintf("✅ Re-post initiated.\nOld Post Deleted from %d chats.\n", deletedCount)
	text += fmt.Sprintf("New PostId: <code>%s</code>\n\n", postId)

	if len(successChats) > 0 {
		text += "✅ Sent to the following chats:\n"
		for _, chatId := range successChats {
			messageLink := helpers.GetMessageLink(chatId, reply.MessageId)
			text += fmt.Sprintf("- <code>%d</code> (<a href='%s'>View</a>)\n", chatId, messageLink)
		}
		text += "\n"
	}

	if len(failedChats) > 0 {
		text += "❌ Failed to send to the following chats:\n"
		for _, chatId := range failedChats {
			text += fmt.Sprintf("- <code>%d</code>\n", chatId)
		}
		text += "\n"
	}

	text += fmt.Sprintf("If you want to delete this post and send another one, use <code>!repost %s</code>", postId)
	_, _, _ = message.EditText(b, text, &gotgbot.EditMessageTextOpts{
		ParseMode:   "HTML",
		ReplyMarkup: helpers.PostButton(postId),
	})
	return nil
}

func editPost(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	if msg.Chat.Type != "private" {
		return nil
	}

	chatIds := isConnected(b, ctx, msg.From.Id)
	if chatIds == nil {
		return nil
	}

	args := ctx.Args()[1:]
	if len(args) < 1 {
		_, err := msg.Reply(b, "Please provide a post ID to edit.\nUsage: <code>!edit post_id</code>", helpers.Shtml())
		return err
	}

	post, err := db.GetPost(args[0])
	if err != nil || post == nil {
		_, _ = msg.Reply(b, "Post not found or an error occurred while retrieving the post.", helpers.Shtml())
		return err
	}

	reply := msg.ReplyToMessage
	if reply == nil {
		_, _ = msg.Reply(b, "To edit a post, you need to reply to a message that you want to share with all connected chats.", helpers.Shtml())
		return nil
	}

	postText, dataType, fileId, buttons, errorMsg := helpers.GetMsgType(msg)
	log.Printf("buttons: %v", buttons)
	if dataType == -1 {
		_, _ = msg.Reply(b, errorMsg, helpers.Shtml())
		return nil
	}

	keyboard := gotgbot.InlineKeyboardMarkup{InlineKeyboard: helpers.BuildKeyboard(buttons)}
	if keyboard.InlineKeyboard == nil {
		keyboard.InlineKeyboard = make([][]gotgbot.InlineKeyboardButton, 0)
	}
	message, err := msg.Reply(b, "Please wait while the post is being edited...", helpers.Shtml())
	if err != nil {
		return err
	}

	userSetting := db.GetUserSettings(msg.From.Id)
	media := gotgbot.InputMedia(nil)
	mediaText := postText
	if mediaText == "" {
		mediaText = "."
	}

	switch dataType {
	case db.PHOTO:
		media = gotgbot.InputMediaPhoto{
			Media:                 gotgbot.InputFileByID(fileId),
			Caption:               mediaText,
			ParseMode:             "HTML",
			ShowCaptionAboveMedia: userSetting.CaptionAbove,
			HasSpoiler:            userSetting.Spoiler,
		}
	case db.GIF:
		media = gotgbot.InputMediaAnimation{
			Media:                 gotgbot.InputFileByID(fileId),
			Caption:               mediaText,
			ParseMode:             "HTML",
			ShowCaptionAboveMedia: userSetting.CaptionAbove,
			HasSpoiler:            userSetting.Spoiler,
		}
	case db.DOCUMENT:
		media = gotgbot.InputMediaDocument{
			Media:     gotgbot.InputFileByID(fileId),
			Caption:   mediaText,
			ParseMode: "HTML",
		}
	case db.AUDIO:
		media = gotgbot.InputMediaAudio{
			Media:     gotgbot.InputFileByID(fileId),
			Caption:   mediaText,
			ParseMode: "HTML",
		}
	case db.VIDEO:
		media = gotgbot.InputMediaVideo{
			Media:                 gotgbot.InputFileByID(fileId),
			Caption:               mediaText,
			ParseMode:             "HTML",
			ShowCaptionAboveMedia: userSetting.CaptionAbove,
			HasSpoiler:            userSetting.Spoiler,
		}
	case db.VOICE:
		media = gotgbot.InputMediaAudio{
			Media:     gotgbot.InputFileByID(fileId),
			Caption:   mediaText,
			ParseMode: "HTML",
		}
	default:
		media = nil
	}

	var successChats []string
	var failedChats []string

	oldDataType := post.MsgType

	newPostId := helpers.GenerateUniqueString()

	for _, chat := range post.Chats {
		chatId := chat.ChatId
		msgId := chat.MsgId

		switch oldDataType {
		case db.TEXT:
			if fileId != "" {
				if media == nil {
					_, _, _ = message.EditText(b, "Something went wrong. Please try again later. or read help menu", nil)
					return nil
				}
				_, _, err := b.EditMessageMedia(
					media,
					&gotgbot.EditMessageMediaOpts{
						ChatId:      chatId,
						MessageId:   msgId,
						ReplyMarkup: keyboard,
					})
				if err != nil {
					failedChats = append(failedChats, fmt.Sprintf("<code>%d</code>", chatId))
					continue
				} else {
					successChats = append(successChats, fmt.Sprintf("<code>%d</code>", chatId))
					_, _ = db.AddPost(newPostId, msg.From.Id, chatId, msgId, dataType, fileId, post.Buttons, post.FilterReply)
				}
			} else {
				_, _, err := b.EditMessageText(mediaText, &gotgbot.EditMessageTextOpts{
					ChatId:             chatId,
					MessageId:          msgId,
					ParseMode:          "HTML",
					ReplyMarkup:        keyboard,
					LinkPreviewOptions: &gotgbot.LinkPreviewOptions{IsDisabled: userSetting.WebPreview},
				})
				if err != nil {
					failedChats = append(failedChats, fmt.Sprintf("<code>%d</code>", chatId))
					continue
				} else {
					successChats = append(successChats, fmt.Sprintf("<code>%d</code>", chatId))
					_, _ = db.AddPost(newPostId, msg.From.Id, chatId, msgId, dataType, fileId, post.Buttons, post.FilterReply)
				}
			}
		case db.PHOTO:
			if fileId != "" {
				_, _, err := b.EditMessageMedia(
					media,
					&gotgbot.EditMessageMediaOpts{
						ChatId:      chatId,
						MessageId:   msgId,
						ReplyMarkup: keyboard,
					})
				if err != nil {
					failedChats = append(failedChats, fmt.Sprintf("<code>%d</code>", chatId))
					continue
				} else {
					successChats = append(successChats, fmt.Sprintf("<code>%d</code>", chatId))
					_, _ = db.AddPost(newPostId, msg.From.Id, chatId, msgId, dataType, fileId, post.Buttons, post.FilterReply)
				}
			} else {
				_, _, err := b.EditMessageCaption(&gotgbot.EditMessageCaptionOpts{
					ChatId:      chatId,
					MessageId:   msgId,
					Caption:     mediaText,
					ParseMode:   "HTML",
					ReplyMarkup: keyboard,
				})
				if err != nil {
					failedChats = append(failedChats, fmt.Sprintf("<code>%d</code>", chatId))
					continue
				} else {
					successChats = append(successChats, fmt.Sprintf("<code>%d</code>", chatId))
					_, _ = db.AddPost(newPostId, msg.From.Id, chatId, msgId, dataType, fileId, post.Buttons, post.FilterReply)
				}
			}
		case db.STICKER:
			_, _, err := b.EditMessageReplyMarkup(&gotgbot.EditMessageReplyMarkupOpts{
				ChatId:      chatId,
				MessageId:   msgId,
				ReplyMarkup: keyboard,
			})
			if err != nil {
				failedChats = append(failedChats, fmt.Sprintf("<code>%d</code>", chatId))
				continue
			} else {
				successChats = append(successChats, fmt.Sprintf("<code>%d</code>", chatId))
				_, _ = db.AddPost(newPostId, msg.From.Id, chatId, msgId, dataType, fileId, post.Buttons, post.FilterReply)
			}
		case db.AUDIO:
			_, _, err := b.EditMessageMedia(
				media,
				&gotgbot.EditMessageMediaOpts{
					ChatId:      chatId,
					MessageId:   msgId,
					ReplyMarkup: keyboard,
				})
			if err != nil {
				failedChats = append(failedChats, fmt.Sprintf("<code>%d</code>", chatId))
				continue
			} else {
				successChats = append(successChats, fmt.Sprintf("<code>%d</code>", chatId))
				_, _ = db.AddPost(newPostId, msg.From.Id, chatId, msgId, dataType, fileId, post.Buttons, post.FilterReply)
			}
		case db.VIDEO:
			_, _, err := b.EditMessageMedia(
				media,
				&gotgbot.EditMessageMediaOpts{
					ChatId:      chatId,
					MessageId:   msgId,
					ReplyMarkup: keyboard,
				})
			if err != nil {
				failedChats = append(failedChats, fmt.Sprintf("<code>%d</code>", chatId))
				continue
			} else {
				successChats = append(successChats, fmt.Sprintf("<code>%d</code>", chatId))
				_, _ = db.AddPost(newPostId, msg.From.Id, chatId, msgId, dataType, fileId, post.Buttons, post.FilterReply)
			}
		case db.VOICE:
			_, _, err := b.EditMessageMedia(
				media,
				&gotgbot.EditMessageMediaOpts{
					ChatId:      chatId,
					MessageId:   msgId,
					ReplyMarkup: keyboard,
				})
			if err != nil {
				failedChats = append(failedChats, fmt.Sprintf("<code>%d</code>", chatId))
				continue
			} else {
				successChats = append(successChats, fmt.Sprintf("<code>%d</code>", chatId))
				_, _ = db.AddPost(newPostId, msg.From.Id, chatId, msgId, dataType, fileId, post.Buttons, post.FilterReply)
			}
		case db.GIF:
			_, _, err := b.EditMessageMedia(
				media,
				&gotgbot.EditMessageMediaOpts{
					ChatId:      chatId,
					MessageId:   msgId,
					ReplyMarkup: keyboard,
				})
			if err != nil {
				failedChats = append(failedChats, fmt.Sprintf("<code>%d</code>", chatId))
				continue
			} else {
				successChats = append(successChats, fmt.Sprintf("<code>%d</code>", chatId))
				_, _ = db.AddPost(newPostId, msg.From.Id, chatId, msgId, dataType, fileId, post.Buttons, post.FilterReply)
			}
		case db.DOCUMENT:
			_, _, err := b.EditMessageMedia(
				media,
				&gotgbot.EditMessageMediaOpts{
					ChatId:      chatId,
					MessageId:   msgId,
					ReplyMarkup: keyboard,
				})
			if err != nil {
				failedChats = append(failedChats, fmt.Sprintf("<code>%d</code>", chatId))
				continue
			} else {
				successChats = append(successChats, fmt.Sprintf("<code>%d</code>", chatId))
				_, _ = db.AddPost(newPostId, msg.From.Id, chatId, msgId, dataType, fileId, post.Buttons, post.FilterReply)
			}
		default:
			_, _, _ = message.EditText(b, "Unknown data type.", &gotgbot.EditMessageTextOpts{ParseMode: "HTML"})
			return nil
		}
	}

	var text string
	// Send failed message
	if len(failedChats) > 0 {
		text += fmt.Sprintf("Failed to re-post to the following chats: %s", strings.Join(failedChats, ", "))
	}

	if len(successChats) > 0 {
		text += fmt.Sprintf("Re-posted to the following chats: %s", strings.Join(successChats, ", "))
	}
	go func() {
		err := db.RemovePost(post.PostId)
		if err != nil {
			log.Printf("repost: Error removing post: %v", err)
		}
	}()

	text += fmt.Sprintf("\n\nNewPost ID: <code>%s</code>", newPostId)
	_, _, err = message.EditText(b, text, &gotgbot.EditMessageTextOpts{ParseMode: "HTML", ReplyMarkup: helpers.PostButton(newPostId)})
	return err
}
