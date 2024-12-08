package modules

import (
	"AshokShau/channelManager/src/config"
	"AshokShau/channelManager/src/db"
	"fmt"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

func start(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	go db.GetUserSettings(msg.From.Id)

	helpText := fmt.Sprintf("Hello, <b>%s</b>! <blockquote>I'm an Advanced channel manager BoT</blockquote>\n\n<blockquote>👉 Features Like Schedule Deleting,Multiple Channels,Repost,Edit Post and More...</blockquote>\n\n<b>Share and Support Us</b>\n\n<b>Use /help for more information.</b>", msg.From.FirstName)
	button := &gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{
					Text: "Add me to your channel",
					Url:  fmt.Sprint("https://t.me/", b.Username, "?startchannel=new"),
				},
			},
			{
				{
					Text: "SUPPORT CHANNEL 🙋‍♀️",
					Url:  config.SupportChat,
				},
			},
		},
	}

	_, _ = msg.Reply(b, helpText, &gotgbot.SendMessageOpts{ParseMode: "HTML", ReplyMarkup: button})
	return nil
}

func help(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	if msg.Chat.Type != "private" {
		return nil
	}
	helpText := `HELLO! I'M AN ADVANCED CHANNEL MANAGER BOT

<b>Connection commands:</b>
<code>/add </code> - Connect to a channel or group

<code>/add chat_id chat_id2 ..</code>: connect to multiple channels
<code>/add</code>: reply to a <b>Forwarded</b> message to connect to a channel or group
<code>/add</code>: bot ask to <b>Forward</b> me the Message of that Channel that you want to Add

<code>/remove</code>:  Disconnect from a channel or multiple channels
<code>!channels</code> - List all connected channels

<b>Post commands:</b>
<code>!del channel_id msg_id</code> - Delete a message from a channel
<code>!create</code> - Create a post or get postId
<code>!get PostId</code> - Share a post in current chat || Get post preview
<code>!send Reply</code> - Send a post to all connected channels
<code>!repost PostId</code> - Re-post a post from all connected channels (del old post and send new post)
<code>!edit PostId</code> - Edit a post from all connected chats

<b>User Settings:</b>
<code>!forward</code> - Toggle forward tag
<code>!silent</code> - Toggle no notification
<code>!protect</code> - Toggle protect
<code>!spoiler</code> - Toggle spoiler
<code>!preview</code> - Toggle web preview
<code>!captionabove</code> - Toggle caption above
<code>!reset</code> - Reset all user settings (Set to default value: off)

<b>Inline Commands:</b>
<code>@%s PostId</code> - Share a post in current chat (Via Inline)

<b>Add Buttons:</b>
Simple buttons:
- The following syntax will create a button called "Google", which will open google.com
-> [Google](buttonurl://google.com)


Buttons on the same line:
- This example creates two buttons ("Google" and "Bing"), which will appear on the same line. This is achieved with the :same tag on the second button.
-> [Google](buttonurl://google.com) [Bing](buttonurl://bing.com:same)
`

	button := &gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{
					Text: "Add me to your channel",
					Url:  fmt.Sprint("https://t.me/", b.Username, "?startchannel=new"),
				},
			},
			{
				{
					Text: "SUPPORT CHANNEL 🙋‍♀️",
					Url:  config.SupportChat,
				},
			},
		},
	}

	_, _ = msg.Reply(b, fmt.Sprintf(helpText, b.Username), &gotgbot.SendMessageOpts{ParseMode: "HTML", ReplyMarkup: button})
	return nil
}
