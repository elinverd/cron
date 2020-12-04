package cron

import (
	"math/rand"
	"time"
)

type Jitter interface {
	// generates and returns a random duration in the set [-j.Deviation, j.Deviation)
	Generate() time.Duration
	WithSource(rand.Source) Jitter
	Max() time.Duration
}

func WrapWithJitter(schedule Schedule, jitter Jitter) Schedule {
	if jitter == nil {
		jitter = UniformJitter{}
	}
	return &ScheduleWithJitter{schedule: schedule, jitter: jitter}
}

type ScheduleWithJitter struct {
	schedule Schedule
	jitter   Jitter
}

func (s *ScheduleWithJitter) Next(t time.Time) time.Time {
	// when's next?
	ts := s.schedule.Next(t)
	// factor jitter in? (only in cases where it is necessary)
	if s.jitter != nil && s.jitter.Max() > 0 {
		// we don't want a job to run twice within the same jitter window
		// this happens in cases of negative jitter
		if ts.Sub(t) < s.jitter.Max() {
			ts = s.schedule.Next(t.Add(s.jitter.Max()))
		}
		if pt := ts.Add(s.jitter.Generate()); pt.After(t) {
			ts = pt
		}
	}
	return ts
}

type UniformJitter struct {
	Deviation time.Duration
}

func (j UniformJitter) WithSource(src rand.Source) Jitter {
	return uniformJitterWithSource{UniformJitter: j, source: src}
}

func (j UniformJitter) adjust(n int64) time.Duration {
	return time.Duration(n%(2*int64(j.Deviation)) - int64(j.Deviation))
}

func (j UniformJitter) Max() time.Duration {
	return j.Deviation
}

// generates and returns a random duration in the set [-j.Deviation, j.Deviation)
func (j UniformJitter) Generate() time.Duration {
	if j.Deviation <= 0 {
		return 0
	}
	return j.adjust(rand.Int63())
}

type uniformJitterWithSource struct {
	UniformJitter
	source rand.Source
}

func (j uniformJitterWithSource) Generate() time.Duration {
	if j.Deviation < time.Second {
		return 0
	}
	return j.adjust(j.source.Int63())
}
