package modules

import (
	"AshokShau/channelManager/src/db"
	"AshokShau/channelManager/src/modules/utils/helpers"
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"strings"
)

func updateUserSettingHandler(b *gotgbot.Bot, ctx *ext.Context, settingName string, updateFunc func(int64, bool)) error {
	msg := ctx.EffectiveMessage
	if msg.Chat.Type != "private" {
		return nil
	}

	args := ctx.Args()[1:]
	user := msg.From
	getUser := db.GetUserSettings(user.Id)

	// Define current setting value based on the setting name
	var currentSetting bool
	switch settingName {
	case "ForwardTag":
		currentSetting = getUser.ForwardTag
	case "NoNotif":
		currentSetting = getUser.NoNotif
	case "Protect":
		currentSetting = getUser.Protect
	case "Spoiler":
		currentSetting = getUser.Spoiler
	case "WebPreview":
		currentSetting = getUser.WebPreview
	case "CaptionAbove":
		currentSetting = getUser.CaptionAbove
	default:
		_, _ = msg.Reply(b, "Invalid setting specified.", helpers.Shtml())
		return nil
	}

	// If no arguments, show current status
	if len(args) < 1 {
		text := fmt.Sprintf("Please provide args <code> y/yes/true/on/n/no/false/off </code> to update the %s setting.\nUsage: <code>!%s args</code>\n\nCurrent setting: %t",
			settingName, strings.ToLower(settingName), currentSetting)
		_, err := msg.Reply(b, text, helpers.Shtml())
		return err
	}

	// Update the setting based on the argument
	if args[0] == "y" || args[0] == "yes" || args[0] == "true" || args[0] == "on" {
		updateFunc(user.Id, true)
		_, _ = msg.Reply(b, fmt.Sprintf("%s has been <b>enabled</b>.", settingName), helpers.Shtml())
	} else if args[0] == "n" || args[0] == "no" || args[0] == "false" || args[0] == "off" {
		updateFunc(user.Id, false)
		_, _ = msg.Reply(b, fmt.Sprintf("%s has been <b>disabled</b>.", settingName), helpers.Shtml())
	} else {
		_, _ = msg.Reply(b, fmt.Sprintf("Invalid argument. Please provide <code> y/yes/true/on/n/no/false/off </code>  to update the %s setting.", settingName), helpers.Shtml())
	}

	return nil
}

func updateForwardTag(b *gotgbot.Bot, ctx *ext.Context) error {
	return updateUserSettingHandler(b, ctx, "ForwardTag", db.UpdateForwardTag)
}

func updateNoNotif(b *gotgbot.Bot, ctx *ext.Context) error {
	return updateUserSettingHandler(b, ctx, "NoNotif", db.UpdateNoNotif)
}

func updateProtect(b *gotgbot.Bot, ctx *ext.Context) error {
	return updateUserSettingHandler(b, ctx, "Protect", db.UpdateProtect)
}

func updateSpoiler(b *gotgbot.Bot, ctx *ext.Context) error {
	return updateUserSettingHandler(b, ctx, "Spoiler", db.UpdateSpoiler)
}

func updateWebPreview(b *gotgbot.Bot, ctx *ext.Context) error {
	return updateUserSettingHandler(b, ctx, "WebPreview", db.UpdateWebPreview)
}

func updateCaptionAbove(b *gotgbot.Bot, ctx *ext.Context) error {
	return updateUserSettingHandler(b, ctx, "CaptionAbove", db.UpdateCaptionAbove)
}

func resetSettings(b *gotgbot.Bot, ctx *ext.Context) error {
	db.ResetUserSettings(ctx.EffectiveMessage.From.Id)
	_, _ = ctx.EffectiveMessage.Reply(b, "All settings have been reset.", helpers.Shtml())
	return nil
}
