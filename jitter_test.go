package cron

import (
	"math/rand"
	"testing"
	"time"
)

type mockRandSource struct {
	seq []int64
	idx int
}

func (src *mockRandSource) Int63() int64 {
	if len(src.seq) == 0 {
		return 0
	}
	n := src.seq[src.idx%len(src.seq)]
	src.idx++
	return n
}
func (src *mockRandSource) Seed(seed int64) {}

func TestJitterGenerate(t *testing.T) {
	tests := []struct {
		jitter   Jitter
		source   rand.Source
		expected time.Duration
	}{
		{UniformJitter{}, &mockRandSource{}, 0},
		{UniformJitter{100 * time.Millisecond}, &mockRandSource{}, 0},
		{UniformJitter{5 * time.Minute}, &mockRandSource{seq: []int64{int64(17 * time.Minute)}}, 2 * time.Minute},
		{UniformJitter{5 * time.Minute}, &mockRandSource{seq: []int64{int64(13 * time.Minute)}}, -2 * time.Minute},
		{UniformJitter{5 * time.Minute}, &mockRandSource{seq: []int64{int64(21 * time.Minute)}}, -4 * time.Minute},
		{UniformJitter{5 * time.Minute}, &mockRandSource{seq: []int64{int64(20 * time.Minute)}}, -5 * time.Minute},
		{UniformJitter{5 * time.Minute}, &mockRandSource{seq: []int64{int64(15 * time.Minute)}}, 0},
		{UniformJitter{5 * time.Minute}, &mockRandSource{seq: []int64{int64(19*time.Minute + 59*time.Second)}}, 4*time.Minute + 59*time.Second},
	}
	for _, test := range tests {
		actual := test.jitter.WithSource(test.source).Generate()
		if actual != test.expected {
			t.Errorf("%v: (expected) %v != %v (actual)", test.jitter, test.expected, actual)
		}
	}
}

func TestScheduleWithJitterNext(t *testing.T) {
	sched, err := ParseStandard("JITTER=10m 30 7 * * *")
	if err != nil {
		t.Error(err)
	}
	// inject a mock rand source into the scheduler
	src := &mockRandSource{seq: []int64{
		int64(17 * time.Minute), // + 7m
		int64(3 * time.Minute),  // - 7m
		int64(11 * time.Minute), // + 1m
	}}
	s, ok := sched.(*ScheduleWithJitter)
	if !ok {
		t.Error("Failed to convert type")
	}
	s.jitter = s.jitter.WithSource(src)

	var now, expected, t1, t2, t3 time.Time
	now, _ = time.Parse(time.RFC3339, "2020-12-07T07:31:00Z")

	expected, _ = time.Parse(time.RFC3339, "2020-12-08T07:37:00Z")
	if t1 = sched.Next(now); t1 != expected {
		t.Errorf("expected %s, got %s", expected, t1)
	}

	expected, _ = time.Parse(time.RFC3339, "2020-12-09T07:23:00Z")
	if t2 = sched.Next(t1); t2 != expected {
		t.Errorf("expected %s, got %s", expected, t2)
	}

	expected, _ = time.Parse(time.RFC3339, "2020-12-10T07:31:00Z")
	if t3 = sched.Next(t2); t3 != expected {
		t.Errorf("expected %s, got %s", expected, t3)
	}
}
