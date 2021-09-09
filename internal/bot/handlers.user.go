package bot

import (
	"fmt"
	"github.com/psyg1k/remindertelbot/internal"
	log "github.com/sirupsen/logrus"
	"github.com/tucnak/tr"
	tb "gopkg.in/tucnak/telebot.v2"
	"strings"
	"time"
)

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
	chat.Loc = parts[1]
	_, _ = b.Reply(m, generateSettingsMessage(chat))
}

func generateSettingsMessage(chat internal.Chat) string {
	format := tr.Lang(string(chat.Language)).Tr("responses/updated")
	var cal string
	if chat.IsJalali {
		cal = tr.Lang(string(chat.Language)).Tr("cal/" + internal.GeoCal)
	} else {
		cal = tr.Lang(string(chat.Language)).Tr("cal/" + internal.HijriCal)
	}
	return fmt.Sprintf(format, chat.ChatID, chat.Loc, cal, chat.Language)
}

func (b *Bot) SetTz(c *tb.Callback) {
	chat, err := b.GetChat(c.Message.Chat.ID)
	ent := log.WithField("chat", chat)
	if err != nil {
		ent.Infof("chat", c.Data)
	}
	_, err = time.LoadLocation(c.Data)
	if err != nil {
		ent.Infof("couldn't load location %s", c.Data)
		b.HandleErrorErr(c.Message, err, chat.Language)
		return
	}
	err = b.UpdateTz(c.Message.Chat.ID, c.Data)
	if err != nil {
		log.Println(err)
		b.HandleError(c.Message, err.Error())
	}
	chat.Loc = c.Data
	_, _ = b.Edit(c.Message, generateSettingsMessage(chat))
}

func (b *Bot) SetLanguage(c *tb.Callback) {
	lang := internal.Language(c.Data)

	chat := internal.Chat{
		ChatID:   c.Message.Chat.ID,
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
		selector.Row(selector.Data(tr.Lang(string(chat.Language)).Tr("cal/gregorian"), CalCall, internal.GeoCal)),
		selector.Row(selector.Data(tr.Lang(string(chat.Language)).Tr("cal/hijri"), CalCall, internal.HijriCal)),
	)

	_, _ = b.Edit(c.Message, tr.Lang(string(chat.Language)).Tr("commands/choose_cal"), selector)

}

func (b *Bot) ChooseLang(message *tb.Message) {
	log.Infof("Choosing lang by user %d ", message.Chat.ID)
	selector := &tb.ReplyMarkup{}

	var langRows []tb.Row

	for _, language := range internal.Languages {
		langRows = append(langRows, selector.Row(selector.Data(language.ToString(), LangCall, string(language))))
	}

	selector.Inline(langRows...)

	_, err := b.Reply(message, tr.Tr("commands/choose_lang"), selector)
	if err != nil {
		log.Println(err)
	}
}

func (b *Bot) ChooseCal(c *tb.Callback) {
	chat, _ := b.GetChat(c.Message.Chat.ID)
	var isHijri bool
	if c.Data == internal.HijriCal {
		isHijri = true
	} else {
		isHijri = false
	}

	_ = b.UpdateCal(c.Message.Chat.ID, isHijri)

	selector := &tb.ReplyMarkup{}

	rows := make([]tb.Row, len(internal.TimeZones)+1)
	for k, v := range internal.TimeZones {
		rows = append(rows, selector.Row(selector.Data(k, TzCall, fmt.Sprintf("\f%s|%s", TzCall, v))))
	}
	rows = append(rows, selector.Row(selector.QueryChat(tr.Lang(string(chat.Language)).Tr("buttons/search"), "")))

	selector.Inline(rows...)
	_, err := b.Edit(c.Message, tr.Lang(string(chat.Language)).Tr("commands/choose_region"), selector)
	if err != nil {
		log.Println(err)
	}

}
