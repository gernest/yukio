package events

import (
	"github.com/gernest/yukio/pkg/models"
	"github.com/prometheus/client_golang/prometheus"
)

var VisitDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name: "visit_duration",
		Help: "Tracks how long a visitor stays on a page",
	},
	[]string{
		Domain,
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
		Referer,
		Path,
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
		Referer,
		Path,
		EntryPage,
		ExitPage,
	},
)

func RecordSession(s *models.Session) {
	VisitDuration.WithLabelValues(
		s.Domain, s.Referrer, s.EntryPage, s.ExitPage,
	).Observe(float64(s.Duration.Milliseconds()))
	if s.IsBounce {
		BounceRate.WithLabelValues(
			s.Domain, s.Referrer, s.EntryPage,
		).Inc()
	}
	Visits.WithLabelValues(
		s.Domain, s.Referrer, s.EntryPage, s.ExitPage,
	).Add(float64(s.Events))
}
