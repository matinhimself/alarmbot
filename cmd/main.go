package main

import (
	_ "github.com/joho/godotenv/autoload"
	"github.com/psyg1k/remindertelbot/internal/bot"
	"github.com/tucnak/tr"
	tb "gopkg.in/tucnak/telebot.v2"
	"log"
)

func initTr() {
	if err := tr.Init("i18n", "en"); err != nil {
		log.Println(err)
	}
}

type fromCall func() string

func (f fromCall) CallbackUnique() string {
	return f()
}

func lambdaString(s string) func() string {
	return func() string {
		return s
	}
}

func main() {
	b := bot.NewBot()
	initTr()
	b.Handle("/start", b.Entry)
	b.Handle(bot.HelpCommand, b.HelpCommand)
	b.Handle(bot.SetTimeZoneCommand, b.SetTzCommand)
	b.Handle(bot.AddReminderCommand, b.AddCommand)
	b.Handle(bot.PingCommand, b.Ping)

	b.Handle("\f"+bot.LangCall, b.SetLanguage)
	b.Handle("\f"+bot.DeleteAlarmCall, b.DeleteReminder)
	b.Handle("\f"+bot.MuteCall, b.ToggleMute)
	b.Handle("\f"+bot.TzCall, b.SetTz)
	b.Handle("\f"+bot.CalCall, b.ChooseCal)
	b.Handle("\f"+bot.UpdateCall, b.UpdateTaskListCall)
	b.Handle("\f"+bot.ReformCall, b.ReformatTaskList)
	b.Handle("\f"+bot.ClearCall, b.ClearTaskList)

	b.Handle(bot.InitTaskList, b.InitTaskList)
	b.Handle(tb.OnQuery, b.Qtz)
	b.Run()
}
