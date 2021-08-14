package events

import (
	"context"
	"math"
	"strconv"
	"time"

	"github.com/gernest/yukio/pkg/config"
	"github.com/gogo/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
	"github.com/prometheus/prometheus/storage/remote"
)

var registry = prometheus.NewRegistry()

var store *remote.Client

func register(c ...prometheus.Collector) {
	registry.MustRegister(c...)
}

func WriteLoop(ctx context.Context, write remote.WriteClient, c *config.Config) {
	tick := time.NewTicker(c.TimeSeries.FlushInterval)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			m, err := registry.Gather()
			if err != nil {
				continue
			}
			s := sample(m)
			b, err := proto.Marshal(&s)
			if err != nil {
				continue
			}
			err = write.Store(ctx, b)
			if err != nil {
				// log
			}
		}
	}
}

func sample(m []*dto.MetricFamily) (r prompb.WriteRequest) {
	for _, mf := range m {
		timeseries(&r, mf)
	}
	return
}

func metaType(a dto.MetricType) prompb.MetricMetadata_MetricType {
	switch a {
	case dto.MetricType_COUNTER:
		return prompb.MetricMetadata_COUNTER
	case dto.MetricType_GAUGE:
		return prompb.MetricMetadata_GAUGE
	case dto.MetricType_SUMMARY:
		return prompb.MetricMetadata_SUMMARY
	case dto.MetricType_HISTOGRAM:
		return prompb.MetricMetadata_HISTOGRAM
	default:
		return prompb.MetricMetadata_UNKNOWN
	}
}

func timeseries(r *prompb.WriteRequest, m *dto.MetricFamily) {
	switch m.GetType() {
	case dto.MetricType_COUNTER:
		counter(r, m, m.GetMetric())
	case dto.MetricType_GAUGE:
		gauge(r, m, m.GetMetric())
	case dto.MetricType_SUMMARY:
		summary(r, m, m.GetMetric())
	case dto.MetricType_HISTOGRAM:
		histogram(r, m, m.GetMetric())
	default:
		untyped(r, m.GetMetric())
	}
}

func counter(r *prompb.WriteRequest, f *dto.MetricFamily, m []*dto.Metric) {
	var ts prompb.TimeSeries
	for _, s := range m {
		v := s.GetCounter()
		if v == nil {
			return
		}
		sp, lb := writeSample(
			s, "", 0, v.GetValue(),
		)

		ts.Samples = append(ts.Samples, sp)
		ts.Labels = append(ts.Labels, lb...)
	}
	r.Metadata = append(r.Metadata, prompb.MetricMetadata{
		Type:             metaType(f.GetType()),
		MetricFamilyName: f.GetName(),
		Help:             f.GetHelp(),
	})
	r.Timeseries = append(r.Timeseries, ts)
}

func gauge(r *prompb.WriteRequest, f *dto.MetricFamily, m []*dto.Metric) {
	var ts prompb.TimeSeries
	for _, s := range m {
		v := s.GetCounter()
		if v == nil {
			return
		}
		sp, lb := writeSample(
			s, "", 0, v.GetValue(),
		)

		ts.Samples = append(ts.Samples, sp)
		ts.Labels = append(ts.Labels, lb...)
	}
	r.Metadata = append(r.Metadata, prompb.MetricMetadata{
		Type:             metaType(f.GetType()),
		MetricFamilyName: f.GetName(),
		Help:             f.GetHelp(),
	})
	r.Timeseries = append(r.Timeseries, ts)
}

func summary(r *prompb.WriteRequest, f *dto.MetricFamily, m []*dto.Metric) {
	var ts, sum, count prompb.TimeSeries
	for _, s := range m {
		v := s.GetSummary()
		if v == nil {
			return
		}
		for _, q := range v.GetQuantile() {
			sample, label := writeSample(
				s, model.QuantileLabel, q.GetQuantile(), q.GetValue(),
			)
			ts.Samples = append(ts.Samples, sample)
			ts.Labels = append(ts.Labels, label...)
		}
		sample, label := writeSample(s, "", 0, v.GetSampleSum())
		sum.Samples = append(sum.Samples, sample)
		sum.Labels = append(sum.Labels, label...)

		sample, label = writeSample(s, "", 0, float64(v.GetSampleCount()))
		count.Samples = append(count.Samples, sample)
		count.Labels = append(count.Labels, label...)
	}
	r.Metadata = append(r.Metadata, prompb.MetricMetadata{
		Type:             metaType(f.GetType()),
		MetricFamilyName: f.GetName(),
		Help:             f.GetHelp(),
	})
	r.Timeseries = append(r.Timeseries, ts)
	r.Metadata = append(r.Metadata, prompb.MetricMetadata{
		Type:             metaType(f.GetType()),
		MetricFamilyName: f.GetName() + "_sum",
		Help:             f.GetHelp(),
	})
	r.Timeseries = append(r.Timeseries, sum)
	r.Metadata = append(r.Metadata, prompb.MetricMetadata{
		Type:             metaType(f.GetType()),
		MetricFamilyName: f.GetName() + "_count",
		Help:             f.GetHelp(),
	})
	r.Timeseries = append(r.Timeseries, count)
}

