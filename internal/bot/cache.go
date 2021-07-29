package bot

import (
	"errors"
	"fmt"
	"github.com/psyg1k/remindertelbot/internal"
)

type Cache interface {
	GetUser(int) (internal.User, error)
	UpdateChatTz(int64, string) error
	GetChat(int64) (internal.Chat, error)
}

var NotFoundError = errors.New("not found")

type MemCache map[interface{}]interface{}

func NewCache() MemCache {
	return make(map[interface{}]interface{})
}

func (m MemCache) GetUser(userId int) (user internal.User, err error) {
	val, ok := m[userId]
	if !ok {
		return user, fmt.Errorf("user not fount")
	}
	return val.(internal.User), nil
}

func (m MemCache) GetChat(chatID int64) (chat internal.Chat, err error) {
	val, ok := m[chatID]
	if !ok {
		return chat, fmt.Errorf("user not fount")
	}
	return val.(internal.Chat), nil
}

func (m MemCache) UpdateChatTz(chatID int64, location string) error {
	val, ok := m[chatID]
	if !ok {
		return fmt.Errorf("chat not found")
	}
	chat := val.(internal.Chat)
	chat.Loc = location
	m[chatID] = chat
	return nil
}

func (b *Bot) GetChat(chatId int64) (chat internal.Chat, err error) {
	chat, err = b.Cache.GetChat(chatId)
	if err != nil {
		chat, err = b.db.GetChat(chatId)
	}
	return chat, err
}

func (b *Bot) GetUser(userId int) (user internal.User, err error) {
	user, err = b.Cache.GetUser(userId)
	if err != nil {
		user, err = b.db.GetUser(userId)
	}
	return user, err
}

func (b *Bot) UpdateTz(chatId int64, loc string) error {
	err := b.Cache.UpdateChatTz(chatId, loc)
	err = b.db.UpdateChatTz(chatId, loc)
	return err
}
