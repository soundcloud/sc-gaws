package stats

import (
	"time"
)

// Metric represents a single metric gathered in an application.
// It maps roughly to the CloudWatch MetricDatum data type. Some
// fields are left out for simplification.
//
// More info: http://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_MetricDatum.html
type Metric struct {
	Name      string
	Value     float32
	Unit      string
	Timestamp time.Time
}

// Stats pusher is an interface that wraps a method Push that can be called to
// push metrics to an aggregator of some kind, like AWS Cloudwatch
type StatsPusher interface {
	Push(metrics []Metric)
}

// Stats represents accumulated metrics inside of a specific Namespace,
// which can be pushed upstream to a service like AWS CloudWatch (or any other)
type Stats struct {

	// Used to push stats upstream
	Pusher StatsPusher

	// Number of samples that will be accumulated to create one metric for
	// sending upstream.  This is to prevent pushing too many metrics.
	AccumulateLimit int

	// Samples currently being collected
	currentSamples map[string][]Metric
}

func NewStats(pusher StatsPusher, accumulateLimit int) *Stats {
	return &Stats{
		pusher,
		accumulateLimit,
		make(map[string][]Metric),
	}
}

// Add a metric to the Stats struct, with averaging applied.
func (s *Stats) addMetric(m Metric) {
	s.currentSamples[m.Name] = append(s.currentSamples[m.Name], m)
}

func (s *Stats) accumulate() []Metric {
	var allMetrics []Metric
	for name, metrics := range s.currentSamples {

		if len(metrics) == 0 {
			// Nothing to accumulate
			continue
		}
		n := len(metrics) / s.AccumulateLimit
		if len(metrics)%s.AccumulateLimit > 0 {
			n = n + 1
		}
		for i := 0; i < n; i++ {
			var currMetrics []Metric
			// Get the right sample to accumulate
			if i == n-1 {
				currMetrics = metrics[(i * s.AccumulateLimit):]
			} else {
				currMetrics = metrics[(i * s.AccumulateLimit):((i + 1) * s.AccumulateLimit)]
			}

			// Use last metric in sample as reference
			m := currMetrics[len(currMetrics)-1]

			// Accumulate based on type of metric. Currently, only Count units
			// have special accumulation behavior. Everything else just falls
			// back to averaging
			switch m.Unit {
			case "Count":
				sum := float32(0.0)
				for _, v := range currMetrics {
					sum += v.Value
				}
				allMetrics = append(allMetrics, Metric{m.Name, sum, m.Unit, m.Timestamp})
			default:
				sum := float32(0.0)
				for _, v := range currMetrics {
					sum += v.Value
				}
				avg := sum / float32(len(currMetrics))
				allMetrics = append(allMetrics, Metric{m.Name, avg, m.Unit, m.Timestamp})
			}
		}
		s.currentSamples[name] = metrics[:0]
	}
	return allMetrics
}

// Accumulate and pushes metrics upstream. This is designed to run in a
// goroutine.
//
// The duration of pushing can be controlled by statsUpdateFrequency. Metrics
// are received on metricChan.
func (s *Stats) AccumulateAndPush(statsUpdateFrequency time.Duration, metricChan <-chan Metric) {
	t := time.Tick(statsUpdateFrequency)
	for {
		select {
		case <-t:
			metrics := s.accumulate()
			if len(metrics) > 0 {
				go s.Pusher.Push(metrics)
			}
		case m := <-metricChan:
			s.addMetric(m)
		}
	}
}
