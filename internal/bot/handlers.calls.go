package bot

import (
	"fmt"
	"github.com/psyg1k/remindertelbot/internal"
	log "github.com/sirupsen/logrus"
	"github.com/tucnak/tr"
	tb "gopkg.in/tucnak/telebot.v2"
	"strings"
)

func extractUnique(s string) (string, string, error) {
	datas := strings.Split(strings.TrimSpace(s), "|")
	if len(datas) < 2 {
		return "", "", fmt.Errorf("wrong format")
	}
	return datas[0], datas[1], nil
}

func extractId(s string) (internal.Priority, string, error) {
	d := strings.Split(s, ":")
	if len(d) < 2 {
		return 0, "", fmt.Errorf("wrong format")
	}

	var p internal.Priority
	if d[0] == MuteUniqueData {
		p = internal.Silent
	} else {
		p = internal.Audible
	}
	return p, d[1], nil
}

func updateInlineMenuMute(menu [][]tb.InlineButton, selector *tb.ReplyMarkup, language internal.Language) *[][]tb.Btn {
	buttons := make([][]tb.Btn, len(menu))
	for i := range buttons {
		buttons[i] = make([]tb.Btn, len(menu[i]))
	}

	for i, rep := range menu {
		for j, button := range rep {
			u, d, err := extractUnique(button.Data)
			if err != nil {
				log.Println("updateInlineMenu: ", err)
			}
			if u != MuteCall {
				buttons[i][j] = selector.Data(button.Text, u, button.Data)
			}
			p, id, err := extractId(d)
			if err != nil {
				continue
			}
			if p == internal.Audible {
				buttons[i][j] = selector.Data(tr.Lang(string(language)).Tr("buttons/"+p.ToString()), MuteCall, fmt.Sprintf("\f%s|%s:%s", MuteCall, p.ToString(), id))
			} else {
				buttons[i][j] = selector.Data(tr.Lang(string(language)).Tr("buttons/"+p.ToString()), MuteCall, fmt.Sprintf("\f%s|%s:%s", MuteCall, p.ToString(), id))
			}
		}
	}
	return &buttons
}

func (b *Bot) ToggleMute(c *tb.Callback) {
	fmt.Println(c.Data)
	chat, _ := b.GetChat(c.Message.Chat.ID)
	var p internal.Priority
	data := strings.Split(c.Data, ":")
	if len(data) < 2 {
		log.Println("callback data is not well formatted")
	}

	id := data[1]

	if data[0] == MuteUniqueData {
		p = internal.Silent

	} else {
		p = internal.Audible
	}
	selector := &tb.ReplyMarkup{}
	buttons := updateInlineMenuMute(c.Message.ReplyMarkup.InlineKeyboard, selector, chat.Language)
	rows := make([]tb.Row, len(*buttons))
	for _, button := range *buttons {
		rows = append(rows, button)
	}

	selector.Inline(rows...)

	err := b.updateReminderPriority(id, p)
	if err != nil {
		log.Println(err)
		return
	}

	_, _ = b.Edit(c.Message, c.Message.Text, selector)

	_ = b.Respond(c, &tb.CallbackResponse{
		CallbackID: c.ID,
		Text:       tr.Lang(string(chat.Language)).Tr(fmt.Sprintf("buttons/%sd", (p.Not()).ToString())),
	})
}

func (b *Bot) DeleteReminder(c *tb.Callback) {
	log.WithField("data", c.Data).Info("Delete reminder")
	chat, err := b.GetChat(c.Message.Chat.ID)
	if err != nil {
		log.WithField("chat", c.Message.Chat.ID).Error("couldn't find chat ")
	}
	rem := c.Data

	err = b.db.DeleteReminder(rem)
	if err != nil {
		log.Error("Couldn't delete reminder from database", rem, err)
	}
	err = b.s.DeleteByIdentifier(rem)
	if err != nil {
		log.Error("Couldn't delete reminder from job handler", rem, err)
	}

	_, _ = b.Edit(c.Message, fmt.Sprintf("%s\n\n%s", c.Message.Text, tr.Lang(string(chat.Language)).Tr("alarm/deleted")))
}
