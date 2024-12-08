package modules

import (
	"AshokShau/channelManager/src/db"
	helpers2 "AshokShau/channelManager/src/modules/utils/helpers"
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"log"
	"strings"
	"time"
)

func sendPostCallback(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	query := ctx.Update.CallbackQuery
	user := query.From

	chatIds := isConnected(b, ctx, query.From.Id)
	if chatIds == nil {
		_, _ = query.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text:      "⚠️ You are not connected to any channels.\nPlease connect to a channel and try again. Bye 👋",
			ShowAlert: true,
		})
		time.Sleep(10 * time.Millisecond)
		_, _ = query.Message.Delete(b, nil)
		return nil
	}

	postId := strings.Split(query.Data, ".")[1]
	post, err := db.GetPost(postId)
	if err != nil || post == nil {
		_, _ = query.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text:      "❌ Post not found.\nPlease verify the post and try again. Bye 👋",
			ShowAlert: true,
		})
		_, _ = query.Message.Delete(b, nil)
		return err
	}

	_, _ = query.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
		Text:      "📤 Sending post to connected chats...\nThis may take some time.",
		ShowAlert: true,
	})

	keyboard := gotgbot.InlineKeyboardMarkup{InlineKeyboard: helpers2.BuildKeyboard(post.Buttons)}
	if keyboard.InlineKeyboard == nil {
		keyboard.InlineKeyboard = make([][]gotgbot.InlineKeyboardButton, 0)
	}
	newPostId := helpers2.GenerateUniqueString()
	var successChats, failedChats []string
	successCount, failedCount := 0, 0

	userSetting := db.GetUserSettings(user.Id)
	for i, chatId := range chatIds {
		// Pause for rate-limiting if needed
		if (successCount+failedCount)%23 == 0 && i > 0 {
			log.Printf("[sendPostCallback] Reached batch limit, pausing for 1 minute...")
			time.Sleep(1 * time.Minute)
			successCount = 0
			failedCount = 0
		}

		message, err := helpers2.PostEnumFuncMap[post.MsgType](b, ctx, chatId, post.FilterReply, post.FileID, &keyboard, userSetting)
		if err != nil {
			log.Printf("Failed to send post to chat %d: %v", chatId, err)
			failedChats = append(failedChats, fmt.Sprintf("<code>%d</code>", chatId))
			failedCount++
		} else {
			messageId := message.MessageId
			successChats = append(successChats, fmt.Sprintf("<a href='%s'>%d</a> \n(<code>!del %d %d</code>)", message.GetLink(), chatId, chatId, messageId))
			_, _ = db.AddPost(newPostId, user.Id, chatId, messageId, post.MsgType, post.FileID, post.Buttons, post.FilterReply)
			successCount++
		}
	}

	go func() {
		_ = db.RemovePost(postId)
	}()

	// Build response text
	var responseText strings.Builder
	responseText.WriteString("<b>📋 Post Result Summary:</b>\n\n")

	if len(successChats) > 0 {
		responseText.WriteString(fmt.Sprintf("✅ <b>Successfully sent to %d chats:</b>\n", len(successChats)))
		responseText.WriteString(strings.Join(successChats, "\n") + "\n\n")
	}

	if len(failedChats) > 0 {
		responseText.WriteString(fmt.Sprintf("❌ <b>Failed to send to %d chats:</b>\n", len(failedChats)))
		responseText.WriteString(strings.Join(failedChats, "\n") + "\n")
	}

	responseText.WriteString(fmt.Sprintf("\n<b>🆔 PostId:</b> <code>%s</code>", newPostId))

	_, err = msg.Reply(b, responseText.String(), &gotgbot.SendMessageOpts{ParseMode: "HTML", ReplyMarkup: helpers2.PostButton(newPostId), ReplyParameters: &gotgbot.ReplyParameters{AllowSendingWithoutReply: true}})
	time.Sleep(10 * time.Millisecond)
	_, _ = msg.Delete(b, nil)
	return err
}

