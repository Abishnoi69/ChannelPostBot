package helpers

import (
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
)

func PostButton(postId string) gotgbot.InlineKeyboardMarkup {
	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "Delete Post", CallbackData: fmt.Sprintf("delete.%s", postId)},
			},
			{
				{Text: "Repost Post", CallbackData: fmt.Sprintf("repost.%s", postId)},
			},
		},
	}
}
