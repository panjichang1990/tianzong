package mtimer

import (
	"sync"
	"time"

	"github.com/roylee0704/gron"
)

type Work struct {
	f    func()
	diff time.Duration
}

type Job struct {
	j gron.Job
	s gron.Schedule
}

var TimerFunBox = make([]*Work, 0)
var TimeJobBox = make([]*Job, 0)
var c *gron.Cron

var initOnce = sync.Once{}

func Init() {
	initOnce.Do(func() {
		c = gron.New()
		for _, v := range TimerFunBox {
			c.AddFunc(gron.Every(v.diff), v.f)
		}
		for _, v := range TimeJobBox {
			c.Add(v.s, v.j)
		}
		c.Start()
	})
}

func RegisterFunc(diff time.Duration, f func()) {
	TimerFunBox = append(TimerFunBox, &Work{
		f:    f,
		diff: diff,
	})
}

func RegisterJob(s gron.Schedule, j gron.Job) {
	TimeJobBox = append(TimeJobBox, &Job{
		j: j,
		s: s,
	})
}
