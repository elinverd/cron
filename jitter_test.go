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
	src := &mockRandSource{seq: []int64{
		int64(17 * time.Minute),
		int64(-2 * time.Minute),
		int64(11 * time.Minute),
	}}

	sched := WrapWithJitter(Every(5*time.Minute), UniformJitter{5 * time.Minute}.WithSource(src))
	var t1, t2, t3 time.Time
	t1, _ = time.Parse(time.RFC3339, "2020-10-08T10:00:00Z")

	expected2, _ := time.Parse(time.RFC3339, "2020-10-08T10:07:00Z")
	if t2 = sched.Next(t1); t2 != expected2 {
		t.Errorf("expected %s, got %s", expected2, t2)
	}
	expected3, _ := time.Parse(time.RFC3339, "2020-10-08T10:12:00Z")
	if t3 = sched.Next(t2); t3 != expected3 {
		t.Errorf("expected %s, got %s", expected3, t3)
	}

}
