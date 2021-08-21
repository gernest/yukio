package events

import (
	"testing"
	"time"
)

func TestQuery(t *testing.T) {
	ts, err := time.Parse(time.RFC822Z, time.RFC822Z)
	if err != nil {
		t.Fatal(err)
	}
	ts = ts.UTC()
	nowFunc := func() time.Time {
		return ts
	}
	sample := []struct {
		period   Period
		duration int //in days
		interval time.Duration
	}{
		{period: Realtime, interval: time.Minute},
		{period: ADay, interval: time.Hour},
		{period: AWeek, interval: Day, duration: 6},
		{period: AFixedMonth, interval: Day, duration: 30},
		{period: AMonth, interval: Day, duration: 30},
		{period: HalfAYear, interval: 30 * Day, duration: 183},
		{period: AYear, interval: 30 * Day, duration: 364},
	}
	for _, s := range sample {
		t.Run(s.period.String(), func(t *testing.T) {
			q := QueryFrom(s.period, nowFunc)
			got := int(q.Range.End.Sub(q.Range.Start) / Day)
			expect := s.duration
			if got != expect {
				t.Errorf("mismatch time range expected %d got %d", s.duration, got)
			}
			if q.Interrval != s.interval {
				t.Errorf("mismatch interval expected %s got %s", s.interval, q.Interrval)
			}
		})
	}
}
