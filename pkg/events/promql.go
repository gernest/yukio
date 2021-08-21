package events

import (
	"time"

	"github.com/jinzhu/now"
)

const Day = 24 * time.Hour

const Month = 30 * Day

type Period uint

const (
	Realtime Period = iota
	ADay
	AWeek
	AMonth
	AFixedMonth //30 days
	HalfAYear
	AYear
)

func (p Period) String() string {
	switch p {
	case ADay:
		return "day"
	case AWeek:
		return "7d"
	case AMonth:
		return "month"
	case AFixedMonth:
		return "30d"
	case HalfAYear:
		return "6mo"
	case AYear:
		return "12mo"
	default:
		return "realtime"
	}
}

type TimeRange struct {
	Start, End time.Time
}

type Query struct {
	Range     TimeRange
	Period    Period
	Interrval time.Duration
}

func Now() time.Time {
	return time.Now().UTC()
}

func QueryFrom(period Period, nowFunc func() time.Time) Query {
	switch period {
	default:
		n := nowFunc()
		return Query{
			Period:    period,
			Interrval: time.Minute,
			Range: TimeRange{
				Start: n,
				End:   n,
			},
		}
	case ADay:
		n := nowFunc()
		return Query{
			Period:    period,
			Interrval: time.Hour,
			Range: TimeRange{
				Start: n,
				End:   n,
			},
		}
	case AWeek:
		end := nowFunc()
		start := end.Add(-6 * Day)
		return Query{
			Period:    period,
			Interrval: Day,
			Range: TimeRange{
				Start: start,
				End:   end,
			},
		}
	case AFixedMonth:
		end := nowFunc()
		start := end.Add(-30 * Day)
		return Query{
			Period:    period,
			Interrval: Day,
			Range: TimeRange{
				Start: start,
				End:   end,
			},
		}
	case AMonth:
		n := nowFunc()
		start := now.With(n).BeginningOfMonth()
		end := now.With(n).EndOfMonth()
		return Query{
			Period:    period,
			Interrval: Day,
			Range: TimeRange{
				Start: start,
				End:   end,
			},
		}
	case HalfAYear:
		n := nowFunc()
		end := now.With(n).EndOfMonth()
		start := now.With(n.Add(-5 * Month)).BeginningOfMonth()
		return Query{
			Period:    period,
			Interrval: Month,
			Range: TimeRange{
				Start: start,
				End:   end,
			},
		}
	case AYear:
		n := nowFunc()
		end := now.With(n).EndOfMonth()
		start := now.With(n.Add(-11 * Month)).BeginningOfMonth()
		return Query{
			Period:    period,
			Interrval: Month,
			Range: TimeRange{
				Start: start,
				End:   end,
			},
		}
	}
}
