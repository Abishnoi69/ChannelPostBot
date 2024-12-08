package modules

import (
	"AshokShau/channelManager/src/db"
	"AshokShau/channelManager/src/modules/utils/helpers"
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func delAllPosts(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	chatIds := isConnected(b, ctx, msg.From.Id)
	if chatIds == nil {
		return nil
	}
	args := ctx.Args()[1:]
	if len(args) < 1 {
		_, err := msg.Reply(b, "Please provide a PostId to delete all posts .", helpers.Shtml())
		return err
	}

	postId := args[0]
	post, err := db.GetPost(postId)
	if err != nil || post == nil {
		_, _ = msg.Reply(b, "Post not found or error retrieving post.", helpers.Shtml())
		return err
	}

	for _, chat := range post.Chats {
		_, err = b.DeleteMessage(chat.ChatId, chat.MsgId, nil)
		if err != nil {
			log.Printf("deleteAllPost: Error deleting message from ChatID %d: %v", chat.ChatId, err)
			continue
		}
		time.Sleep(100 * time.Millisecond)
	}
	_ = db.RemovePost(postId)
	_, _ = msg.Reply(b, "All posts deleted.", helpers.Shtml())
	return nil

}

func deletePost(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	chatIds := isConnected(b, ctx, msg.From.Id)
	if chatIds == nil {
		return nil
	}

	args := ctx.Args()[1:]
	if len(args) < 2 {
		_, err := msg.Reply(b, "Please provide both a channel ID and a message ID to delete.\nUsage: <code>!delete channel_id msg_id</code>", helpers.Shtml())
		return err
	}

	_, err := b.DeleteMessage(helpers.ToInt64(args[0]), helpers.ToInt64(args[1]), nil)
	if err != nil {
		_, _ = msg.Reply(b, "Error deleting message.", helpers.Shtml())
		return err
	}

	_, _ = msg.Reply(b, "Message deleted.", helpers.Shtml())
	return nil
}

func createPost(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	if msg.Chat.Type != "private" {
		return nil
	}

	args := ctx.Args()[1:]
	if len(args) == 0 && msg.ReplyToMessage == nil {
		_, _ = msg.Reply(b, "Please provide a message to send.", helpers.Shtml())
		return nil
	}

	text, dataType, fileId, buttons, errorMsg := helpers.GetMsgType(msg)
	if dataType == -1 {
		_, _ = msg.Reply(b, errorMsg, helpers.Shtml())
		return ext.EndGroups
	}

	postId := helpers.GenerateUniqueString()
	buttonNoText := helpers.RevertButtons(buttons)
	if buttonNoText == "" {
		buttonNoText = "No buttons"
	}

	keyboard := gotgbot.InlineKeyboardMarkup{InlineKeyboard: helpers.BuildKeyboard(buttons)}
	if keyboard.InlineKeyboard == nil {
		keyboard.InlineKeyboard = make([][]gotgbot.InlineKeyboardButton, 0)
	}

	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []gotgbot.InlineKeyboardButton{
		{Text: "Send Post to all chats", CallbackData: fmt.Sprintf("send.%s", postId)},
		{Text: "Copy Button Text", CopyText: &gotgbot.CopyTextButton{Text: buttonNoText}},
	})

	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []gotgbot.InlineKeyboardButton{
		{Text: "Share Vai Inline", CopyText: &gotgbot.CopyTextButton{Text: fmt.Sprintf("@%s %s", b.Username, postId)}},
	})

	userSettings := db.GetUserSettings(msg.From.Id)
	send, err := helpers.PostEnumFuncMap[dataType](b, ctx, ctx.EffectiveChat.Id, text, fileId, &keyboard, userSettings)
	if err != nil {
		_, _ = msg.Reply(b, "Error creating Post.\n\n<code>"+err.Error()+"</code>", helpers.Shtml())
		return fmt.Errorf("createPost: error in sending message: %v", err)
	}

	_, err = db.AddPost(postId, msg.From.Id, b.Id, send.MessageId, dataType, fileId, buttons, text)
	if err != nil {
		_, _ = msg.Reply(b, "Error creating Post.", helpers.Shtml())
		return err
	}
	return ext.EndGroups
}

