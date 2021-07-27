package bot

import (
	"fmt"
	"github.com/psyg1k/remindertelbot/internal"
)

type Cache interface {
	GetUser(int) (internal.User, error)
	UpdateUserTz(int, internal.Offset) error
}

type NotFoundError struct{}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("couldn't find document")
}

type MemCache map[int]internal.User

func NewCache() MemCache {
	return make(map[int]internal.User)
}

func (m MemCache) UpdateUserTz(userId int, tz internal.Offset) error {
	val, ok := m[userId]
	if !ok {
		return fmt.Errorf("user not found")
	}
	val.Offset = tz
	return nil
}

func (m MemCache) GetUser(userId int) (user internal.User, err error) {
	val, ok := m[userId]
	if !ok {
		return user, fmt.Errorf("user not fount")
	}
	return val, nil
}

func (b *Bot) GetUser(userId int) (user internal.User, err error) {
	user, err = b.Cache.GetUser(userId)
	if err != nil {
		user, err = b.db.GetUser(userId)
	}
	return user, err
}

func (b *Bot) UpdateTz(userId int, tz internal.Offset) error {
	err := b.Cache.UpdateUserTz(userId, tz)
	if err != nil {
	}

	err = b.db.UpdateUserTz(tz, userId)
	return err
}
