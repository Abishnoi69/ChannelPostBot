package helpers

import (
	"AshokShau/channelManager/src/db"
	"fmt"
	tgmd2html "github.com/PaulSonOfLars/gotg_md2html"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"regexp"
	"strings"
)

// RevertButtons converts []db.Button to a string
func RevertButtons(buttons []db.Button) string {
	var res string
	for _, btn := range buttons {
		format := "\n[%s](buttonurl://%s"
		if btn.SameLine {
			format += ":same"
		}
		format += ")"
		res += fmt.Sprintf(format, btn.Name, btn.Url)
	}
	return res
}

// BuildKeyboard builds a keyboard from a list of buttons.
func BuildKeyboard(buttons []db.Button) [][]gotgbot.InlineKeyboardButton {
	keyB := make([][]gotgbot.InlineKeyboardButton, 0)
	for _, btn := range buttons {
		if btn.SameLine && len(keyB) > 0 {
			keyB[len(keyB)-1] = append(keyB[len(keyB)-1], gotgbot.InlineKeyboardButton{Text: btn.Name, Url: btn.Url})
		} else {
			k := make([]gotgbot.InlineKeyboardButton, 1)
			k[0] = gotgbot.InlineKeyboardButton{Text: btn.Name, Url: btn.Url}
			keyB = append(keyB, k)
		}
	}
	return keyB
}

// ConvertButtonV2ToDbButton converts a list of ButtonV2 to a list of db.Button.
func ConvertButtonV2ToDbButton(buttons []tgmd2html.ButtonV2) (btnS []db.Button) {
	btnS = make([]db.Button, len(buttons))
	for i, btn := range buttons {
		btnS[i] = db.Button{
			Name:     btn.Name,
			Url:      btn.Content,
			SameLine: btn.SameLine,
		}
	}
	return
}

// InlineKeyboardMarkupToTgmd2htmlButtonV2 converts gotgbot.InlineKeyboardMarkup to []tgmd2html.ButtonV2.
func InlineKeyboardMarkupToTgmd2htmlButtonV2(replyMarkup *gotgbot.InlineKeyboardMarkup) (btns []tgmd2html.ButtonV2) {
	for _, inlineKeyboard := range replyMarkup.InlineKeyboard {
		for i, button := range inlineKeyboard {
			if button.Url == "" {
				continue
			}
			sameline := i != 0
			btns = append(btns, tgmd2html.ButtonV2{
				Name:     button.Text,
				Content:  button.Url,
				SameLine: sameline,
			})
		}
	}
	return
}

// preFixes checks the message before saving it to a database.
func preFixes(buttons []tgmd2html.ButtonV2, defaultNameButton string, text *string, dataType *int, fileid string, dbButtons *[]db.Button, errorMsg *string) {
	if *dataType == db.TEXT && len(*text) > 4096 {
		*dataType = -1
		*errorMsg = fmt.Sprintf("Your message text is %d characters long. The maximum length for text is 4096; please trim it to a smaller size. Note that markdown characters may take more space than expected.", len(*text))
	} else if *dataType != db.TEXT && len(*text) > 1024 {
		*dataType = -1
		*errorMsg = fmt.Sprintf("Your message caption is %d characters long. The maximum caption length is 1024; please trim it to a smaller size. Note that markdown characters may take more space than expected.", len(*text))
	} else {
		for i, button := range buttons {
			if button.Name == "" {
				buttons[i].Name = defaultNameButton
			}
		}

		buttonUrlFixer := func(_buttons *[]tgmd2html.ButtonV2) {
			buttonUrlPattern, _ := regexp.Compile(`[(htps)?:/w.a-zA-Z\d@%_+~#=]{2,256}\.[a-z]{2,6}\b([-a-zA-Z\d@:%_+.~#?&/=]*)`)
			buttons = *_buttons
			for i, btn := range *_buttons {
				if !buttonUrlPattern.MatchString(btn.Content) {
					buttons = append(buttons[:i], buttons[i+1:]...)
				}
			}
			*_buttons = buttons
		}

		buttonUrlFixer(&buttons)
		*dbButtons = ConvertButtonV2ToDbButton(buttons)

		*text = strings.Trim(*text, "\n\t\r ")
		if *text == "" && fileid == "" {
			*dataType = -1
		}
	}
}