func getPost(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	args := ctx.Args()[1:]
	if len(args) < 1 {
		_, _ = msg.Reply(b, "Please provide a PostId to share.\nUsage: <code>!share PostId</code>", helpers.Shtml())
		return nil
	}

	postId := args[0]
	post, err := db.GetPost(postId)
	if err != nil || post == nil {
		_, _ = msg.Reply(b, "Post not found or error retrieving post.", helpers.Shtml())
		return err
	}

	keyboard := gotgbot.InlineKeyboardMarkup{InlineKeyboard: helpers.BuildKeyboard(post.Buttons)}
	if keyboard.InlineKeyboard == nil {
		keyboard.InlineKeyboard = make([][]gotgbot.InlineKeyboardButton, 0)
	}
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []gotgbot.InlineKeyboardButton{
		{Text: "Send Post to all chats", CallbackData: fmt.Sprintf("send.%s", postId)},
	})

	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []gotgbot.InlineKeyboardButton{
		{Text: "Share Vai Inline", CopyText: &gotgbot.CopyTextButton{Text: fmt.Sprintf("@%s %s", b.Username, postId)}},
	})

	userSettings := db.GetUserSettings(msg.From.Id)
	_, err = helpers.PostEnumFuncMap[post.MsgType](b, ctx, ctx.EffectiveChat.Id, post.FilterReply, post.FileID, &keyboard, userSettings)
	if err != nil {
		_, _ = msg.Reply(b, "Error sending post.", helpers.Shtml())
		return fmt.Errorf("getPost: error in sending message: %v", err)
	}

	return nil
}

