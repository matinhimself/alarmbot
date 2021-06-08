package internal

import "time"

type Chat struct {
	ChatID    int64         `bson:"_id"`
	TaskList  int           `bson:"task_list"`
	UTCOffset time.Duration `bson:"utc_offset"`
	lang      Language      `bson:"lang"`
}
