// Tests for stats.go
package stats

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

var (
	stats []Metric
	rGen  *rand.Rand
	mu    sync.Mutex
)

const (
	pushFrequency   = 2 * time.Second
	accumulateLimit = 10
)

func init() {
	rGen = rand.New(rand.NewSource(time.Now().Unix()))
}

// MockStatsPusher implements the StatsPusher interface and just
// collects metrics which we can
type MockStatsPusher struct{}

func (m MockStatsPusher) Push(metrics []Metric) {
	setStats(metrics)
}

func newStatsChannel() chan<- Metric {
	m := make(chan Metric)

	s := &Stats{
		Pusher:          &MockStatsPusher{},
		AccumulateLimit: accumulateLimit,
		currentSamples:  make(map[string][]Metric),
	}

	go s.AccumulateAndPush(pushFrequency, m)

	return m
}

func TestSingleMetric(t *testing.T) {
	setStats(nil)
	metrics := generateRandomMetrics(accumulateLimit*2, "TestMetrics1", "Milliseconds")
	c := newStatsChannel()

	for i := 0; i < 2; i++ {
		// Check that we havent seen anything after (accumulateLimit-1) metrics
		avg := float32(0.0)
		start := i * accumulateLimit
		for _, m := range metrics[start:(start + accumulateLimit)] {
			avg += m.Value
			c <- m
		}

		avg = avg / float32(accumulateLimit)

		waitForPush()

		// Assert that one metric with the right average has been pushed
		if len(stats) != 1 {
			t.Fatalf("1 metric should have been pushed. %d metrics found.", len(stats))
		}

		if stats[0].Value != avg {
			t.Fatalf("Expected metric averaged value: %d. Found value: %d", avg, stats[0].Value)
		}

		if stats[0].Name != "TestMetrics1" {
			t.Fatalf("Metric name '%s' does not match expected '%s'", stats[0].Name, "TestMetrics1")
		}

		if stats[0].Timestamp != metrics[start+accumulateLimit-1].Timestamp {
			t.Fatal("Metric timestamp should be set")
		}

		setStats(nil)
	}
}

func TestMultipleMetrics(t *testing.T) {
	setStats(nil)
	c := newStatsChannel()

	metrics1 := generateRandomMetrics(accumulateLimit, "TestMetrics1", "Milliseconds")
	metrics2 := generateRandomMetrics(accumulateLimit, "TestMetrics2", "Milliseconds")

	// Check that we havent seen anything after (accumulateLimit-1) metrics
	avg1 := float32(0.0)
	avg2 := float32(0.0)
	for i := 0; i < accumulateLimit; i++ {
		m1 := metrics1[i]
		m2 := metrics2[i]
		avg1 += m1.Value
		avg2 += m2.Value
		c <- m1
		c <- m2
	}

	avg1 = avg1 / float32(accumulateLimit)
	avg2 = avg2 / float32(accumulateLimit)

	waitForPush()

	// Assert that one metric with the right average has been pushed
	if len(stats) != 2 {
		t.Fatalf("2 metric should have been pushed. %d metrics found.", len(stats))
	}
}

func TestCountMetrics(t *testing.T) {
	setStats(nil)
	c := newStatsChannel()
	metrics := generateRandomMetrics(accumulateLimit, "TestMetrics1Count", "Count")
	sum := float32(0.0)
	for i := 0; i < accumulateLimit; i++ {
		sum += metrics[i].Value
		c <- metrics[i]
	}

	waitForPush()

	// Assert that one metric with the right count has been pushed
	if len(stats) != 1 {
		t.Fatalf("1 metric should have been pushed. %d metrics found.", len(stats))
	}

	if stats[0].Value != sum {
		t.Fatalf("expected metric value: %d. found value: %d", sum, stats[0].Value)
	}
}

func generateRandomMetrics(count int, name string, unit string) (metrics []Metric) {

	for i := 0; i < count; i++ {
		val := rGen.Float32()*50.0 + 200.0
		metric := Metric{Name: name,
			Value:     val,
			Unit:      unit,
			Timestamp: time.Now(),
		}
		metrics = append(metrics, metric)
	}
	return
}

func waitForPush() {
	setStats(nil)
	fmt.Println("Waiting for push attempt")
	time.Sleep(pushFrequency + 1*time.Second)
}

func setStats(metrics []Metric) {
	mu.Lock()
	stats = metrics
	defer mu.Unlock()
}