// function used to get rawtext from gotgbot.Message
func setRawText(msg *gotgbot.Message, args []string, rawText *string) {
	replyMsg := msg.ReplyToMessage
	if replyMsg == nil {
		if msg.Text == "" && msg.Caption != "" {
			*rawText = strings.SplitN(msg.OriginalCaptionMDV2(), " ", 2)[1]
		} else if msg.Text != "" && msg.Caption == "" {
			*rawText = strings.SplitN(msg.OriginalMDV2(), " ", 2)[1]
		}
	} else {
		if replyMsg.Text == "" && replyMsg.Caption != "" {
			*rawText = replyMsg.OriginalCaptionMDV2()
		} else if replyMsg.Caption == "" && len(args) >= 2 {
			*rawText = strings.SplitN(msg.OriginalMDV2(), " ", 3)[1]
		} else {
			*rawText = replyMsg.OriginalMDV2()
		}
	}
}
func GetMsgType(msg *gotgbot.Message) (text string, dataType int, fileId string, buttons []db.Button, errorMsg string) {
	dataType = -1
	errorMsg = fmt.Sprintf("You need to give me some content to post!")
	var (
		rawText string
		args    = strings.Fields(msg.Text)[1:]
	)

	_buttons := make([]tgmd2html.ButtonV2, 0)
	replyMsg := msg.ReplyToMessage

	// set rawText from helper function
	setRawText(msg, args, &rawText)

	if len(args) >= 1 && msg.ReplyToMessage == nil {
		fileId = ""
		text, _buttons = tgmd2html.MD2HTMLButtonsV2(rawText)
		dataType = db.TEXT
	} else if msg.ReplyToMessage != nil {
		if replyMsg.ReplyMarkup == nil {
			text, _buttons = tgmd2html.MD2HTMLButtonsV2(rawText)
		} else {
			text, _ = tgmd2html.MD2HTMLButtonsV2(rawText)
			_buttons = InlineKeyboardMarkupToTgmd2htmlButtonV2(replyMsg.ReplyMarkup)
		}
		if len(args) == 0 && replyMsg.Text != "" {
			dataType = db.TEXT
		} else if replyMsg.Sticker != nil {
			fileId = replyMsg.Sticker.FileId
			dataType = db.STICKER
			// Extract buttons from args when the message is a sticker
			if len(args) > 0 {
				_, _buttons = tgmd2html.MD2HTMLButtonsV2(strings.Join(args, " "))
			}
		} else if replyMsg.Document != nil {
			fileId = replyMsg.Document.FileId
			dataType = db.DOCUMENT
		} else if len(replyMsg.Photo) > 0 {
			fileId = replyMsg.Photo[len(replyMsg.Photo)-1].FileId
			dataType = db.PHOTO
		} else if replyMsg.Audio != nil {
			fileId = replyMsg.Audio.FileId
			dataType = db.AUDIO
		} else if replyMsg.Voice != nil {
			fileId = replyMsg.Voice.FileId
			dataType = db.VOICE
		} else if replyMsg.Video != nil {
			fileId = replyMsg.Video.FileId
			dataType = db.VIDEO
		} else if replyMsg.VideoNote != nil {
			fileId = replyMsg.VideoNote.FileId
			dataType = db.VideoNote
		} else if replyMsg.Animation != nil {
			fileId = replyMsg.Animation.FileId
			dataType = db.GIF
		}
	}

	// pre-fix the data before sending it back
	preFixes(_buttons, "Button", &text, &dataType, fileId, &buttons, &errorMsg)
	return
}
