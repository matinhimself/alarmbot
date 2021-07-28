package main

import (
	"github.com/psyg1k/remindertelbot/internal/bot"
	"github.com/tucnak/tr"
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
	b.Handle("\f"+bot.LangCall, b.SetLanguage)
	b.Handle("\f"+bot.TzCall, b.SetTz)
	b.Run()
}
