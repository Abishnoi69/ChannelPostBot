package modules

import (
	"AshokShau/channelManager/src/config"
	"AshokShau/channelManager/src/db"
	"AshokShau/channelManager/src/modules/utils/helpers"
	"AshokShau/channelManager/src/modules/utils/onlyAdmins"
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
	"log"
	"strconv"
	"strings"
	"time"
)

func isConnected(b *gotgbot.Bot, ctx *ext.Context, userId int64) []int64 {
	msg := ctx.EffectiveMessage

	// If the chat is not private, return the chat ID
	if msg.Chat.Type != "private" {
		return []int64{msg.Chat.Id}
	}

	// Fetch user connection data
	conn := db.Connection(userId)
	connectedChats := conn.ChatIds

	// If no connected chats, prompt the user to connect
	if connectedChats == nil || len(connectedChats) == 0 {
		text := "⚠️ You are not connected to any chats.\n\nUse <code>/add chat_id</code> to connect."
		_, _ = msg.Reply(b, text, helpers.Shtml())
		return nil
	}

	var errorMessage strings.Builder

	// Validate connected chats
	for _, chatId := range connectedChats {
		getChat := onlyAdmins.GetChatCache(chatId)

		// Load chat cache if not cached
		if !getChat.Cached {
			getChat = onlyAdmins.LoadChatCache(b, chatId)
		}

		// If chat still not cached, disconnect and log error
		if !getChat.Cached {
			db.DisconnectId(userId, chatId)
			errorMessage.WriteString(fmt.Sprintf("<code>%d</code> (Chat not found)\n", chatId))
			continue
		}

		// Check if user is an admin
		userCached, _ := onlyAdmins.IsUserAdmin(chatId, userId)
		if !userCached {
			time.Sleep(20 * time.Millisecond)
			if reloaded := onlyAdmins.LoadAdminCache(b, chatId); !reloaded.Cached {
				db.DisconnectId(userId, chatId)
				errorMessage.WriteString(fmt.Sprintf("<code>%d</code> (Failed to verify admin status)\n", chatId))
				continue
			}
		}

		// Verify admin status of the user
		userCached, isAdmin := onlyAdmins.IsUserAdmin(chatId, userId)
		if userCached && !isAdmin {
			db.DisconnectId(userId, chatId)
			errorMessage.WriteString(fmt.Sprintf("<code>%d</code> (You are not an admin)\n", chatId))
			continue
		}

		if !isAdmin {
			db.DisconnectId(userId, chatId)
			errorMessage.WriteString(fmt.Sprintf("<code>%d</code> (You are not an admin)\n", chatId))
			continue
		}

		// Verify bot's admin status
		_, isBotAdmin := onlyAdmins.IsUserAdmin(chatId, b.Id)
		if !isBotAdmin {
			if reloaded := onlyAdmins.LoadAdminCache(b, chatId); !reloaded.Cached {
				db.DisconnectId(userId, chatId)
				errorMessage.WriteString(fmt.Sprintf("<code>%d</code> (Failed to verify admin status)\n", chatId))
				continue
			}
		}

		time.Sleep(50 * time.Millisecond)
	}

	// Reply with accumulated error messages if any
	if errorMessage.Len() > 0 {
		_, _ = msg.Reply(b, errorMessage.String(), helpers.Shtml())
		return nil
	}

	// If no valid connections remain, notify the user
	if len(conn.ChatIds) == 0 {
		text := "⚠️ No valid connections found. Use <code>/add chat_id</code> to connect."
		_, _ = msg.Reply(b, text, helpers.Shtml())
		return nil
	}

	return conn.ChatIds
}

