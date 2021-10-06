package bot

import (
	"fmt"
	"log"
	"strings"
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

func ClosestDayOfWeek(atTime time.Time, day time.Weekday) time.Time {
	fmt.Println(atTime.Weekday())
	if atTime.Weekday() == day {
		return atTime
	}
	return ClosestDayOfWeek(atTime.Add(24*time.Hour), day)
}

var daysOfWeek = map[string]time.Weekday{
	"saturday": time.Saturday,
	"sat":      time.Saturday,

	"sunday": time.Sunday,
	"sun":    time.Sunday,

	"monday": time.Monday,
	"mon":    time.Monday,

	"tuesday": time.Tuesday,
	"tue":     time.Tuesday,

	"wednesday": time.Wednesday,
	"wed":       time.Wednesday,

	"thursday": time.Thursday,
	"thu":      time.Thursday,

	"friday": time.Friday,
	"fri":    time.Friday,
}

func parseWeekday(v string) (time.Weekday, error) {
	vs := strings.ToLower(v)
	if d, ok := daysOfWeek[vs]; ok {
		return d, nil
	}

	return time.Sunday, fmt.Errorf("invalid weekday '%s'", v)
}
