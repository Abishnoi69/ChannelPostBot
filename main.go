package main

import (
	"AshokShau/channelManager/src/config"
	"AshokShau/channelManager/src/db"
	"AshokShau/channelManager/src/modules"
	"AshokShau/channelManager/src/modules/utils/onlyAdmins"
	"fmt"
	"log"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

const secretToken = "idkWhatIsThis"

var allowedUpdates = []string{"message", "callback_query", "my_chat_member", "inline_query"}

func initBot() (*gotgbot.Bot, *ext.Updater, error) {
	if config.Token == "" {
		return nil, nil, fmt.Errorf("TOKEN is not provided")
	}

	bot, err := gotgbot.NewBot(config.Token, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create new bot: %w", err)
	}

	updater := ext.NewUpdater(modules.Dispatcher, nil)
	return bot, updater, nil
}

func configureWebhook(bot *gotgbot.Bot, updater *ext.Updater) error {
	if config.WebhookUrl == "" {
		return fmt.Errorf("WEBHOOK_URL is not provided")
	}

	_, err := bot.SetWebhook(config.WebhookUrl+config.Token, &gotgbot.SetWebhookOpts{
		MaxConnections:     40,
		DropPendingUpdates: true,
		AllowedUpdates:     allowedUpdates,
		SecretToken:        secretToken,
		RequestOpts: &gotgbot.RequestOpts{
			Timeout: 20 * time.Second,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to set webhook: %w", err)
	}

	return updater.StartWebhook(bot, config.Token, ext.WebhookOpts{
		ListenAddr:  "0.0.0.0:" + config.Port,
		SecretToken: secretToken,
		ReadTimeout: 20 * time.Second,
	})
}

func startPolling(bot *gotgbot.Bot, updater *ext.Updater) error {
	_, _ = bot.DeleteWebhook(nil)
	return updater.StartPolling(bot, &ext.PollingOpts{
		DropPendingUpdates: true,
		GetUpdatesOpts:     &gotgbot.GetUpdatesOpts{AllowedUpdates: allowedUpdates},
	})
}

func main() {
	bot, updater, err := initBot()
	if err != nil {
		log.Fatalf("Initialization error: %s", err)
	}
	
	mode := "Webhook"
	if err = configureWebhook(bot, updater); err != nil {
		log.Printf("Webhook configuration failed: %s", err)
		mode = "Polling"
		if err = startPolling(bot, updater); err != nil {
			log.Fatalf("Polling start failed: %s", err)
		}
	}

	log.Printf("Bot has been started as %s[%s] using %s", bot.FirstName, bot.Username, mode)

	updater.Idle()
	log.Printf("Bot has been stopped")
	db.Close()
	onlyAdmins.CloseRedis()
	log.Printf("Bye!")
}
