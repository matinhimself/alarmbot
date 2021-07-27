package bot

import (
	"log"
	"time"
)

func (b *Bot) LoadReminders() error {

	reminders, err := b.db.GetRemindersAfter(time.Now())
	if err != nil {
		log.Println(err)
		return err
	}

	for _, r := range reminders {
		b.AddReminder(&r)
	}
	return nil

}