func histogram(r *prompb.WriteRequest, f *dto.MetricFamily, m []*dto.Metric) {
	var ts, sum, count prompb.TimeSeries
	for _, s := range m {
		v := s.GetHistogram()
		if v == nil {
			return
		}
		infSeen := false
		for _, b := range v.GetBucket() {
			sample, label := writeSample(
				s, model.BucketLabel, b.GetUpperBound(), float64(b.GetCumulativeCount()),
			)
			if math.IsInf(b.GetUpperBound(), +1) {
				infSeen = true
			}
			ts.Samples = append(ts.Samples, sample)
			ts.Labels = append(ts.Labels, label...)
		}
		if infSeen {
			sample, label := writeSample(
				s, model.BucketLabel, math.Inf(+1), float64(v.GetSampleCount()),
			)
			ts.Samples = append(ts.Samples, sample)
			ts.Labels = append(ts.Labels, label...)
		}
		sample, label := writeSample(s, "", 0, v.GetSampleSum())
		sum.Samples = append(sum.Samples, sample)
		sum.Labels = append(sum.Labels, label...)

		sample, label = writeSample(s, "", 0, float64(v.GetSampleCount()))
		count.Samples = append(count.Samples, sample)
		count.Labels = append(count.Labels, label...)
	}
	r.Metadata = append(r.Metadata, prompb.MetricMetadata{
		Type:             metaType(f.GetType()),
		MetricFamilyName: f.GetName() + "_bucket",
		Help:             f.GetHelp(),
	})
	r.Timeseries = append(r.Timeseries, ts)
	r.Metadata = append(r.Metadata, prompb.MetricMetadata{
		Type:             metaType(f.GetType()),
		MetricFamilyName: f.GetName() + "_sum",
		Help:             f.GetHelp(),
	})
	r.Timeseries = append(r.Timeseries, sum)
	r.Metadata = append(r.Metadata, prompb.MetricMetadata{
		Type:             metaType(f.GetType()),
		MetricFamilyName: f.GetName() + "_count",
		Help:             f.GetHelp(),
	})
	r.Timeseries = append(r.Timeseries, count)
}

func untyped(r *prompb.WriteRequest, m []*dto.Metric) {
	var ts prompb.TimeSeries
	for _, s := range m {
		v := s.GetCounter()
		if v == nil {
			return
		}
		sp, lb := writeSample(
			s, "", 0, v.GetValue(),
		)
		ts.Samples = append(ts.Samples, sp)
		ts.Labels = append(ts.Labels, lb...)
	}
	r.Timeseries = append(r.Timeseries, ts)
}
func writeSample(
	metric *dto.Metric,
	additionalLabelName string, additionalLabelValue float64,
	value float64,
) (s prompb.Sample, lp []prompb.Label) {
	lp = writeLabelPairs(
		metric.Label, additionalLabelName, additionalLabelValue,
	)
	s.Value = value
	if metric.TimestampMs != nil {
		s.Timestamp = metric.GetTimestampMs()
	}
	return
}

func writeLabelPairs(
	in []*dto.LabelPair,
	additionalLabelName string, additionalLabelValue float64,
) (ls []prompb.Label) {
	if len(in) == 0 && additionalLabelName == "" {
		return
	}
	for _, lp := range in {

		ls = append(ls, prompb.Label{
			Name:  lp.GetName(),
			Value: lp.GetValue(),
		})
	}
	if additionalLabelName != "" {
		ls = append(ls, prompb.Label{
			Name:  additionalLabelName,
			Value: strconv.FormatFloat(additionalLabelValue, 'f', -1, 64),
		})
	}
	return
}
