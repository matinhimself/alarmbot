package bot

import (
	"context"
	"fmt"
	"github.com/jalaali/go-jalaali"
	"github.com/psyg1k/remindertelbot/internal"
	log "github.com/sirupsen/logrus"
	"github.com/tucnak/tr"
	tb "gopkg.in/tucnak/telebot.v2"
	"strconv"
	"time"
)

const DateFormat = "Mon _2 Jan 06"

func (b *Bot) InitTaskList(m *tb.Message) {
	isSuperGp := m.Chat.Type == tb.ChatSuperGroup

	loge := log.WithField("chat", m.Chat.ID)
	loge.Info("initializing chat task list")

	chat, err := b.GetChat(m.Chat.ID)
	if err != nil {
		loge.Errorf("Couldn't find chat in database/cache. %v", err)
		b.HandleError(m, "Couldn't find chat")
	}

	msg, err := b.Send(m.Chat, tr.Lang(string(chat.Language)).Tr("responds/task_list_registered"))
	if err != nil {
		loge.Errorf("Couldn't send message")
		return
	}
	err = b.updateChatTaskList(&chat, msg.ID)
	if err != nil {
		loge.Errorf("Couldn't initialize chat task list. %v", err)
		b.HandleError(m, "Couldn't initialize chat task list")
		return
	}

	b.updateTaskList(&chat, msg, false, false, isSuperGp)

}

func (b *Bot) UpdateTaskListCall(c *tb.Callback) {
	isSuperGp := c.Message.Chat.Type == tb.ChatSuperGroup

	chat, _ := b.GetChat(c.Message.Chat.ID)

	ctx := context.Background()
	ctx = context.WithValue(ctx, "chat", chat)
	var isDur = false
	if c.Data == "dur" {
		isDur = true
	}
	b.updateTaskList(&chat, &tb.Message{ID: chat.TaskList, Chat: &tb.Chat{ID: c.Message.Chat.ID}}, false, isDur, isSuperGp)
	_ = b.Respond(c, &tb.CallbackResponse{Text: tr.Lang(string(chat.Language)).Tr("responds/update")})
}

func (b *Bot) ReformatTaskList(c *tb.Callback) {
	isSuperGp := c.Message.Chat.Type == tb.ChatSuperGroup

	chat, _ := b.GetChat(c.Message.Chat.ID)
	var isDur = true
	if c.Data == "dur" {
		isDur = false
	}
	b.updateTaskList(&chat, &tb.Message{ID: chat.TaskList, Chat: &tb.Chat{ID: c.Message.Chat.ID}}, true, isDur, isSuperGp)
	_ = b.Respond(c, &tb.CallbackResponse{Text: tr.Lang(string(chat.Language)).Tr("responds/reformat")})

}

func (b *Bot) ClearTaskList(c *tb.Callback) {
	isSuperGp := c.Message.Chat.Type == tb.ChatSuperGroup

	chat, _ := b.GetChat(c.Message.Chat.ID)
	var isDur = true
	if c.Data == "dur" {
		isDur = false
	}

	deleted, err := b.DeletePassedReminders(chat.ChatID)
	if err != nil {
		log.WithField("chat", chat.ChatID).Error("Couldn't delete reminders")
		_ = b.Respond(c, &tb.CallbackResponse{Text: "Error"})
		return
	}
	b.updateTaskList(&chat, &tb.Message{ID: chat.TaskList, Chat: &tb.Chat{ID: c.Message.Chat.ID}}, true, isDur, isSuperGp)
	_ = b.Respond(c, &tb.CallbackResponse{Text: fmt.Sprintf(tr.Lang(string(chat.Language)).Tr("responds/clear"), deleted)})

}

func (b *Bot) updateTaskList(c *internal.Chat, taskList *tb.Message, reformat, isDur, isSuperGp bool) {
	reminders, err := b.db.GetRemindersByChatID(c.ChatID)

	if err != nil {
		log.Println(err)
		return
	}

	// xor isdur based on reformat boolean
	isDur = isDur != reformat
	message := generateTaskListMessage(reminders, c, isDur, isSuperGp)
	selector := generateTasklistSelector(isDur, c)

	_, err = b.Edit(taskList, message, selector)
	if err != nil && err != tb.ErrSameMessageContent {
		log.WithField("chat", c.ChatID).Errorf("Couldn't update task list message. %v", err)
	}

}

