package events

import (
	"github.com/gernest/yukio/pkg/models"
	"github.com/prometheus/client_golang/prometheus"
)

var PageDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name: "page_duration",
		Help: "Tracks how long a visitor stays on a page",
	},
	[]string{
		Domain,
		Referer,
		EntryPage,
		ExitPage,
	},
)

var Bounce = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "bounce",
		Help: "Counts a single page view",
	},
	[]string{
		Domain,
		Referer,
		Path,
		EntryPage,
	},
)

var Events = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "session_events_count",
		Help: "Counts a number of events per session",
	},
	[]string{
		Domain,
		Referer,
		Path,
		EntryPage,
	},
)

func RecordSession(s *models.Session) {
	PageDuration.WithLabelValues(
		s.Domain, s.Referrer, s.EntryPage, s.ExitPage,
	).Observe(float64(s.Duration.Milliseconds()))
	if s.IsBounce {
		Bounce.WithLabelValues(
			s.Domain, s.Referrer, s.EntryPage,
		).Add(float64(s.Events))
	}
}
