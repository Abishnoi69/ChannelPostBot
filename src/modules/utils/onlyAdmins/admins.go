package onlyAdmins

import (
	"AshokShau/channelManager/src/config"
)

const (
	groupAnonymousBot = 1087968824
	tgAdmin           = 777000
)

var tgAdminList = []int64{groupAnonymousBot, config.OwnerId, tgAdmin}

// IsUserAdmin returns a boolean indicating whether the user is an admin
// or not and whether the check was done from the cache or not.
//
// If the user is an admin (i.e. the first return value is true), the second
// return value will be true if the check was done from the cache, or false
// if the check was done from the Telegram API.
func IsUserAdmin(chatId, userId int64) (bool, bool) {
	//for _, admin := range tgAdminList {
	//	if admin == userId {
	//		return true, true
	//	}
	//}

	adminsAvail, admins := GetAdminCacheList(chatId)
	if !admins.Cached || !adminsAvail {
		return false, false
	}

	for _, admin := range admins.UserInfo {
		if admin.User.Id == userId {
			return true, true
		}
	}

	return true, false
}