func generateTasklistSelector(dur bool, c *internal.Chat) *tb.ReplyMarkup {

	selector := &tb.ReplyMarkup{}

	var reformData string
	var updateData string
	if dur {
		updateData = "dur"
	} else {
		reformData = "dur"
	}
	btnClear := selector.Data(tr.Lang(string(c.Language)).Tr("buttons/clear"), ClearCall, updateData)

	btnUpdate := selector.Data(tr.Lang(string(c.Language)).Tr("buttons/update"), UpdateCall, updateData)

	btnReform := selector.Data(tr.Lang(string(c.Language)).Tr("buttons/reformat"), ReformCall, reformData)

	selector.Inline(selector.Row(btnUpdate, btnReform), selector.Row(btnClear))

	return selector
}

func generateTaskListMessage(reminders []internal.Reminder, c *internal.Chat, isDur, isSuperGp bool) string {

	var message string
	var t string
	if c.IsJalali {
		t, _ = jalaali.Now().JFormat(DateFormat)
	} else {
		t = time.Now().Format(DateFormat)
	}
	if len(reminders) == 0 {
		message = fmt.Sprintf("%s\n%s", t, tr.Lang(string(c.Language)).Tr("reminders/no_rem"))
	} else {
		message = fmt.Sprintf("%s\n", t)
	}

	loc, _ := time.LoadLocation(c.Loc)

	for i, reminder := range reminders {
		if reminder.Description != "" {
			reminder.Description = "*" + reminder.Description + "* "
		}
		var msg, link string

		if isDur {
			msg = reminderToStringDur(&reminder, loc, c.Language)
		} else {
			msg = reminderToStringDate(&reminder, loc, c.Language, c.IsJalali)
		}
		if isSuperGp {
			st := strconv.FormatInt(c.ChatID, 10)
			link = fmt.Sprintf("t.me/c/%s/%d", st[4:], reminder.Message)
			msg = fmt.Sprintf("%d.%s[%c](%s)\n", i+1, msg, '🔗', link)
		}
		message = fmt.Sprintf("%s%s", message, msg)

	}
	return message
}

func reminderToStringDur(reminder *internal.Reminder, loc *time.Location, lang internal.Language) string {
	var message string
	if reminder.IsRepeated {
		diff := reminder.AtTime.Unix() - time.Now().Unix()
		mod := diff % int64(reminder.Every.Seconds())
		message = fmt.Sprintf("%s %s %s",
			reminder.Description, "➿", time.Duration(mod)*time.Second)
		return message
	}

	timeDuration := reminder.AtTime.Sub(time.Now().In(loc))
	durationString, now := DurationToString(timeDuration)

	var emoji rune
	if reminder.Description != "" {
		emoji = '|'
	} else {
		emoji = ' '
	}

	if now {
		durationString = tr.Lang(string(lang)).Tr("reminders/done")
		emoji = '☑'
	}

	return fmt.Sprintf("%s %c%s \n",
		reminder.Description, emoji, durationString)
}

func reminderToStringDate(reminder *internal.Reminder, loc *time.Location, lang internal.Language, isJalali bool) string {

	var timeString string
	var emoji string

	if reminder.IsRepeated {
		t := time.Now().In(loc)
		diff := reminder.AtTime.Unix() - t.Unix()
		mod := diff % int64(reminder.Every.Seconds())

		if isJalali {
			timeString, _ = jalaali.From(t.Add(time.Duration(mod) * time.Second).In(loc)).JFormat("Mon _2 Jan 06 | 15:04")
		} else {
			timeString = t.Add(time.Duration(mod) * time.Second).In(loc).Format("Mon _2 Jan 06 | 15:04")
		}
		emoji = "➿"
	} else {
		if isJalali {
			timeString, _ = jalaali.From(reminder.AtTime.In(loc)).JFormat("Mon _2 Jan 06 | 15:04")
		} else {
			timeString = reminder.AtTime.In(loc).Format("Mon _2 Jan 06 | 15:04")
		}
		if reminder.Description != "" {
			emoji = "|"
		} else {
			emoji = ""
		}

	}
	return fmt.Sprintf("%s %s %s", reminder.Description, emoji, timeString)
}