func deletePostCallback(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	query := ctx.Update.CallbackQuery
	chatIds := isConnected(b, ctx, query.From.Id)
	if chatIds == nil {
		return nil
	}

	postId := strings.Split(query.Data, ".")[1]
	post, err := db.GetPost(postId)
	if err != nil {
		return err
	}

	if post == nil {
		_, _ = query.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Post not found. \nBye 👋", ShowAlert: true})
		_, _ = query.Message.Delete(b, nil)
		return nil
	}

	_, _ = query.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
		Text:      "📤 Deleting post from connected chats...\nThis may take some time.",
		ShowAlert: true,
	})
	postChats := post.Chats
	for _, chat := range postChats {
		_, err = b.DeleteMessage(chat.ChatId, chat.MsgId, nil)
		if err != nil {
			log.Printf("deletePost: Error deleting message from ChatID %d: %v", chat.ChatId, err)
			continue
		}
		time.Sleep(200 * time.Millisecond)
	}

	_ = db.RemovePost(postId)
	_, _ = msg.Delete(b, nil)
	return nil
}

func repostCallback(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	query := ctx.Update.CallbackQuery
	chatIds := isConnected(b, ctx, query.From.Id)
	if chatIds == nil {
		return nil
	}

	postId := strings.Split(query.Data, ".")[1]
	post, err := db.GetPost(postId)
	if err != nil || post == nil {
		_, _ = query.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Post not found.\nPlease try again. bye 👋", ShowAlert: true})
		_, _ = query.Message.Delete(b, nil)
		return err
	}

	keyboard := gotgbot.InlineKeyboardMarkup{InlineKeyboard: helpers2.BuildKeyboard(post.Buttons)}
	if keyboard.InlineKeyboard == nil {
		keyboard.InlineKeyboard = make([][]gotgbot.InlineKeyboardButton, 0)
	}

	_, _ = query.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "📤 Reposting post to connected chats...\nThis may take some time.", ShowAlert: true})

	// delete old post
	for _, chat := range post.Chats {
		_, err = b.DeleteMessage(chat.ChatId, chat.MsgId, nil)
		if err != nil {
			log.Printf("repost: Error deleting message from ChatID %d: %v", chat.ChatId, err)
			continue
		}
		time.Sleep(100 * time.Millisecond)
	}

	// send new post
	newPostId := helpers2.GenerateUniqueString()
	var successChats, failedChats []string
	successCount := 0
	failedCount := 0
	userSetting := db.GetUserSettings(query.From.Id)
	for i, chatId := range chatIds {
		// Check if we need to sleep before processing the next batch
		if (successCount+failedCount)%23 == 0 && i > 0 {
			log.Printf("[repostCallback] Sleeping for 1 minute before processing the next batch...")
			time.Sleep(1 * time.Minute)
			successCount = 0
			failedCount = 0
		}

		// Attempt to send the message
		message, err := helpers2.PostEnumFuncMap[post.MsgType](b, ctx, chatId, post.FilterReply, post.FileID, &keyboard, userSetting)
		if err != nil {
			failedChats = append(failedChats, fmt.Sprintf("<code>%d</code>", chatId))
			failedCount++
		} else {
			// Save the post in the database
			_, _ = db.AddPost(newPostId, msg.From.Id, chatId, message.MessageId, post.MsgType, post.FileID, post.Buttons, post.FilterReply)
			successChats = append(successChats, fmt.Sprintf("<a href='%s'>%d</a> \n(<code>!del %d %d</code>)", message.GetLink(), chatId, chatId, message.MessageId))
			successCount++
		}
	}

	// prepare the response text
	var responseText strings.Builder
	responseText.WriteString("<b>Post Result Summary:</b>\n\n")

	if len(successChats) > 0 {
		responseText.WriteString("✅ <b>Successfully sent to:</b>\n")
		responseText.WriteString(strings.Join(successChats, "\n") + "\n\n")
	}

	if len(failedChats) > 0 {
		responseText.WriteString("❌ <b>Failed to send to:</b>\n")
		responseText.WriteString(strings.Join(failedChats, "\n") + "\n")
	}

	responseText.WriteString(fmt.Sprintf("<b>PostId:</b> <code>%s</code>", newPostId))

	_, _ = msg.Reply(b, responseText.String(), &gotgbot.SendMessageOpts{
		ParseMode:          "HTML",
		ReplyMarkup:        helpers2.PostButton(newPostId),
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{IsDisabled: true},
		ReplyParameters:    &gotgbot.ReplyParameters{AllowSendingWithoutReply: true},
	})

	_, _ = msg.Delete(b, nil)
	return err
}
