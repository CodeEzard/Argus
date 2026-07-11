package detector

import (
	"math"
)

// Anomaly represents a detected problem in a metric.
type Anomaly struct {
	MetricName string
	Value      float64
	ZScore     float64 // how many standard deviations away from normal
	Severity   string  // "low", "medium", "high"
}

// Detector holds a rolling window of past values for a metric.
// This is how we establish "what's normal" without any external config.
type Detector struct {
	// map of metric name -> ring buffer of recent values
	history map[string][]float64
	// how many data points to keep per metric
	windowSize int
}

// NewDetector creates a Detector with a given window size.
// A window of 60 means we remember the last 60 readings to define "normal".
func NewDetector(windowSize int) *Detector {
	return &Detector{
		history:    make(map[string][]float64),
		windowSize: windowSize,
	}
}

// Observe records a new value for a metric. Call this every poll cycle.
func (d *Detector) Observe(metricName string, value float64) {
	buf := d.history[metricName]
	buf = append(buf, value)

	// Keep only the last windowSize values — sliding window
	if len(buf) > d.windowSize {
		buf = buf[len(buf)-d.windowSize:]
	}
	d.history[metricName] = buf
}

// Check returns an Anomaly if the latest value is unusual, or nil if normal.
// We need at least 10 data points before we can say anything meaningful.
func (d *Detector) Check(metricName string, currentValue float64) *Anomaly {
	buf := d.history[metricName]
	if len(buf) < 10 {
		return nil // not enough history yet
	}

	mean, stddev := stats(buf)

	// Avoid division by zero — if stddev is tiny, the metric is flat (stable)
	if stddev < 0.0001 {
		return nil
	}

	// Z-score: how many standard deviations is this value from the mean?
	// z = (value - mean) / stddev
	zScore := math.Abs((currentValue - mean) / stddev)

	// Only flag if it's more than 2.5 standard deviations away
	if zScore < 0.5 {
		return nil
	}

	return &Anomaly{
		MetricName: metricName,
		Value:      currentValue,
		ZScore:     zScore,
		Severity:   severity(zScore),
	}
}

// severity maps a z-score to a human-readable level
func severity(z float64) string {
	switch {
	case z >= 4.0:
		return "high"
	case z >= 3.0:
		return "medium"
	default:
		return "low"
	}
}

// stats computes mean and standard deviation of a slice of float64.
// This is the math behind z-score detection — nothing exotic.
func stats(values []float64) (mean, stddev float64) {
	n := float64(len(values))

	// compute mean
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean = sum / n

	// compute standard deviation
	variance := 0.0
	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}
	stddev = math.Sqrt(variance / n)

	return mean, stddev
}
