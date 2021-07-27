package bot

import (
	"github.com/psyg1k/remindertelbot/internal"
	"github.com/tucnak/tr"
	s2d "github.com/xhit/go-str2duration/v2"
	"go.mongodb.org/mongo-driver/mongo"
	tb "gopkg.in/tucnak/telebot.v2"
	"log"
	"strings"
)

const (
	LangCall = "lang"
	TzCall   = "tz"
)

func (b *Bot) SetTzCommand(m *tb.Message) {
	user, _ := b.GetUser(m.Sender.ID)

	parts := strings.SplitN(m.Text, " ", 2)
	if len(parts) < 2 {
		b.HandleError(m, tr.Lang(string(user.Language)).Tr("wrong_tz_format"))
	}
	duration, err := s2d.ParseDuration(parts[1])
	if err != nil {
		b.HandleError(m, tr.Lang(string(user.Language)).Tr("wrong_tz_format"))
	}

	err = b.UpdateTz(user.UserId, internal.Offset(duration))
	if err != nil {
		_, _ = b.Reply(m, err.Error())
	}
}

func (b *Bot) HandleError(m *tb.Message, s string) {
	_, _ = b.Reply(m, s)
}

func (b *Bot) SetTz(c *tb.Callback) {
	user, _ := b.GetUser(c.Sender.ID)
	offset, _ := s2d.ParseDuration(c.Data)
	err := b.UpdateTz(c.Sender.ID, internal.Offset(offset))
	if err != nil {
		log.Println(err)
	}
	_, _ = b.Edit(c.Message, tr.Lang(string(user.Language)).Tr("registered"))
}

func (b *Bot) SetLanguage(call *tb.Callback) {
	lang := internal.Language(call.Data)

	user := internal.User{
		UserId:   call.Sender.ID,
		Username: call.Sender.Username,
		Name:     strings.Join([]string{call.Sender.FirstName, call.Sender.LastName}, " "),
		Language: lang,
		Offset:   internal.Offset(0),
	}

	err := b.db.InsertUser(user)
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
	_, err := b.db.GetUser(m.Sender.ID)
	if err == mongo.ErrNoDocuments {
		b.ChooseLang(m)
	} else if err != nil {
		log.Println(err)
	}
}
