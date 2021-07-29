package bot

import (
	"fmt"
	"github.com/psyg1k/remindertelbot/internal"
	m "github.com/psyg1k/remindertelbot/internal/mongo"
	sc "github.com/psyg1k/remindertelbot/pkg/shceduler"
	"github.com/tucnak/tr"
	tb "gopkg.in/tucnak/telebot.v2"
	"log"
	"os"
	"strings"
	"time"
)

type Bot struct {
	*tb.Bot
	db  *m.Db
	log *log.Logger

	s         *sc.Scheduler
	alarmChan chan interface{}
	Cache     Cache
}

const (
	MongoKey           = "MONGO_URI"
	TokenKey           = "BOT_TOKEN"
	SetTimeZoneCommand = "/settz"
)

func setDataBase() (*m.Db, error) {
	uri, ok := os.LookupEnv(MongoKey)
	if !ok {
		return nil, fmt.Errorf("couldn't find mongo uri in env variables")
	}

	log.Println("Connecting to database")
	db, err := m.NewDb(uri)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func setBot() (*tb.Bot, error) {

	token, ok := os.LookupEnv(TokenKey)
	if !ok {
		log.Fatal("couldn't find bot token in env variables")
	}

	b, err := tb.NewBot(tb.Settings{
		Reporter: func(err error) {
			log.Println(err.Error())
		},

		Token:     token,
		Poller:    &tb.LongPoller{Timeout: 10 * time.Second},
		ParseMode: tb.ModeMarkdown,
	})

	return b, err
}

func NewBot() *Bot {

	if err := tr.Init("i18n", "en"); err != nil {
		log.Println(err)
	}

	db, err := setDataBase()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to database successfully")

	bot, err := setBot()
	if err != nil {
		log.Fatal(err)
	}

	alarmChannel := make(chan interface{})

	scheduler := sc.NewScheduler(time.NewTicker(time.Second), alarmChannel)

	return &Bot{
		Bot:       bot,
		alarmChan: alarmChannel,
		s:         scheduler,
		db:        db,
		Cache:     NewCache(),
	}
}

func (b *Bot) Run() {

	// starting scheduler
	go b.s.Start()
	log.Println("Scheduler started")

	// starting telegram bot
	go b.Start()
	log.Println("Telegram bot started")

	log.Println("Loading reminders from database")
	err := b.LoadReminders()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Reminders loaded successfully")

	log.Println("Start listening to alarm channel")
	for {
		select {
		case alarm := <-b.alarmChan:
			r := alarm.(*internal.Reminder)
			if r.Priority != 0 || r.AtTime.Unix() < time.Now().Unix() {
				b.sendAlarm(r)
			}
		}
	}

}

func (b *Bot) AddReminder(r *internal.Reminder) {
	fmt.Printf("%v", r)
	b.s.AddJob(r, r.AtTime, r.Every, r.From)
}

func (b *Bot) Qtz(q *tb.Query) {
	results := make(tb.Results, 0)
	var c int = 0
	for _, url := range tzs {
		if s := strings.Split(url, "/"); strings.HasPrefix(strings.ToLower(s[len(s)-1]), q.Text) {
			c++
			result := &tb.ArticleResult{
				Title:   s[len(s)-1],
				Text:    fmt.Sprintf("%s %s", SetTimeZoneCommand, url),
				HideURL: true,
			}
			result.SetResultID(url)
			results = append(results, result)
		}
		if c > 2 {
			break
		}
	}

	err := b.Answer(q, &tb.QueryResponse{
		Results:   results,
		CacheTime: 60,
	})

	if err != nil {
		log.Println(err)
	}
}

func (b *Bot) sendAlarm(r *internal.Reminder) {
	fmt.Printf("%v", r.AtTime.Add(210*time.Minute))
	_, err := b.Send(&tb.Chat{ID: r.ChatId}, "message")
	if err != nil {
		fmt.Println(err)
	}
}
