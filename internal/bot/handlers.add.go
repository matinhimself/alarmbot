package bot

import (
	"fmt"
	"github.com/jalaali/go-jalaali"
	"github.com/psyg1k/remindertelbot/internal"
	"github.com/tucnak/tr"
	tb "gopkg.in/tucnak/telebot.v2"
	"log"
	"regexp"
	"strings"
	"time"
)

func ParseParam(parts []string, rem *internal.Reminder, t time.Time, isJalali bool, loc *time.Location) error {
	if strings.ToLower(parts[0]) == "repeat" {
		rem = internal.WithRepeat(rem)
		return nil
	}

	if duration, err := time.ParseDuration(parts[0]); err == nil {
		rem.AtTime = t.Add(duration)
		return nil
	}

	var date time.Time
	var err error

	if isJalali {
		date, err = stringToJalaliDate(parts[0], loc)
	} else {
		date, err = stringToDate(parts[0], loc)
	}

	if err != nil {
		return InvalidTimeFormat
	}

	rem.AtTime = date

	return nil
}
func ParseAddCommand(command string, rem *internal.Reminder, loc string, isJalali bool) error {
	location, err := time.LoadLocation(loc)
	if err != nil {
		log.Println(err)
	}
	t := time.Now().UTC()

	if !StringHasSubmatch(command, "-s") {
		rem.Priority = internal.Audible
	}

	parts := strings.Split(command, " ")
	if len(parts) < 2 {
		return InvalidCommand
	}

	// Parse Command Parameters, Time/duration/repeat
	err = ParseParam(parts[1:], rem, t, isJalali, location)
	if err != nil {
		return err
	}

	param, err := GetParam(command, "-t", "time")
	if err == nil {
		err := formatDateTime(param, &rem.AtTime, location)
		if err != nil {
			return err
		}
	}

	if rem.AtTime.Unix() < t.Unix() {
		return ErrPassedTime
	}

	param, err = GetParam(command, "-f", "from")
	if err == nil {
		d, err := time.ParseDuration(param)
		if err != nil {
			return ErrInvalidFromFormat
		}
		rem.From = rem.AtTime.Add(-1 * d)
	} else if rem.IsRepeated {
		rem.From = t
	}
	re := regexp.MustCompile(`-m "([^"]*)"`)
	names := re.FindStringSubmatch(command)
	if len(names) >= 2 {
		rem.Description = names[1]
	}

	param, err = GetParam(command, "-e", "every")
	if err == nil {
		d, err := time.ParseDuration(param)
		if err != nil || d < 5*time.Second {
			return ErrInvalidEveryFormat
		}
		rem.Every = d
	}
	return nil

}

func (b *Bot) AddCommand(m *tb.Message) {
	var msg int
	if m.IsReply() {
		msg = m.ReplyTo.ID
	}

	chat, err := b.GetChat(m.Chat.ID)
	if err != nil {
		log.Printf("%v %v", err, m.Chat)
	}
	rem := internal.Reminder{
		Message: msg,
		ChatId:  m.Chat.ID,
		Every:   DefaultEvery,
		From:    time.Now().UTC().Add(-1 * DefaultFromDuration),
		Created: time.Now().UTC(),
	}

	err = ParseAddCommand(m.Text, &rem, chat.Loc, chat.IsJalali)
	if err == ErrInvalidEveryFormat {
		b.HandleError(m, tr.Lang(string(chat.Language)).Tr("errors/every_format"))
		return
	} else if err == ErrInvalidFromFormat {
		b.HandleError(m, tr.Lang(string(chat.Language)).Tr("errors/from_format"))
		return
	} else if err == ErrPassedTime {
		b.HandleError(m, tr.Lang(string(chat.Language)).Tr("errors/passed_time"))
		return
	}

	reply, err := b.Reply(m, tr.Lang(string(chat.Language)).Tr("adding_reminder"))

	rem, err = b.db.InsertReminder(rem)
	if err != nil {
		log.Println(err)
	}

	var format string
	if rem.IsRepeated {
		format = tr.Lang(string(chat.Language)).Tr("alarm/add_repeat")
	} else {
		format = tr.Lang(string(chat.Language)).Tr("alarm/add_normal")
	}

	selector := createAlarmSelector(&rem, chat.Language, true, true)
	message := generateFirstAlarmMessage(format, &rem, chat)

	_, _ = b.Edit(reply, message, selector)

	b.AddReminder(&rem)
}

func createAlarmSelector(rem *internal.Reminder, lang internal.Language, forceDelete bool, forceMute bool) *tb.ReplyMarkup {
	selector := &tb.ReplyMarkup{}

	var btnSec tb.Btn
	if rem.Priority == 0 && forceMute {
		btnSec = selector.Data(tr.Lang(string(lang)).Tr("buttons/unmute"), MuteCall, fmt.Sprintf("%s:%s", UnmuteUniqueData, rem.Id.Hex()))
	} else if forceMute {
		btnSec = selector.Data(tr.Lang(string(lang)).Tr("buttons/mute"), MuteCall, fmt.Sprintf("%s:%s", MuteUniqueData, rem.Id.Hex()))
	}

	remaining := rem.AtTime.Sub(time.Now().UTC()).Round(rem.Every)
	if remaining < time.Second || forceDelete {
		btnDlt := selector.Data(tr.Lang(string(lang)).Tr("buttons/delete"), DeleteAlarmCall, fmt.Sprintf("%s", rem.Id.Hex()))
		selector.Inline(
			selector.Row(btnDlt, btnSec),
		)

	} else {
		selector.Inline(
			selector.Row(btnSec),
		)
	}

	return selector
}

func generateFirstAlarmMessage(format string, rem *internal.Reminder, chat internal.Chat) string {
	loc, _ := time.LoadLocation(chat.Loc)
	var t string

	if rem.IsRepeated {
		t = rem.Every.String()
	} else {
		if chat.IsJalali {
			t, _ = jalaali.From(rem.AtTime.In(loc)).JFormat("Mon _2 Jan 06 | 15:04:05")
		} else {
			t = rem.AtTime.In(loc).Format("Mon _2 Jan 06 | 15:04:05")
		}

	}

	return fmt.Sprintf(format, rem.Description, t)
}

func generateAlarmMessage(format string, rem *internal.Reminder, chat internal.Chat) string {
	loc, _ := time.LoadLocation(chat.Loc)
	t := time.Now().In(loc)
	message := rem.Description

	if rem.IsRepeated {
		message = fmt.Sprintf(format, message, rem.Every.String())

	} else {
		remaining := rem.AtTime.Sub(t).Round(rem.Every)

		remainingString := DurationToString(remaining)
		emoji := selectEmoji(remaining)

		message = fmt.Sprintf(format, message, remainingString, emoji)

	}

	return message
}