func connect(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	if msg.Chat.Type != "private" {
		return nil
	}

	args := ctx.Args()[1:]
	reply := msg.ReplyToMessage
	if len(args) == 0 && (reply == nil || reply.ForwardOrigin == nil) {
		_, _ = msg.Reply(b, "Please forward a message from a chat so I can get the chat ID.\n\nMake sure that you & i are admin in the chat.", nil)
		return handlers.NextConversationState(CHATID)
	}

	var successfullyConnected []string
	var failedConnections []string

	if reply != nil && reply.ForwardOrigin != nil {
		origin := reply.ForwardOrigin.MergeMessageOrigin()
		if origin.Type == "channel" {
			chatId := origin.Chat.Id
			args = append(args, strconv.FormatInt(chatId, 10))
		} else if origin.Type == "chat" {
			chatId := origin.Chat.Id
			args = append(args, strconv.FormatInt(chatId, 10))
		} else {
			_, err := msg.Reply(b, "Please provide at least one chat ID to connect. Use /connect <chat_id> to connect to a chat.", nil)
			return err
		}
	}

	for _, arg := range args {
		chatId, err := strconv.ParseInt(arg, 10, 64)
		if err != nil {
			failedConnections = append(failedConnections, fmt.Sprintf("<code>%s</code> (Invalid ID)", arg))
			continue
		}

		getChat := onlyAdmins.GetChatCache(chatId)
		if !getChat.Cached {
			time.Sleep(400 * time.Millisecond)
			getChat = onlyAdmins.LoadChatCache(b, chatId)
		}

		if !getChat.Cached {
			failedConnections = append(failedConnections, fmt.Sprintf("<code>%d</code> (Chat not found)", chatId))
			continue
		}

		cached, _ := onlyAdmins.IsUserAdmin(chatId, msg.From.Id)
		if !cached {
			time.Sleep(100 * time.Millisecond)
			if admins := onlyAdmins.LoadAdminCache(b, chatId); !admins.Cached {
				failedConnections = append(failedConnections, fmt.Sprintf("<code>%d</code> (Failed to verify admin status)", chatId))
				continue
			}
		}

		if _, isUserAdmin := onlyAdmins.IsUserAdmin(chatId, msg.From.Id); !isUserAdmin {
			failedConnections = append(failedConnections, fmt.Sprintf("<code>%d</code> (You are not an admin)", chatId))
			continue
		}

		db.ConnectId(msg.From.Id, chatId)
		successfullyConnected = append(successfullyConnected, fmt.Sprintf("<b>%s</b> (<code>%d</code>)", getChat.ChatInfo.Title, chatId))
	}

	// Build reply text
	var text string
	if len(successfullyConnected) > 0 {
		text += "✅ Successfully connected to:\n" + strings.Join(successfullyConnected, "\n") + "\n\n"
	}
	if len(failedConnections) > 0 {
		text += "❌ Failed to connect to:\n" + strings.Join(failedConnections, "\n")
	}

	if text == "" {
		text = "No connections were made. Please check the chat IDs and try again."
	}

	_, _ = msg.Reply(b, text, helpers.Shtml())
	return handlers.EndConversation()
}
func disconnect(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	if msg.Chat.Type != "private" {
		return nil
	}

	args := ctx.Args()[1:]
	if len(args) == 0 {
		_, err := msg.Reply(b, "⚠️ Please provide at least one chat ID to disconnect.", nil)
		return err
	}

	connectedChats := isConnected(b, ctx, msg.From.Id)
	if connectedChats == nil || len(connectedChats) == 0 {
		log.Printf("[disconnect] No connected chats found for user %d", msg.From.Id)
		return nil
	}

	log.Printf("[disconnect] Connected chats before update: %v", connectedChats)

	var validChats []string
	var invalidChats []string

	for _, arg := range args {
		chatId, err := strconv.ParseInt(arg, 10, 64)
		if err != nil {
			log.Printf("[disconnect] Invalid chat ID: %v (%v)", arg, err)
			invalidChats = append(invalidChats, arg)
			continue
		}

		if !helpers.Contains(connectedChats, chatId) {
			log.Printf("[disconnect] Chat ID %d is not connected", chatId)
			invalidChats = append(invalidChats, arg)
			continue
		}

		db.DisconnectId(msg.From.Id, chatId)
		validChats = append(validChats, fmt.Sprintf("<code>%d</code>", chatId))
	}

	connectedChatsAfter := isConnected(b, ctx, msg.From.Id)
	log.Printf("[disconnect] Connected chats after update: %v", connectedChatsAfter)

	var response strings.Builder
	if len(validChats) > 0 {
		response.WriteString("✅ Successfully disconnected from the following chat(s):\n")
		response.WriteString(strings.Join(validChats, ", ") + "\n")
	}
	if len(invalidChats) > 0 {
		response.WriteString("⚠️ The following chat ID(s) are invalid or not connected:\n")
		response.WriteString(strings.Join(invalidChats, ", ") + "\n")
	}
	if response.Len() == 0 {
		response.WriteString("⚠️ No valid chat IDs provided. Please check and try again.")
	}

	_, err := msg.Reply(b, response.String(), helpers.Shtml())
	if err != nil {
		log.Printf("[disconnect] Reply Error: %v", err)
	}
	return nil
}

