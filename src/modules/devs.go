package modules

import (
	"AshokShau/channelManager/src/config"
	"AshokShau/channelManager/src/db"
	"AshokShau/channelManager/src/modules/utils/helpers"
	"bufio"
	"bytes"
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"strconv"
	"time"
)

func banUser(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	if msg.Chat.Type != "private" {
		return nil
	}

	if msg.From.Id != config.OwnerId {
		_, _ = msg.Reply(b, "You must be the owner to use this command.", helpers.Shtml())
		return nil
	}

	args := ctx.Args()[1:]
	if len(args) < 1 {
		_, _ = msg.Reply(b, "Please provide a user ID to ban.", helpers.Shtml())
		return nil
	}

	userId, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		_, _ = msg.Reply(b, "Invalid user ID.", helpers.Shtml())
		return nil
	}

	// Ban the user
	err = db.AddBan(userId)
	if err != nil {
		_, _ = msg.Reply(b, "Error banning user.\n\n"+err.Error(), helpers.Shtml())
		return err
	}
	_, _ = b.SendMessage(userId, "You have been banned from using my bot.", nil)
	_, _ = msg.Reply(b, "User banned successfully.", helpers.Shtml())
	return nil
}

func unbanUser(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	if msg.Chat.Type != "private" {
		return nil
	}

	if msg.From.Id != config.OwnerId {
		_, _ = msg.Reply(b, "You must be the owner to use this command.", helpers.Shtml())
		return nil
	}

	args := ctx.Args()[1:]
	if len(args) < 1 {
		_, _ = msg.Reply(b, "Please provide a user ID to unban.", helpers.Shtml())
		return nil
	}

	userId, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		_, _ = msg.Reply(b, "Invalid user ID.", helpers.Shtml())
		return nil
	}

	// Unban the user
	err = db.RemoveBan(userId)
	if err != nil {
		_, _ = msg.Reply(b, "Error unbanning user.\n\n"+err.Error(), helpers.Shtml())
		return err
	}
	_, _ = b.SendMessage(userId, "You have been unbanned from using my bot.", nil)
	_, _ = msg.Reply(b, "User unbanned successfully.", helpers.Shtml())
	return nil
}

func getBans(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	if msg.Chat.Type != "private" {
		return nil
	}

	if msg.From.Id != config.OwnerId {
		_, _ = msg.Reply(b, "You must be the owner to use this command.", helpers.Shtml())
		return nil
	}

	bans, err := db.GetBans()
	if err != nil {
		_, _ = msg.Reply(b, "Error retrieving banned users.\n\n"+err.Error(), helpers.Shtml())
		return err
	}

	if len(bans) == 0 {
		_, _ = msg.Reply(b, "No banned users found.", helpers.Shtml())
		return nil
	}

	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	_, _ = w.WriteString("Banned Users:\n\n")
	for i, ban := range bans {
		_, _ = w.WriteString(fmt.Sprintf("%d. User ID: %d\n", i+1, ban.UserId))
	}

	_ = w.Flush()

	if _, err = b.SendDocument(msg.Chat.Id, gotgbot.InputFileByReader("bans.txt", &buf), &gotgbot.SendDocumentOpts{Caption: "Banned Users", ParseMode: "HTML"}); err != nil {
		return fmt.Errorf("failed to send backup file: %w", err)
	}

	return nil
}

func broadcast(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	if msg.Chat.Type != "private" {
		return nil
	}

	if msg.From.Id != config.OwnerId {
		_, _ = msg.Reply(b, "You must be the owner to use this command.", helpers.Shtml())
		return nil
	}

	reply := ctx.EffectiveMessage.ReplyToMessage
	if reply == nil {
		_, err := ctx.EffectiveMessage.Reply(b, "❌ <b>Reply to a message to broadcast</b>", &gotgbot.SendMessageOpts{ParseMode: "HTML"})
		if err != nil {
			return fmt.Errorf("error while replying to user: %v", err)
		}
		return ext.EndGroups
	}

	button := &gotgbot.InlineKeyboardMarkup{}
	if reply.ReplyMarkup != nil {
		button.InlineKeyboard = reply.ReplyMarkup.InlineKeyboard
	}

	servedUsers, err := db.GetAllUsers()
	if err != nil {
		_, _ = msg.Reply(b, "Error getting users.\n\n"+err.Error(), helpers.Shtml())
		return err
	}

	successfulBroadcasts := 0
	for _, user := range servedUsers {
		_, err = b.CopyMessage(user, ctx.EffectiveMessage.Chat.Id, reply.MessageId, &gotgbot.CopyMessageOpts{ReplyMarkup: button})
		if err == nil {
			successfulBroadcasts++
		}
		time.Sleep(69 * time.Millisecond)
	}

	_, err = ctx.EffectiveMessage.Reply(b, fmt.Sprintf("✅ <b>Broadcast successfully to %d users</b>", successfulBroadcasts), &gotgbot.SendMessageOpts{ParseMode: "HTML"})
	if err != nil {
		return fmt.Errorf("[broadcast] failed to send reply message" + err.Error())
	}

	return ext.EndGroups
}

func stats(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	if msg.Chat.Type != "private" {
		return nil
	}

	if msg.From.Id != config.OwnerId {
		_, _ = msg.Reply(b, "You must be the owner to use this command.", helpers.Shtml())
		return nil
	}
	servedUsers, err := db.GetAllUsers()
	if err != nil {
		_, _ = msg.Reply(b, "Error getting users.\n\n"+err.Error(), helpers.Shtml())
		return err
	}

	_, _ = msg.Reply(b, fmt.Sprintf("Total served users: %d", len(servedUsers)), helpers.Shtml())
	return nil
}
