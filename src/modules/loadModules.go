package modules

import (
	"AshokShau/channelManager/src"
	"AshokShau/channelManager/src/config"
	"encoding/json"
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/conversation"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/inlinequery"
	"html"
	"log"

	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

var Dispatcher = newDispatcher()

func newDispatcher() *ext.Dispatcher {
	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{Error: errorHandler, MaxRoutines: 50})
	loadModules(dispatcher)
	return dispatcher
}

func loadModules(dispatcher *ext.Dispatcher) {
	modulesToLoad := []func(*ext.Dispatcher){
		loadCallbacks, loadPost, loadSettings,
	}
	for _, module := range modulesToLoad {
		module(dispatcher)
	}
	log.Printf("Loaded %d modules\n", len(modulesToLoad))
}

func errorHandler(bot *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
	var msg string
	if ctx.Update != nil {
		if updateBytes, err := json.MarshalIndent(ctx.Update, "", "  "); err == nil {
			msg = fmt.Sprintf("%s", html.EscapeString(string(updateBytes)))
		} else {
			msg = "failed to marshal update"
		}
	} else {
		msg = "no update"
	}

	message := fmt.Sprintf("<blockquote expandable>New Error:\n%s\n\n%s</blockquote>", err.Error(), msg)
	if _, err = bot.SendMessage(config.LoggerId, message, nil); err != nil {
		log.Printf(message)
		_, _ = bot.SendMessage(config.LoggerId, err.Error(), nil)
		return ext.DispatcherActionNoop
	}

	return ext.DispatcherActionNoop
}

func loadCallbacks(d *ext.Dispatcher) {
	d.AddHandler(handlers.NewCallback(callbackquery.Prefix("send."), sendPostCallback))
	d.AddHandler(handlers.NewCallback(callbackquery.Prefix("delete."), deletePostCallback))
	d.AddHandler(handlers.NewCallback(callbackquery.Prefix("repost."), repostCallback))
}

func loadPost(d *ext.Dispatcher) {
	src.AddCommand(d, []string{"del", "delete"}, deletePost)
	src.AddCommand(d, []string{"delAll", "deleteAll"}, delAllPosts)
	src.AddCommand(d, []string{"create", "new", "share"}, createPost)
	src.AddCommand(d, []string{"getPost", "get"}, getPost)
	src.AddCommand(d, []string{"send", "post"}, sendPost)
	src.AddCommand(d, []string{"report"}, repost)
	src.AddCommand(d, []string{"edit"}, editPost)
	src.AddCommand(d, []string{"repost"}, repost)
	src.AddCommand(d, []string{"start"}, start)
	src.AddCommand(d, []string{"help"}, help)

	src.AddCommand(d, []string{"ban"}, banUser)
	src.AddCommand(d, []string{"unban"}, unbanUser)
	src.AddCommand(d, []string{"bans"}, getBans)
	src.AddCommand(d, []string{"broadcast"}, broadcast)
	src.AddCommand(d, []string{"stats"}, stats)

	src.AddCommand(d, []string{"disconnect", "remove"}, disconnect)
	src.AddCommand(d, []string{"connection", "channels"}, connection)
	d.AddHandler(handlers.NewConversation(
		[]ext.Handler{handlers.NewCommand("add", connect)},
		map[string][]ext.Handler{
			CHATID: {handlers.NewMessage(noCommands, askChatID)},
		},
		&handlers.ConversationOpts{
			Exits:        []ext.Handler{handlers.NewCommand("cancel", cancel)},
			StateStorage: conversation.NewInMemoryStorage(conversation.KeyStrategySenderAndChat),
			AllowReEntry: true,
		},
	))

	d.AddHandler(handlers.NewInlineQuery(inlinequery.All, inlineSharePost))
}

func loadSettings(d *ext.Dispatcher) {
	src.AddCommand(d, []string{"forwardTag", "forward"}, updateForwardTag)
	src.AddCommand(d, []string{"noNotif", "notify", "silent"}, updateNoNotif)
	src.AddCommand(d, []string{"protect"}, updateProtect)
	src.AddCommand(d, []string{"spoiler"}, updateSpoiler)
	src.AddCommand(d, []string{"webPreview", "preview"}, updateWebPreview)
	src.AddCommand(d, []string{"captionAbove"}, updateCaptionAbove)
	src.AddCommand(d, []string{"reset"}, resetSettings)
}
