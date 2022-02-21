package schduler

import (
	"strconv"
	"testing"
	"time"
)

type data int

func (d data) GetIdentifier() string {
	return strconv.Itoa(int(d))
}

func TestRepeatReminder(t *testing.T) {
	c := make(chan interface{}, 0)
	s := NewScheduler(time.NewTicker(1*time.Second), c)
	s.AddJob(data(1), time.Now().Add(1 * time.Hour), time.Minute * 20, time.Now(), true)

}
