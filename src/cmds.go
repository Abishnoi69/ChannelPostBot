package src

import (
	"AshokShau/channelManager/src/db"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
)

// AddCommand adds multiple commands with alias support and checks if the user is banned before processing
func AddCommand(dispatcher *ext.Dispatcher, aliases []string, r handlers.Response) {
	for _, alias := range aliases {
		originalResponse := r
		command := handlers.NewCommand(alias, func(b *gotgbot.Bot, ctx *ext.Context) error {
			if db.IsUserBanned(ctx.EffectiveUser.Id) {
				return nil
			}
			return originalResponse(b, ctx)
		})
		command.Triggers = []rune{'/', '!'}
		dispatcher.AddHandler(command)
	}
}
