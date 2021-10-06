package bot

import (
	"fmt"
	"github.com/psyg1k/remindertelbot/internal"
	m "github.com/psyg1k/remindertelbot/internal/mongo"
	sc "github.com/psyg1k/remindertelbot/pkg/shceduler"
	log "github.com/sirupsen/logrus"
	"github.com/tucnak/tr"
	tb "gopkg.in/tucnak/telebot.v2"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type Bot struct {
	*tb.Bot
	db        *m.Db
	s         *sc.Scheduler
	alarmChan chan interface{}
	Cache     Cache
}

const (
	MongoKey = "MONGO_URI"
	TokenKey = "BOT_TOKEN"
	ProxyKey = "BOT_PROXY"
)

func (b *Bot) DeletePassedReminders(chatId int64) (deleted int64, err error) {
	deleted, err = b.db.DeleteRemindersBefore(chatId, time.Now().UTC())
	return
}
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
	proxy, ok := os.LookupEnv(ProxyKey)
	var client *http.Client
	if ok {
		log.Printf("using %s as proxy.", proxy)
		proxyUrl, err := url.Parse(proxy)
		if err != nil {
			log.Fatal(err)
		}
		client = &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}
	}

	b, err := tb.NewBot(tb.Settings{

		//Reporter: func(err error) {
		//	log.Println(err.Error())
		//},
		//Verbose: true,
		Client:    client,
		Token:     token,
		Poller:    &tb.LongPoller{Timeout: 10 * time.Second},
		ParseMode: tb.ModeMarkdown,
	})
	if err != nil {
		log.Fatal(err)
	}

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

	for alarm := range b.alarmChan {
		r := alarm.(*internal.Reminder)
		if r.Priority != 0 || r.AtTime.Unix() < time.Now().Unix() {
			b.sendAlarm(r)
		}
	}

}

func (b *Bot) AddReminder(r *internal.Reminder) {
	b.s.AddJob(r, r.AtTime, r.Every, r.From, r.IsRepeated)
}

func (b *Bot) Qtz(q *tb.Query) {
	results := make(tb.Results, 0)
	var c = 0
	for _, u := range tzs {
		if s := strings.Split(u, "/"); strings.HasPrefix(strings.ToLower(s[len(s)-1]), q.Text) {
			c++
			result := &tb.ArticleResult{
				Title:   s[len(s)-1],
				Text:    fmt.Sprintf("%s %s", SetTimeZoneCommand, u),
				HideURL: true,
			}
			result.SetResultID(u)
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

func (b *Bot) sendAlarm(rem *internal.Reminder) {
	chat, _ := b.GetChat(rem.ChatId)

	var selector *tb.ReplyMarkup
	if rem.AtTime.Before(time.Now().UTC()) {
		selector = createAlarmSelector(rem, chat.Language, true, false)
	} else {
		selector = createAlarmSelector(rem, chat.Language, false, true)
	}

	var format string
	if rem.IsRepeated {
		format = tr.Lang(string(chat.Language)).Tr("alarm/repeat")
	} else {
		format = tr.Lang(string(chat.Language)).Tr("alarm/normal")
	}
	message := generateAlarmMessage(format, rem, chat)

	if rem.Message != 0 {
		_, err := b.Reply(&tb.Message{ID: rem.Message, Chat: &tb.Chat{ID: rem.ChatId}}, message, selector)
		if err != nil {
			message = fmt.Sprintf("%s\n%s", "__Description message deleted__", message)
		} else {
			return
		}
	}
	_, err := b.Send(tb.ChatID(rem.ChatId), message, selector)
	if err != nil {
		log.Println(err)
	}

}

func (b *Bot) updateReminderPriority(id string, p internal.Priority) error {
	err := b.db.UpdatePriority(id, p)
	if err != nil {
		return err
	}
	data, found := b.s.GetJobData(id)
	if found {
		data.(*internal.Reminder).Priority = p
	}
	return nil
}

func (b *Bot) updateChatTaskList(chat *internal.Chat, id2 int) error {
	chat.TaskList = id2
	_ = b.Cache.UpdateChatTaskList(chat.ChatID, id2)
	err := b.db.UpdateChatTaskList(chat.ChatID, id2)
	return err
}
