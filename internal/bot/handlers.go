package bot

import (
	"errors"
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
	LangCall = "lang"
	TzCall   = "tz"

	DefaultEvery        = time.Hour * 12
	DefaultFromDuration = time.Hour * 72
)

func (b *Bot) SetTzCommand(m *tb.Message) {
	chat, _ := b.GetChat(m.Chat.ID)

	parts := strings.SplitN(m.Text, " ", 2)
	if len(parts) < 2 {
		b.HandleError(m, tr.Lang(string(chat.Language)).Tr("wrong_tz_format"))
	}
	duration, err := time.ParseDuration(parts[1])
	if err != nil {
		b.HandleError(m, tr.Lang(string(chat.Language)).Tr("wrong_tz_format"))
	}

	err = b.UpdateTz(chat.ChatID, internal.Offset(duration))
	if err != nil {
		_, _ = b.Reply(m, err.Error())
	}
}

func (b *Bot) HandleError(m *tb.Message, s string) {
	_, _ = b.Reply(m, s)
}

func (b *Bot) SetTz(c *tb.Callback) {
	chat, _ := b.GetChat(c.Message.Chat.ID)
	offset, _ := time.ParseDuration(c.Data)
	err := b.UpdateTz(c.Message.Chat.ID, internal.Offset(offset))
	if err != nil {
		log.Println(err)
	}
	_, _ = b.Edit(c.Message, tr.Lang(string(chat.Language)).Tr("registered"))
}

func (b *Bot) SetLanguage(call *tb.Callback) {
	lang := internal.Language(call.Data)

	chat := internal.Chat{
		ChatID:    call.Message.Chat.ID,
		TaskList:  0,
		UTCOffset: 0,
		Language:  lang,
		IsJalali:  true,
	}

	err := b.db.InsertChat(chat)
	if err != nil {
		//TODO
	}

	selector := &tb.ReplyMarkup{}
	selector.Inline(
		selector.Row(selector.Data("Iran", TzCall, internal.IRAN.String())),
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

func ParseParam(parts []string, rem *internal.Reminder, t time.Time, isJalali bool) error {
	if strings.ToLower(parts[0]) == "repeat" {
		rem = internal.WithRepeat(rem)
		return nil
	}

	if duration, err := time.ParseDuration(parts[0]); err != nil {
		rem.AtTime = t.Add(duration)
		return nil
	}

	var date time.Time
	var err error

	if isJalali {
		date, err = stringToJalaliDate(parts[0])
	} else {
		date, err = stringToDate(parts[0])
	}

	if err != nil {
		return InvalidTimeFormat
	}

	rem.AtTime = date

	return nil
}
func ParseAddCommand(command string, rem *internal.Reminder, offset internal.Offset, isJalali bool) error {
	t := time.Now().UTC()

	if StringHasSubmatch(command, "-s") {
		rem.Priority = internal.Silent
	}

	parts := strings.Split(command, " ")
	if len(parts) < 2 {
		return InvaliCommand
	}

	err := ParseParam(parts[1:], rem, t, isJalali)
	if err != nil {
		return err
	}

	param, err := GetParam(command, "-t", "time")
	if err == nil {
		err := formatDateTime(param, &rem.AtTime, offset)
		if err != nil {
			return err
		}
	}

	if rem.AtTime.Unix() > t.Unix() {
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

func (b *Bot) addCommand(m *tb.Message) {
	chat, err := b.GetChat(m.Chat.ID)
	if err != nil {
		log.Printf("%v %v", err, m.Chat)
	}
	rem := internal.Reminder{
		Every:   DefaultEvery,
		From:    time.Now().UTC().Add(-1 * DefaultFromDuration),
		Created: time.Now().UTC(),
	}

	err = ParseAddCommand(m.Text, &rem, chat.UTCOffset, chat.IsJalali)
	if err == ErrInvalidEveryFormat {
		b.HandleError(m, tr.Lang(string(chat.Language)).Tr("errors/every_format"))
	} else if err == ErrInvalidFromFormat {
		b.HandleError(m, tr.Lang(string(chat.Language)).Tr("errors/from_format"))
	} else if err == ErrPassedTime {
		b.HandleError(m, tr.Lang(string(chat.Language)).Tr("errors/passed_time"))
	}

	reminder, err := b.db.InsertReminder(rem)
	if err != nil {
		log.Println(err)
	}

	// TODO: init alarm

}

var ErrInvalidFromFormat error = errors.New("invalid from format")
var ErrPassedTime error = errors.New("passed time")
var ErrInvalidEveryFormat error = errors.New("invalid every format")
