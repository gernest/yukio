package events

import (
	"bytes"
	"strings"
	"time"

	"github.com/jinzhu/now"
	"github.com/prometheus/prometheus/pkg/labels"
)

const Day = 24 * time.Hour

const Month = 30 * Day

type Period uint

const (
	Realtime Period = iota
	Current
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
	Series    string
	Range     TimeRange
	Period    Period
	Interrval time.Duration
	Filters   map[string]*labels.Matcher
}

func (q *Query) Name(name string) *Query {
	q.Series = name
	return q
}

func (q *Query) filter(name string, fn func(f *labels.Matcher)) *Query {
	if q.Filters == nil {
		q.Filters = make(map[string]*labels.Matcher)
	}
	v, ok := q.Filters[name]
	if !ok {
		v = &labels.Matcher{Name: name}
		q.Filters[name] = v
	}
	fn(v)
	return q
}
func (q *Query) Equal(name, value string) *Query {

	return q.filter(name, func(f *labels.Matcher) {
		f.Type = labels.MatchEqual
		f.Value = value
	})
}

func (q *Query) Re(name, value string) *Query {
	return q.filter(name, func(f *labels.Matcher) {
		f.Type = labels.MatchEqual
		if f.Value != "" {
			f.Value += "|" + value
		}
	})
}

func (q *Query) NotRe(name, value string) *Query {
	return q.filter(name, func(f *labels.Matcher) {
		f.Type = labels.MatchNotRegexp
		if f.Value != "" {
			f.Value += "|" + value
		}
	})
}

func (q *Query) NotEqual(name, value string) *Query {
	return q.filter(name, func(f *labels.Matcher) {
		f.Type = labels.MatchNotEqual
		f.Value = value
	})
}

func Now() time.Time {
	return time.Now().UTC()
}

func QueryFrom(period Period, nowFunc func() time.Time) *Query {
	switch period {
	default:
		end := nowFunc()
		start := end.Add(-30 * time.Minute)
		return &Query{
			Period:    period,
			Interrval: time.Minute,
			Range: TimeRange{
				Start: start,
				End:   end,
			},
		}
	case Current:
		end := nowFunc()
		start := end.Add(-5 * time.Minute)
		return &Query{
			Period:    period,
			Interrval: time.Minute,
			Range: TimeRange{
				Start: start,
				End:   end,
			},
		}
	case ADay:
		n := nowFunc()
		return &Query{
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
		return &Query{
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
		return &Query{
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
		return &Query{
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
		return &Query{
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
		return &Query{
			Period:    period,
			Interrval: Month,
			Range: TimeRange{
				Start: start,
				End:   end,
			},
		}
	}
}

func (q *Query) Write(b *bytes.Buffer) {
	b.WriteString(q.Series)
	var i int
	for _, s := range q.Filters {
		if i == 0 {
			b.WriteRune('{')
		} else {
			b.WriteRune(',')
		}
		b.WriteString(s.String())
		i++
	}
	if i > 0 {
		b.WriteRune('}')
	}
	x := q.Range.End.Sub(q.Range.Start)
	if x > 0 {
		b.WriteRune('[')
		b.WriteString(x.String())
		b.WriteRune(']')
	}
}

func (q *Query) Request(modifirer ...func(string) string) QueryRangeRequest {
	var b bytes.Buffer
	b.WriteString(q.Series)
	var i int
	for _, s := range q.Filters {
		if i == 0 {
			b.WriteRune('{')
		} else {
			b.WriteRune(',')
		}
		b.WriteString(s.String())
		i++
	}
	if i > 0 {
		b.WriteRune('}')
	}
	x := q.Range.End.Sub(q.Range.Start)
	if x > 0 {
		b.WriteRune('[')
		b.WriteString(x.String())
		b.WriteRune(']')
	}
	s := b.String()
	for _, v := range modifirer {
		s = v(s)
	}
	return QueryRangeRequest{
		Start: q.Range.Start,
		End:   q.Range.End,
		Query: s,
	}
}

func GroupBy(q string, labels ...string) string {
	var b bytes.Buffer
	b.WriteString("group by (")
	b.WriteString(strings.Join(labels, ","))
	b.WriteRune(')')
	b.WriteRune('(')
	b.WriteString(q)
	b.WriteRune(')')
	return b.String()
}
