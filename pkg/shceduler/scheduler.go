package schduler

import (
	"fmt"
	"time"
)

type G interface {
	GetIdentifier() string
}

type Job struct {
	atTime  time.Time
	every   time.Duration
	from    time.Time
	lastRun time.Time
	nextRun time.Time

	data G

	isOnRepeat bool
	passed     bool
	isLastRun  bool
}

func NewJob(atTime time.Time, every time.Duration, from time.Time, data G, isOnRepeat bool) *Job {
	return &Job{
		atTime:     atTime,
		every:      every,
		from:       from,
		lastRun:    time.Time{},
		nextRun:    time.Time{},
		data:       data,
		isOnRepeat: isOnRepeat,
		passed:     false,
		isLastRun:  false,
	}
}

func (job *Job) IsTime() bool {
	return !job.passed && time.Now().UTC().After(job.nextRun)
}

func (s *Scheduler) Delete(index int) error {
	if index < 0 || index > len(s.jobs)-1 {
		return fmt.Errorf("index out of bound")
	}
	s.jobs[index] = s.jobs[len(s.jobs)-1]
	s.jobs = s.jobs[:len(s.jobs)-1]

	return nil
}

func (s *Scheduler) DeleteByIdentifier(id string) error {
	for i := 0; i < len(s.jobs); i++ {
		if s.jobs[i].data.GetIdentifier() == id {
			err := s.Delete(i)
			return err
		}
	}
	return nil
}

type Scheduler struct {
	lock          bool
	jobs          []*Job
	ticker        *time.Ticker
	passedChannel chan interface{}
}

func (s *Scheduler) GetJobData(identifier string) (G, bool) {
	for _, job := range s.jobs {
		if (job.data).GetIdentifier() == identifier {
			return job.data, true
		}
	}
	return nil, false
}

func NewScheduler(ticker *time.Ticker, passedChannel chan interface{}) *Scheduler {
	return &Scheduler{
		lock:          false,
		jobs:          nil,
		ticker:        ticker,
		passedChannel: passedChannel,
	}
}

func (s *Scheduler) AddJob(data G, atTime time.Time, every time.Duration,
	from time.Time, isOnRepeat bool) {

	job := NewJob(atTime, every, from, data, isOnRepeat)
	job.PlanFirstRun()
	s.jobs = append(s.jobs, job)
}

func (job *Job) PlanNext() {
	// plan next onRepeat run
	if job.isOnRepeat {
		job.nextRun = job.lastRun.Add(job.every)
		return
	}

	if job.isLastRun {
		// last tick
		job.passed = true
	} else if sum := job.lastRun.Add(job.every); sum.Unix() < job.atTime.Unix() {
		// plan next run for normal ticks
		job.nextRun = sum
	} else {
		// plan for last tick
		job.isLastRun = true
		job.nextRun = job.atTime
	}
}

func (job *Job) PlanFirstRun() {
	t := time.Now().UTC().Round(time.Second)

	if job.isOnRepeat {
		job.nextRun = job.atTime

		for job.nextRun.Before(t) {
			job.nextRun = job.nextRun.Add(job.every)
		}

		return
	}

	if job.from.Unix() < t.Unix() {
		if job.atTime.Sub(t) <= job.every {
			job.isLastRun = true
		}
		diff := job.atTime.Unix() - t.Unix()
		mod := diff % int64(job.every.Seconds())
		job.nextRun = t.Add(time.Duration(mod) * time.Second)
	} else {
		if job.from.Add(job.every).Unix() >= job.atTime.Unix() {
			job.isLastRun = true
		}
		diff := job.atTime.Unix() - job.from.Unix()
		mod := diff % int64(job.every.Seconds())
		job.nextRun = job.from.Add(time.Duration(mod) * time.Second).Round(time.Second)

	}
}

func (s *Scheduler) Start() {
	for range s.ticker.C {
		for i := 0; i < len(s.jobs); i++ {
			if s.lock {
				break
			}
			s.lock = true
			if s.jobs[i].passed {
				_ = s.Delete(i)
			} else if s.jobs[i].IsTime() {
				s.passedChannel <- s.jobs[i].data
				s.jobs[i].lastRun = s.jobs[i].nextRun
				s.jobs[i].PlanNext()
			}
			s.lock = false
		}
	}
}
