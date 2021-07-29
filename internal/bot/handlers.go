package bot

import (
	"errors"
	"fmt"
	"github.com/jalaali/go-jalaali"
	"github.com/psyg1k/remindertelbot/internal"
	"github.com/tucnak/tr"
	"go.mongodb.org/mongo-driver/mongo"
	tb "gopkg.in/tucnak/telebot.v2"
	"log"
	"regexp"
	"strings"
	"time"
)

const (
	LangCall        = "lang"
	TzCall          = "tz"
	DeleteAlarmCall = "del"
	MuteAlarmCall   = "mute"
	UnMuteAlarmCall = "mute"

	DefaultEvery        = time.Hour * 12
	DefaultFromDuration = time.Hour * 72
)

var ErrInvalidFromFormat error = errors.New("invalid from format")
var ErrPassedTime error = errors.New("passed time")
var ErrInvalidEveryFormat error = errors.New("invalid every format")

func (b *Bot) SetTzCommand(m *tb.Message) {
	chat, _ := b.GetChat(m.Chat.ID)
	parts := strings.Split(m.Text, " ")
	if len(parts) < 2 {
		b.HandleErrorErr(m, ErrInvalidTz, chat.Language)
		return
	}
	_, err := time.LoadLocation(parts[1])
	if err != nil {
		b.HandleErrorErr(m, ErrInvalidTz, chat.Language)
		return
	}
	err = b.UpdateTz(chat.ChatID, parts[1])
	if err != nil {
		log.Println(err)
	}
	b.Reply(m, tr.Lang(string(chat.Language)).Tr("tz_changed"))
}

func (b *Bot) HandleError(m *tb.Message, s string) {
	_, _ = b.Reply(m, s)
}

func (b *Bot) HandleErrorErr(m *tb.Message, e error, lang internal.Language) {
	if e == ErrInvalidTz {
		_, _ = b.Reply(m, tr.Lang(string(lang)).Tr(fmt.Sprintf("errors/%s", e.Error())))
	}
}

var ErrInvalidTz error = errors.New("invalid_timezone")

func (b *Bot) SetTz(c *tb.Callback) {
	chat, _ := b.GetChat(c.Message.Chat.ID)
	_, err := time.LoadLocation(c.Data)
	if err != nil {
		b.HandleErrorErr(c.Message, err, chat.Language)
	}
	err = b.UpdateTz(c.Message.Chat.ID, c.Data)
	if err != nil {
		log.Println(err)
	}
	_, _ = b.Edit(c.Message, tr.Lang(string(chat.Language)).Tr("registered"))
}

func (b *Bot) SetLanguage(call *tb.Callback) {
	lang := internal.Language(call.Data)

	chat := internal.Chat{
		ChatID:   call.Message.Chat.ID,
		TaskList: 0,
		Loc:      "UTC",
		Language: lang,
		IsJalali: true,
	}

	err := b.db.InsertChat(chat)
	if err != nil {
		log.Println(err)
	}

	selector := &tb.ReplyMarkup{}
	selector.Inline(
		selector.Row(selector.Data("Iran", TzCall, internal.IRAN)),
	)
	_, err = b.Edit(call.Message, tr.Lang(string(lang)).Tr("choose_region"), selector)
	if err != nil {
		log.Println(err)
	}
}

func (b *Bot) ChooseLang(message *tb.Message) {
	selector := &tb.ReplyMarkup{}

	var langRows []tb.Row

	for _, language := range internal.Languages {
		langRows = append(langRows, selector.Row(selector.Data(language.ToString(), LangCall, string(language))))
	}

	selector.Inline(langRows...)

	_, err := b.Reply(message, tr.Tr("choose_lang"), selector)
	if err != nil {
		log.Println(err)
	}
}

func (b *Bot) Entry(m *tb.Message) {
	_, err := b.db.GetChat(m.Chat.ID)
	if err == mongo.ErrNoDocuments {
		b.ChooseLang(m)
	} else if err != nil {
		log.Println(err)
	}
}

var InvaliCommand = errors.New("invalid command")

var InvalidTimeFormat = errors.New("invalid time format")

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

	if StringHasSubmatch(command, "-s") {
		rem.Priority = internal.Silent
	}

	parts := strings.Split(command, " ")
	if len(parts) < 2 {
		return InvaliCommand
	}

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

	chat, err := b.GetChat(m.Chat.ID)
	if err != nil {
		log.Printf("%v %v", err, m.Chat)
	}
	rem := internal.Reminder{
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

	selector := createAlarmSelector(&rem, chat.Language)

	var message string
	loc, err := time.LoadLocation(chat.Loc)
	if rem.IsRepeated {
		messageFormat := tr.Lang(string(chat.Language)).Tr("alarm/add_repeat")
		message = fmt.Sprintf(messageFormat, rem.Description, rem.Every)
	} else {
		messageFormat := tr.Lang(string(chat.Language)).Tr("alarm/add_normal")

		if chat.IsJalali {
			jalaliTime, _ := jalaali.From(rem.AtTime.In(loc)).JFormat("Mon _2 Jan 06 | 15:04:05")
			message = fmt.Sprintf(messageFormat, rem.Description, jalaliTime)
		} else {
			message = fmt.Sprintf(messageFormat, rem.Description, rem.AtTime.In(loc).Format("Mon _2 Jan 06 | 15:04:05"))
		}

	}
	_, _ = b.Edit(reply, message, selector)

	b.AddReminder(&rem)
}

func createAlarmSelector(rem *internal.Reminder, lang internal.Language) *tb.ReplyMarkup {
	selector := &tb.ReplyMarkup{}

	btnDlt := selector.Data(tr.Lang(string(lang)).Tr("buttons/delete"), DeleteAlarmCall, fmt.Sprintf("%d", rem.Id))

	var btnSec tb.Btn
	if rem.Priority == 0 {
		btnSec = selector.Data(tr.Lang(string(lang)).Tr("buttons/unmute"), UnMuteAlarmCall, fmt.Sprintf("%d", rem.Id))
	} else {
		btnSec = selector.Data(tr.Lang(string(lang)).Tr("buttons/mute"), MuteAlarmCall, fmt.Sprintf("%d", rem.Id))
	}
	selector.Inline(
		selector.Row(btnDlt, btnSec),
	)

	return selector
}
