package bot

import (
	"errors"
	"fmt"
	"github.com/psyg1k/remindertelbot/internal"
	"github.com/tucnak/tr"
	"go.mongodb.org/mongo-driver/mongo"
	tb "gopkg.in/tucnak/telebot.v2"
	"log"
	"time"
)

const (
	LangCall        = "lang"
	TzCall          = "tz"
	DeleteAlarmCall = "del"
	MuteCall        = "p"
	CalCall         = "call"

	MuteUniqueData   = "mute"
	UnmuteUniqueData = "unmute"

	SetTimeZoneCommand = "/settz"
	AddReminderCommand = "/add"

	DefaultEvery        = time.Hour * 12
	DefaultFromDuration = time.Hour * 72
)

var ErrInvalidFromFormat = errors.New("invalid from format")
var ErrPassedTime = errors.New("passed time")
var ErrInvalidEveryFormat = errors.New("invalid every format")
var ErrInvalidTz = errors.New("invalid_timezone")
var InvalidCommand = errors.New("invalid command")
var InvalidTimeFormat = errors.New("invalid time format")

func (b *Bot) HandleError(m *tb.Message, s string) {
	_, _ = b.Reply(m, s)
}

func (b *Bot) HandleErrorErr(m *tb.Message, e error, lang internal.Language) {
	if e == ErrInvalidTz {
		_, _ = b.Reply(m, tr.Lang(string(lang)).Tr(fmt.Sprintf("errors/%s", e.Error())))
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
