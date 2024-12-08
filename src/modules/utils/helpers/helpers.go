package helpers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"strconv"
	"strings"
	"time"
)

func ToInt64(s string) int64 {
	i, _ := strconv.ParseInt(s, 10, 64)
	return i
}

func ToInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func Contains(slice []int64, value int64) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

func Shtml() *gotgbot.SendMessageOpts {
	return &gotgbot.SendMessageOpts{
		ParseMode:          "HTML",
		ReplyParameters:    &gotgbot.ReplyParameters{AllowSendingWithoutReply: true},
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{IsDisabled: true},
	}
}
func GenerateUniqueString() string {
	// Get the current timestamp in milliseconds (for a shorter format)
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)

	// Generate a random string (optional, but ensures uniqueness)
	randomBytes := make([]byte, 4) // 4 bytes = 8 characters in hexadecimal
	_, err := rand.Read(randomBytes)
	if err != nil {
		// Handle error if random generation fails
		fmt.Printf("Failed to generate random string: %v\n", err)
		return ""
	}
	randomString := hex.EncodeToString(randomBytes)

	// Combine timestamp and random string to create a unique short ID
	uniqueString := fmt.Sprintf("%d%s", timestamp, randomString)
	return uniqueString
}

func GetMessageLink(chatId, messageId int64) string {
	chatIdStr := strconv.FormatInt(chatId, 10)
	if strings.HasPrefix(chatIdStr, "-100") {
		chatIdStr = chatIdStr[4:]
	}

	return fmt.Sprintf("https://t.me/c/%s/%d", chatIdStr, messageId)
}