func connection(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	if msg.Chat.Type != "private" {
		return nil
	}

	chatIds := isConnected(b, ctx, msg.From.Id)
	if chatIds == nil || len(chatIds) == 0 {
		return nil
	}

	reply, err := msg.Reply(b, "Please wait, fetching connection info...\n\nif you have many chats, This may take a few minutes. Please be patient.", helpers.Shtml())
	if err != nil {
		log.Printf("[connection] Reply Error: %v", err)
		return err
	}

	var text string
	var TotalUsers int64
	if helpers.Contains(config.FakeDevs, msg.From.Id) {
		text = "<b>You are currently connected to the following chats:</b>\n\n"
		for i, chatId := range chatIds {
			var link string
			getChat := onlyAdmins.GetChatCache(chatId)
			if getChat.ChatInfo.Username != "" {
				link = fmt.Sprintf("https://t.me/%s", getChat.ChatInfo.Username)
			} else if strings.HasPrefix(strconv.FormatInt(chatId, 10), "-100") {
				link = fmt.Sprintf("https://t.me/c/%d/1", chatId*-1-1000000000000)
			} else {
				link = fmt.Sprintf("https://t.me/%d", chatId)
			}

			userCount, err := b.GetChatMemberCount(chatId, nil)
			if err != nil {
				log.Printf("[connection] GetChatMemberCount Error: %v", err)
				userCount = 00
			}

			TotalUsers += userCount
			text += fmt.Sprintf(
				"%d. <b><a href='%s'>%s</a></b>\nChat ID: <code>%d</code>\nMembers: <code>%d</code>\n\n",
				i+1,                    // Numbering starts from 1
				link,                   // Chat link
				getChat.ChatInfo.Title, // Chat title
				chatId,                 // Chat ID
				userCount,              // User count
			)

			time.Sleep(200 * time.Millisecond)
		}

		text += fmt.Sprintf("<b>Total Users:</b> <code>%d</code>", TotalUsers)
		_, _, _ = reply.EditText(b, text, &gotgbot.EditMessageTextOpts{ParseMode: "HTML", LinkPreviewOptions: &gotgbot.LinkPreviewOptions{IsDisabled: true}})
		return err
	}

	text += "<b>You are currently connected to the following chats:</b>\n\n"
	for i, chatId := range chatIds {
		getChat := onlyAdmins.GetChatCache(chatId)
		if !getChat.Cached {
			time.Sleep(200 * time.Millisecond)
			getChat = onlyAdmins.LoadChatCache(b, chatId)
		}

		if !getChat.Cached {
			db.DisconnectId(msg.From.Id, chatId)
			continue
		}

		var link string

		// Determine the chat link
		if getChat.ChatInfo.Username != "" {
			link = fmt.Sprintf("https://t.me/%s", getChat.ChatInfo.Username)
		} else if strings.HasPrefix(strconv.FormatInt(chatId, 10), "-100") {
			link = fmt.Sprintf("https://t.me/c/%d/1", chatId*-1-1000000000000)
		} else {
			link = fmt.Sprintf("https://t.me/%d", chatId)
		}

		// Add numbered chat info with a hyperlink for the title
		text += fmt.Sprintf(
			"%d. <b><a href='%s'>%s</a></b>\nChat ID: <code>%d</code>\n\n",
			i+1,                    // Numbering starts from 1
			link,                   // Chat link
			getChat.ChatInfo.Title, // Chat title
			chatId,                 // Chat ID
		)
	}

	if text == "<b>You are currently connected to the following chats:</b>\n\n" {
		text = "⚠️ You are not connected to any chats at the moment."
	}

	_, _, _ = reply.EditText(b, text, &gotgbot.EditMessageTextOpts{ParseMode: "HTML", LinkPreviewOptions: &gotgbot.LinkPreviewOptions{IsDisabled: true}})
	return err
}