func sendPost(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	var successChats, failedChats []string

	chatIds := isConnected(b, ctx, msg.From.Id)
	if chatIds == nil {
		_, _ = msg.Reply(b, "You are not connected to any chats. Please connect to at least one chat to send posts.", helpers.Shtml())
		return nil
	}

	reply := msg.ReplyToMessage
	if reply == nil {
		_, _ = msg.Reply(b, "To send a post, you need to reply to a message that you want to share with all connected chats.", helpers.Shtml())
		return nil
	}

	message, err := msg.Reply(b, "📤 Sending post to connected chats...\nThis may take some time.", helpers.Shtml())
	if err != nil {
		return err
	}

	postId := helpers.GenerateUniqueString()
	userSettings := db.GetUserSettings(msg.From.Id)
	forwardTag := userSettings.ForwardTag

	if forwardTag {
		postText, dataType, fileId, buttons, _ := helpers.GetMsgType(msg)
		successCount, failedCount := 0, 0

		for i, chatId := range chatIds {
			if (successCount+failedCount)%23 == 0 && i > 0 {
				// if send count is 23, sleep for 1 minute
				log.Printf("[sendPost] Sleeping for 1 minute...")
				time.Sleep(1 * time.Minute)
				successCount, failedCount = 0, 0
			}

			message, err := b.ForwardMessage(chatId, msg.Chat.Id, reply.MessageId, &gotgbot.ForwardMessageOpts{
				DisableNotification: userSettings.NoNotif,
				ProtectContent:      userSettings.Protect,
			})

			if err != nil {
				log.Printf("Failed to forward message to chatId %d: %v", chatId, err)
				failedCount++
				continue
			}

			_, _ = db.AddPost(postId, msg.From.Id, chatId, message.MessageId, dataType, fileId, buttons, postText)
			successCount++
			time.Sleep(50 * time.Millisecond)
		}

		_, _, err = message.EditText(b, fmt.Sprintf("Message forwarded to all connected chats.\n\n<b>PostId:</b> <code>%s</code>", postId), &gotgbot.EditMessageTextOpts{
			ParseMode:   "HTML",
			ReplyMarkup: helpers.PostButton(postId),
		})

		return err
	}

	postText, dataType, fileId, buttons, errorMsg := helpers.GetMsgType(msg)
	if dataType == -1 {
		_, _, err = message.EditText(b, errorMsg, &gotgbot.EditMessageTextOpts{
			ParseMode: "HTML",
		})
		return err
	}

	keyboard := gotgbot.InlineKeyboardMarkup{InlineKeyboard: helpers.BuildKeyboard(buttons)}
	if len(keyboard.InlineKeyboard) == 0 {
		keyboard.InlineKeyboard = [][]gotgbot.InlineKeyboardButton{}
	}

	successCount, failedCount := 0, 0
	for i, chatId := range chatIds {
		if (successCount+failedCount)%23 == 0 && i > 0 {
			time.Sleep(1 * time.Minute)
			successCount, failedCount = 0, 0
		}

		message, err := helpers.PostEnumFuncMap[dataType](b, ctx, chatId, postText, fileId, &keyboard, userSettings)
		if err != nil {
			log.Printf("Failed to send post to chatId %d: %v", chatId, err)
			failedChats = append(failedChats, fmt.Sprintf("<code>%d</code>", chatId))
			failedCount++
		} else {
			// Include the !del command for easy deletion
			successChats = append(successChats, fmt.Sprintf(
				"<a href='%s'>%d</a>\n(<code>!del %d %d</code>)",
				message.GetLink(), chatId, chatId, message.MessageId,
			))
			_, _ = db.AddPost(postId, msg.From.Id, chatId, message.MessageId, dataType, fileId, buttons, postText)
			successCount++
		}

		time.Sleep(50 * time.Millisecond)
	}

	responseText := fmt.Sprintf(
		"<b>Post Result Summary:</b>\n\n✅ <b>Successfully sent to:</b>\n%s\n\n❌ <b>Failed to send to:</b>\n%s\n<b>PostId:</b> <code>%s</code>",
		strings.Join(successChats, "\n"),
		strings.Join(failedChats, "\n"),
		postId,
	)

	_, _, err = message.EditText(b, responseText, &gotgbot.EditMessageTextOpts{
		ParseMode:   "HTML",
		ReplyMarkup: helpers.PostButton(postId),
	})
	return err
}

// sendEmptyQueryResponse sends a response for an empty inline query.
func sendEmptyQueryResponse(bot *gotgbot.Bot, ctx *ext.Context) error {
	_, err := ctx.InlineQuery.Answer(bot, nil, &gotgbot.AnswerInlineQueryOpts{
		IsPersonal: true,
		CacheTime:  10,
		Button: &gotgbot.InlineQueryResultsButton{
			Text:           "Give me postId !",
			StartParameter: "start",
		},
	})
	return err
}

// noResultsArticle creates an inline query result indicating no results were found.
func noResultsArticle(query string) gotgbot.InlineQueryResult {
	return gotgbot.InlineQueryResultArticle{
		Id:    strconv.Itoa(rand.Intn(100000)),
		Title: "Post not found or error retrieving post.",
		InputMessageContent: gotgbot.InputTextMessageContent{
			MessageText: fmt.Sprintf("<i>👋 Sorry, I couldn't find any results for '%s'!</i>", query),
			ParseMode:   gotgbot.ParseModeHTML,
		},
		Description: "No results found for your query.",
		ReplyMarkup: &gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
				{{Text: "Search Again", SwitchInlineQueryCurrentChat: &query}},
			},
		},
	}
}

