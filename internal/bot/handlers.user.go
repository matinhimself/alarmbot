package bot

import (
	"github.com/psyg1k/remindertelbot/internal"
	"github.com/tucnak/tr"
	tb "gopkg.in/tucnak/telebot.v2"
	"log"
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
	_, _ = b.Reply(m, tr.Lang(string(chat.Language)).Tr("tz_changed"))
}

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

	selector := &tb.ReplyMarkup{}
	selector.Inline(
		selector.Row(selector.Data(tr.Lang(string(chat.Language)).Tr("cal/gregorian"), CalCall, internal.GeoCal)),
		selector.Row(selector.Data(tr.Lang(string(chat.Language)).Tr("cal/hijri"), CalCall, internal.HijriCal)),
	)

	_, _ = b.Edit(c.Message, tr.Lang(string(chat.Language)).Tr("choose_cal"), selector)
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

func (b *Bot) ChooseCal(call *tb.Callback) {
	chat, _ := b.GetChat(call.Message.Chat.ID)
	var isHijri bool
	if call.Data == internal.HijriCal {
		isHijri = true
	} else {
		isHijri = false
	}

	_ = b.UpdateCal(call.Message.Chat.ID, isHijri)

	b.Edit(call.Message, tr.Lang(string(chat.Language)).Tr("registered"))
}
