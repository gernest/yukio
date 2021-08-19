package events

import (
	"context"
	"math"
	"strconv"
	"time"

	"github.com/gernest/yukio/pkg/loga"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
	"github.com/prometheus/prometheus/promql"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/storage/remote"
	"go.uber.org/zap"
)

var registry = prometheus.NewRegistry()

var store *remote.Client

func register(c ...prometheus.Collector) {
	registry.MustRegister(c...)
}

func WriteLoop(ctx context.Context, write remote.WriteClient, flush time.Duration) {
	tick := time.NewTicker(flush)
	defer tick.Stop()
	log := loga.Get(ctx).Named("ts-write-loop")
	log.Info("started write loop for events")
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			m, err := registry.Gather()
			if err != nil {
				log.Error("Failed to gather metrics",
					zap.Error(err))
				continue
			}
			s := createRequest(m)
			if s.Size() == 0 {
				continue
			}
			b, err := proto.Marshal(&s)
			if err != nil {
				continue
			}
			err = write.Store(ctx, snappy.Encode(nil, b))
			if err != nil {
				log.Error("Failed to store series to remote store",
					zap.Error(err))
				continue
			}
			log.Info("Sync", zap.Int("size", s.Size()))
		}
	}
}

func createRequest(m []*dto.MetricFamily) (r prompb.WriteRequest) {
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
		counter(r, m)
	case dto.MetricType_GAUGE:
		gauge(r, m)
	case dto.MetricType_SUMMARY:
		summary(r, m)
	case dto.MetricType_HISTOGRAM:
		histogram(r, m)
	default:
		untyped(r, m)
	}
}

