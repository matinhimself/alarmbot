package main

import (
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

func main() {
	b := bot.NewBot()
	initTr()
	b.Handle("/start", b.Entry)
	b.Handle("/settz", b.SetTzCommand)
	b.Handle("/add", b.AddCommand)
	b.Handle("\f"+bot.LangCall, b.SetLanguage)
	b.Handle("\f"+bot.DeleteAlarmCall, b.DeleteReminder)
	b.Handle("\f"+bot.MuteCall, b.ToggleMute)
	b.Handle("\f"+bot.TzCall, b.SetTz)
	b.Handle(tb.OnQuery, b.Qtz)
	b.Run()
}