// inlineSharePost handles inline queries to share a post by its PostId
func inlineSharePost(b *gotgbot.Bot, ctx *ext.Context) error {
	query := strings.TrimSpace(ctx.InlineQuery.Query)
	parts := strings.Fields(query)

	if len(parts) == 0 {
		return sendEmptyQueryResponse(b, ctx)
	}

	postId := parts[0]

	// Retrieve the post from the database
	post, err := db.GetPost(postId)
	if err != nil || post == nil {
		_, _ = ctx.InlineQuery.Answer(b, []gotgbot.InlineQueryResult{noResultsArticle(query)}, &gotgbot.AnswerInlineQueryOpts{
			IsPersonal: true,
			CacheTime:  500,
		})
		return err
	}

	dataType, fileId, buttons, postText := post.MsgType, post.FileID, post.Buttons, post.FilterReply
	keyboard := gotgbot.InlineKeyboardMarkup{InlineKeyboard: helpers.BuildKeyboard(buttons)}
	if keyboard.InlineKeyboard == nil {
		keyboard.InlineKeyboard = make([][]gotgbot.InlineKeyboardButton, 0)
	}

	// Create a random inline query result ID
	resultId := strconv.Itoa(rand.Intn(100000))
	userSettings := db.GetUserSettings(ctx.InlineQuery.From.Id)
	var results []gotgbot.InlineQueryResult

	// Handle the different MsgType cases
	switch dataType {
	case db.TEXT:
		results = append(results, gotgbot.InlineQueryResultArticle{
			Id:    resultId,
			Title: "Text Post",
			InputMessageContent: gotgbot.InputTextMessageContent{
				MessageText: postText,
				ParseMode:   gotgbot.ParseModeHTML,
				LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
					IsDisabled: userSettings.WebPreview,
				},
			},
			ReplyMarkup: &keyboard,
		})

	case db.PHOTO:
		results = append(results, gotgbot.InlineQueryResultCachedPhoto{
			Id:                    resultId,
			PhotoFileId:           fileId,
			Caption:               postText,
			ParseMode:             gotgbot.ParseModeHTML,
			ReplyMarkup:           &keyboard,
			ShowCaptionAboveMedia: userSettings.CaptionAbove,
		})

	case db.VIDEO:
		results = append(results, gotgbot.InlineQueryResultCachedVideo{
			Id:                    resultId,
			VideoFileId:           fileId,
			Caption:               postText,
			ParseMode:             gotgbot.ParseModeHTML,
			ReplyMarkup:           &keyboard,
			ShowCaptionAboveMedia: userSettings.CaptionAbove,
		})

	case db.DOCUMENT:
		results = append(results, gotgbot.InlineQueryResultCachedDocument{
			Id:             resultId,
			Title:          "Document",
			DocumentFileId: fileId,
			Caption:        postText,
			ParseMode:      gotgbot.ParseModeHTML,
			ReplyMarkup:    &keyboard,
		})

	case db.STICKER:
		results = append(results, gotgbot.InlineQueryResultCachedSticker{
			Id:            resultId,
			StickerFileId: fileId,
			ReplyMarkup:   &keyboard,
		})

	case db.GIF:
		results = append(results, gotgbot.InlineQueryResultCachedGif{
			Id:                    resultId,
			GifFileId:             fileId,
			Caption:               postText,
			ParseMode:             gotgbot.ParseModeHTML,
			ReplyMarkup:           &keyboard,
			ShowCaptionAboveMedia: userSettings.CaptionAbove,
		})

	case db.AUDIO:
		results = append(results, gotgbot.InlineQueryResultCachedAudio{
			Id:          resultId,
			AudioFileId: fileId,
			Caption:     postText,
			ParseMode:   gotgbot.ParseModeHTML,
			ReplyMarkup: &keyboard,
		})

	case db.VOICE:
		results = append(results, gotgbot.InlineQueryResultCachedVoice{
			Id:          resultId,
			VoiceFileId: fileId,
			Caption:     postText,
			ParseMode:   gotgbot.ParseModeHTML,
			ReplyMarkup: &keyboard,
		})
	default:
		results = append(results, noResultsArticle(postId))
	}

	// Send the results
	_, _ = ctx.InlineQuery.Answer(b, results, &gotgbot.AnswerInlineQueryOpts{
		IsPersonal: true,
		CacheTime:  500,
	})

	return nil
}
