package onlyAdmins

import (
	"AshokShau/channelManager/src/config"
	"context"
	"log"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/dgraph-io/ristretto"
	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/marshaler"
	"github.com/eko/gocache/lib/v4/store"
	redisstore "github.com/eko/gocache/store/redis/v4"
	ristrettostore "github.com/eko/gocache/store/ristretto/v4"
	"github.com/redis/go-redis/v9"
)

var (
	ctx         = context.Background()
	Marshal     *marshaler.Marshaler
	redisClient *redis.Client
)

type AdminCache struct {
	ChatId   int64
	UserInfo []gotgbot.MergedChatMember
	Cached   bool
}

type ChatCache struct {
	ChatId   int64
	ChatInfo gotgbot.Chat
	Cached   bool
}

var expireTime = 20 * time.Minute

func init() {
	opt, err := redis.ParseURL(config.RedisURI)
	if err != nil {
		log.Fatalf("failed to parse redis url: %v", err)
	}

	redisClient = redis.NewClient(opt)
	if err = redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("failed to ping redis: %v", err)
	}

	ristrettoCache, err := ristretto.NewCache(&ristretto.Config{NumCounters: 1000, MaxCost: 100, BufferItems: 64})
	if err != nil {
		log.Fatalf("failed to create ristretto cache: %v", err)
	}

	cacheManager := cache.NewChain[any](
		cache.New[any](ristrettostore.NewRistretto(ristrettoCache)),
		cache.New[any](redisstore.NewRedis(redisClient)),
	)

	Marshal = marshaler.New(cacheManager)
}

func CloseRedis() {
	if err := redisClient.Close(); err != nil {
		log.Printf("error closing Redis client: %v", err)
	}
	log.Printf("redis client closed")
}

func LoadAdminCache(b *gotgbot.Bot, chatId int64) AdminCache {
	adminList, err := b.GetChatAdministrators(chatId, nil)
	if err != nil {
		log.Printf("LoadAdminCache: error getting admin list: %v", err)
		return AdminCache{}
	}

	userList := make([]gotgbot.MergedChatMember, len(adminList))
	for i, admin := range adminList {
		userList[i] = admin.MergeChatMember()
	}

	err = Marshal.Set(ctx, AdminCache{ChatId: chatId}, AdminCache{ChatId: chatId, UserInfo: userList, Cached: true}, store.WithExpiration(expireTime))
	if err != nil {
		log.Printf("error setting admin list: %v", err)
		return AdminCache{}
	}

	_, newAdminList := GetAdminCacheList(chatId)
	return newAdminList
}

func GetAdminCacheList(chatId int64) (bool, AdminCache) {
	gotAdminList, err := Marshal.Get(ctx, AdminCache{ChatId: chatId}, new(AdminCache))
	if err != nil || gotAdminList == nil {
		return false, AdminCache{}
	}

	return true, *gotAdminList.(*AdminCache)
}

func LoadChatCache(b *gotgbot.Bot, chatId int64) ChatCache {
	fullChat, err := b.GetChat(chatId, nil)
	if err != nil {
		log.Printf("LoadChatCache: error getting chat info: %v", err)
		return ChatCache{}
	}

	chat := fullChat.ToChat()

	chatInfo := &chat

	err = Marshal.Set(ctx, ChatCache{ChatId: chatId}, ChatCache{ChatId: chatId, ChatInfo: *chatInfo, Cached: true}, store.WithExpiration(expireTime))
	if err != nil {
		log.Printf("error setting chat info: %v", err)
		return ChatCache{}
	}

	return GetChatCache(chatId)
}

func GetChatCache(chatId int64) ChatCache {
	gotChatInfo, err := Marshal.Get(ctx, ChatCache{ChatId: chatId}, new(ChatCache))
	if err != nil || gotChatInfo == nil {
		return ChatCache{}
	}

	return *gotChatInfo.(*ChatCache)
}
