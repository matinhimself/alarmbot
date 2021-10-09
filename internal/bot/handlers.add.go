package bot

import (
	"fmt"
	"github.com/jalaali/go-jalaali"
	"github.com/karrick/tparse/v2"
	"github.com/psyg1k/remindertelbot/internal"
	log "github.com/sirupsen/logrus"
	"github.com/tucnak/tr"
	tb "gopkg.in/tucnak/telebot.v2"
	"regexp"
	"strings"
	"time"
)

func ParseAtTime(parts []string, rem *internal.Reminder, t time.Time, isJalali bool, loc *time.Location) error {
	if strings.ToLower(parts[0]) == "repeat" {
		rem.AtTime = t.In(loc)
		rem.IsRepeated = true
		return nil
	}

	if atTime, err := tparse.AddDuration(t, parts[0]); err == nil {
		rem.AtTime = atTime
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
	err = ParseAtTime(parts[1:], rem, t, isJalali, location)
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

	if rem.AtTime.Unix() < t.Unix() && !rem.IsRepeated {
		return ErrPassedTime
	}

	param, err = GetParam(command, "-f", "from")

	if err == nil {
		f, err := tparse.AddDuration(rem.AtTime, fmt.Sprintf("-%s", param))
		if err != nil {
			return ErrInvalidFromFormat
		}
		rem.From = f.In(location)
	} else if !rem.IsRepeated {
		rem.From = rem.AtTime.Add(-1 * DefaultFromDuration)
	}

	re := regexp.MustCompile(`-m "([^"]*)"`)
	names := re.FindStringSubmatch(command)
	if len(names) >= 2 {
		rem.Description = names[1]
	}

	param, err = GetParam(command, "-e", "every")
	if err == nil {
		if rem.IsRepeated {
			weekday, err := parseWeekday(param)
			if err != nil {
				return err
			}

			// Add 24 hours if time is passed
			// to make sure it will start from
			// next week.
			if weekday == rem.AtTime.In(location).Weekday() && rem.AtTime.In(location).Before(t.In(location)) {
				rem.AtTime = rem.AtTime.Add(24 * time.Hour)
			}
			rem.AtTime = ClosestDayOfWeek(rem.AtTime, weekday)
			rem.Every = 24 * time.Hour * 7

		} else {
			d, err := tparse.AbsoluteDuration(t, param)
			if err != nil || d < 5*time.Second {
				return ErrInvalidEveryFormat
			}
			rem.Every = d
		}

	}
	return nil

}

func (b *Bot) HelpCommand(m *tb.Message) {
	c, err := b.GetChat(m.Chat.ID)
	var message string
	if err != nil {
		message = tr.Lang("en").Tr("commands/help")
	} else {
		message = tr.Lang(string(c.Language)).Tr("commands/help")
	}
	_, err = b.Reply(m, message)
	if err != nil {
		log.Error(err)
	}
}

func (b *Bot) AddCommand(m *tb.Message) {
	var msg int

	chat, err := b.GetChat(m.Chat.ID)
	if err != nil {
		log.WithField("chat", m.Chat.ID).Errorf("%v", err)
	}

	log.WithField("chat", m.Chat.ID).Info("Adding reminder")
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
	} else if err != nil {
		b.HandleError(m, err.Error())
		return
	}
	reply, _ := b.Reply(m, tr.Lang(string(chat.Language)).Tr("commands/adding_reminder"))

	if m.IsReply() {
		rem.Message = m.ReplyTo.ID
	} else {
		rem.Message = reply.ID
	}

	rem, err = b.db.InsertReminder(rem)
	if err != nil {
		log.Error(err)
		return
	}

	log.WithFields(map[string]interface{}{
		"for":   rem.AtTime,
		"every": rem.Every,
		"from":  rem.From,
		"chat":  rem.ChatId,
	}).Info("Reminder added to database.")

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
		btnDlt := selector.Data(tr.Lang(string(lang)).Tr("buttons/delete"), DeleteAlarmCall, rem.Id.Hex())
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
		remString, _ := DurationToString(rem.Every)
		message = fmt.Sprintf(format, message, remString)

	} else {
		remaining := rem.AtTime.Sub(t).Round(rem.Every)

		remainingString, now := DurationToString(remaining)
		if now {
			remainingString = tr.Lang(string(chat.Language)).Tr("alarm/now")
		}
		emoji := selectEmoji(remaining)

		message = fmt.Sprintf(format, message, remainingString, emoji)

	}

	return message
}