func counter(r *prompb.WriteRequest, f *dto.MetricFamily) {
	var ts prompb.TimeSeries
	name := f.GetName()
	for _, s := range f.GetMetric() {
		v := s.GetCounter()
		if v == nil {
			return
		}
		sp, lb := writeSample(
			name, s, "", 0, v.GetValue(),
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

func gauge(r *prompb.WriteRequest, f *dto.MetricFamily) {
	var ts prompb.TimeSeries
	name := f.GetName()
	for _, s := range f.GetMetric() {
		v := s.GetCounter()
		if v == nil {
			return
		}
		sp, lb := writeSample(
			name, s, "", 0, v.GetValue(),
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

func summary(r *prompb.WriteRequest, f *dto.MetricFamily) {
	var ts, sum, count prompb.TimeSeries
	name := f.GetName()
	for _, s := range f.GetMetric() {
		v := s.GetSummary()
		if v == nil {
			return
		}
		for _, q := range v.GetQuantile() {
			sample, label := writeSample(
				name, s, model.QuantileLabel, q.GetQuantile(), q.GetValue(),
			)
			ts.Samples = append(ts.Samples, sample)
			ts.Labels = append(ts.Labels, label...)
		}
		sample, label := writeSample(name+"_sum", s, "", 0, v.GetSampleSum())
		sum.Samples = append(sum.Samples, sample)
		sum.Labels = append(sum.Labels, label...)

		sample, label = writeSample(name+"_count", s, "", 0, float64(v.GetSampleCount()))
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

func histogram(r *prompb.WriteRequest, f *dto.MetricFamily) {
	name := f.GetName()
	bucketLabelName := name + "_bucket"
	sumLabelName := name + "_sum"
	countLabelName := name + "_count"
	var ts, sum, count prompb.TimeSeries
	for _, s := range f.GetMetric() {
		v := s.GetHistogram()
		if v == nil {
			return
		}
		infSeen := false
		for _, b := range v.GetBucket() {
			sample, label := writeSample(
				bucketLabelName, s, model.BucketLabel, b.GetUpperBound(), float64(b.GetCumulativeCount()),
			)
			if math.IsInf(b.GetUpperBound(), +1) {
				infSeen = true
			}
			ts.Samples = append(ts.Samples, sample)
			ts.Labels = append(ts.Labels, label...)
		}
		if infSeen {
			sample, label := writeSample(
				bucketLabelName, s, model.BucketLabel, math.Inf(+1), float64(v.GetSampleCount()),
			)
			ts.Samples = append(ts.Samples, sample)
			ts.Labels = append(ts.Labels, label...)
		}
		sample, label := writeSample(sumLabelName, s, "", 0, v.GetSampleSum())
		sum.Samples = append(sum.Samples, sample)
		sum.Labels = append(sum.Labels, label...)

		sample, label = writeSample(countLabelName, s, "", 0, float64(v.GetSampleCount()))
		count.Samples = append(count.Samples, sample)
		count.Labels = append(count.Labels, label...)
	}
	r.Metadata = append(r.Metadata, prompb.MetricMetadata{
		Type:             metaType(f.GetType()),
		MetricFamilyName: bucketLabelName,
		Help:             f.GetHelp(),
	})
	r.Timeseries = append(r.Timeseries, ts)
	r.Metadata = append(r.Metadata, prompb.MetricMetadata{
		Type:             metaType(f.GetType()),
		MetricFamilyName: sumLabelName,
		Help:             f.GetHelp(),
	})
	r.Timeseries = append(r.Timeseries, sum)
	r.Metadata = append(r.Metadata, prompb.MetricMetadata{
		Type:             metaType(f.GetType()),
		MetricFamilyName: countLabelName,
		Help:             f.GetHelp(),
	})
	r.Timeseries = append(r.Timeseries, count)
}

func untyped(r *prompb.WriteRequest, f *dto.MetricFamily) {
	var ts prompb.TimeSeries
	name := f.GetName()
	for _, s := range f.GetMetric() {
		v := s.GetCounter()
		if v == nil {
			return
		}
		sp, lb := writeSample(
			name, s, "", 0, v.GetValue(),
		)
		ts.Samples = append(ts.Samples, sp)
		ts.Labels = append(ts.Labels, lb...)
	}
	r.Timeseries = append(r.Timeseries, ts)
}
func writeSample(
	name string,
	metric *dto.Metric,
	additionalLabelName string, additionalLabelValue float64,
	value float64,
) (s prompb.Sample, lp []prompb.Label) {
	lp = writeLabelPairs(
		name, metric.Label, additionalLabelName, additionalLabelValue,
	)
	s.Value = value
	if metric.TimestampMs != nil {
		s.Timestamp = metric.GetTimestampMs()
	}
	return
}

func writeLabelPairs(
	name string,
	in []*dto.LabelPair,
	additionalLabelName string, additionalLabelValue float64,
) (ls []prompb.Label) {
	if len(in) == 0 && additionalLabelName == "" {
		return
	}
	ls = append(ls, prompb.Label{
		Name:  model.MetricNameLabel,
		Value: name,
	})
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

type Engine struct {
	Queryable storage.Queryable
	Engine    *promql.Engine
}

type engineKey struct{}

func SetupQuery(ctx context.Context, log *zap.Logger, read remote.ReadClient) context.Context {
	chunk := remote.NewSampleAndChunkQueryableClient(
		read, nil, nil, false, startTime,
	)
	engine := promql.NewEngine(promql.EngineOpts{})
	return context.WithValue(ctx, engineKey{}, &Engine{
		Queryable: chunk,
		Engine:    engine,
	})
}

type QueryRequest struct {
	Timestamp time.Time
	Query     string
}

func InstantQuery(ctx context.Context, req *QueryRequest) (*promql.Result, error) {
	e := get(ctx)
	q, err := e.Engine.NewInstantQuery(e.Queryable, req.Query, req.Timestamp)
	if err != nil {
		return nil, err
	}
	res := q.Exec(ctx)
	if res.Err != nil {
		return nil, err
	}
	return res, nil
}

func startTime() (int64, error) {
	return int64(model.Latest), nil
}

type QueryRangeRequest struct {
	Start, End time.Time
	Query      string
	Step       time.Duration
}

func RangeQuery(ctx context.Context, req *QueryRangeRequest) (*promql.Result, error) {
	e := get(ctx)
	q, err := e.Engine.NewRangeQuery(e.Queryable, req.Query, req.Start, req.End, req.Step)
	if err != nil {
		return nil, err
	}
	res := q.Exec(ctx)
	if res.Err != nil {
		return nil, err
	}
	return res, nil
}

func get(ctx context.Context) *Engine {
	return ctx.Value(engineKey{}).(*Engine)
}
