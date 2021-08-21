package events

import (
	"github.com/gernest/yukio/pkg/models"
	"github.com/golang/protobuf/ptypes"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	visitDuration = metricname("visit_duration")
	visits        = metricname("visits")
	bounceRate    = metricname("bounce_rate")
)

var VisitDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name: "visit_duration",
		Help: "Tracks how long a visitor stays on a page",
	},
	[]string{
		Domain,
		SessionID,
		Referer,
		EntryPage,
		ExitPage,
	},
)

var BounceRate = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "bounce_rate",
		Help: "Counts a single page view",
	},
	[]string{
		Domain,
		SessionID,
		Referer,
		EntryPage,
	},
)

var Visits = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "visits",
		Help: "Counts a number of events per session",
	},
	[]string{
		Domain,
		SessionID,
		Referer,
		EntryPage,
		ExitPage,
	},
)

func init() {
	register(VisitDuration, Visits)
}

func RecordSession(s *models.Session) {
	duration, _ := ptypes.Duration(s.Duration)
	id := formatNumber(s.Id)
	VisitDuration.WithLabelValues(
		s.Domain,
		id,
		s.Referrer,
		s.EntryPage,
		s.ExitPage,
	).Observe(float64(duration.Milliseconds()))
	if s.IsBounce {
		BounceRate.WithLabelValues(
			s.Domain,
			id,
			s.Referrer,
			s.EntryPage,
		).Inc()
	}
	Visits.WithLabelValues(
		s.Domain,
		id,
		s.Referrer,
		s.EntryPage,
		s.ExitPage,
	).Add(float64(s.Events))
}

func sessionSeries(m ...string) (r []metricname) {
	for _, v := range m {
		switch metricname(v) {
		case visitDuration, visits, bounceRate:
			r = append(r, metricname(v))
		}
	}
	return
}