const (
	CHATID = "conn_chat_id"
)

func noCommands(msg *gotgbot.Message) bool {
	if msg.ForwardOrigin != nil && message.Text(msg) && !message.Command(msg) {
		return true
	}
	return false
}

func askChatID(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	textChatNotFound := "Chat not found.\nForward a message where i am an admin to get the chat ID."
	textUserNotAdmin := "You are not an admin in this chat.\nForward a message where you & i are admin to get the chat ID."
	if msg.ForwardOrigin != nil {
		origin := msg.ForwardOrigin.MergeMessageOrigin()
		if origin.Type == "channel" {
			chatId := origin.Chat.Id
			getChat := onlyAdmins.GetChatCache(chatId)
			if !getChat.Cached {
				time.Sleep(200 * time.Millisecond)
				getChat = onlyAdmins.LoadChatCache(b, chatId)
			}
			if !getChat.Cached {
				_, _ = msg.Reply(b, textChatNotFound, helpers.Shtml())
				return handlers.NextConversationState(CHATID)
			}

			cached, _ := onlyAdmins.IsUserAdmin(chatId, msg.From.Id)
			if !cached {
				time.Sleep(100 * time.Millisecond)
				if admins := onlyAdmins.LoadAdminCache(b, chatId); !admins.Cached {
					_, _ = msg.Reply(b, textUserNotAdmin, helpers.Shtml())
					return handlers.NextConversationState(CHATID)
				}
			}

			if _, isUserAdmin := onlyAdmins.IsUserAdmin(chatId, msg.From.Id); !isUserAdmin {
				_, _ = msg.Reply(b, textUserNotAdmin, helpers.Shtml())
				return handlers.NextConversationState(CHATID)
			}

			db.ConnectId(msg.From.Id, chatId)
			_, _ = msg.Reply(b, fmt.Sprint("You are now connected to ", getChat.ChatInfo.Title, " (", chatId, ")"), helpers.Shtml())
			return handlers.EndConversation()

		} else if origin.Type == "chat" {
			chatId := origin.Chat.Id
			getChat := onlyAdmins.GetChatCache(chatId)
			if !getChat.Cached {
				time.Sleep(200 * time.Millisecond)
				getChat = onlyAdmins.LoadChatCache(b, chatId)
			}
			if !getChat.Cached {
				_, _ = msg.Reply(b, textChatNotFound, helpers.Shtml())
				return handlers.NextConversationState(CHATID)
			}

			cached, _ := onlyAdmins.IsUserAdmin(chatId, msg.From.Id)
			if !cached {
				time.Sleep(100 * time.Millisecond)
				if admins := onlyAdmins.LoadAdminCache(b, chatId); !admins.Cached {
					_, _ = msg.Reply(b, textUserNotAdmin, helpers.Shtml())
					return handlers.NextConversationState(CHATID)
				}
			}

			if _, isUserAdmin := onlyAdmins.IsUserAdmin(chatId, msg.From.Id); !isUserAdmin {
				_, _ = msg.Reply(b, textUserNotAdmin, helpers.Shtml())
				return handlers.NextConversationState(CHATID)
			}

			_, _ = msg.Reply(b, fmt.Sprint("You are now connected to ", getChat.ChatInfo.Title, " (", chatId, ")"), helpers.Shtml())
			return handlers.EndConversation()
		} else {
			_, _ = msg.Reply(b, "Please forward a message from a chat so I can get the chat ID.", nil)
			return handlers.NextConversationState(CHATID)
		}

	}
	_, _ = msg.Reply(b, "Please forward a message from a chat so I can get the chat ID.", nil)
	return handlers.NextConversationState(CHATID)
}

// cancel ends the conversation.
func cancel(_ *gotgbot.Bot, _ *ext.Context) error {
	log.Printf("Conversation canceled")
	return handlers.EndConversation()
}
