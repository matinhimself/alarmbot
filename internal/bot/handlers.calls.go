package bot

import (
	"fmt"
	"github.com/psyg1k/remindertelbot/internal"
	"github.com/tucnak/tr"
	tb "gopkg.in/tucnak/telebot.v2"
	"log"
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

func updateInlineMenuMute(menu [][]tb.InlineButton, p internal.Priority, selector *tb.ReplyMarkup, language internal.Language) *[][]tb.Btn {
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
				buttons[i][j] = selector.Data(button.Text, u, d)
			}
			p, id, err := extractId(d)
			if err != nil {
				continue
			}
			if p == internal.Audible {
				buttons[i][j] = selector.Data(tr.Lang(string(language)).Tr("buttons/"+p.ToString()), MuteCall, fmt.Sprintf("%s:%s", p.ToString(), id))
			} else {
				buttons[i][j] = selector.Data(tr.Lang(string(language)).Tr("buttons/"+p.ToString()), MuteCall, fmt.Sprintf("%s:%s", p.ToString(), id))
			}
		}
	}
	return &buttons
}

func (b *Bot) ToggleMute(c *tb.Callback) {
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
	buttons := updateInlineMenuMute(c.Message.ReplyMarkup.InlineKeyboard, p, selector, chat.Language)
	rows := make([]tb.Row, 0)
	for _, button := range *buttons {
		rows = append(rows, button)
	}

	selector.Inline(rows...)

	err := b.db.UpdatePriority(id, p)
	if err != nil {
		log.Println("ToggleMute", err)
		return
	}

	_, _ = b.Edit(c.Message, c.Message.Text, selector)

	_ = b.Respond(c, &tb.CallbackResponse{
		CallbackID: c.ID,
		Text:       fmt.Sprintf("%s", tr.Lang(string(chat.Language)).Tr(fmt.Sprintf("buttons/%sd", (p.Not()).ToString()))),
	})
}
