package detector

import (
	"math"
)

type Signal struct {
    Kind    string
    Value   float64
}

type MultiAnomaly struct {
    MetricName string
    Value      float64
    Signals    []Signal
    Severity   string
}

func (d *Detector) checkTrend(name string) *Signal {
    buf := d.history[name]
    if len(buf) < 10 {
        return nil
    }

    slope := linearSlope(buf)

    // slope > 0.01 per observation means steady upward drift
    if slope > 0.01 {
        return &Signal{Kind: "trend", Value: slope}
    }
    return nil
}

func (d *Detector) checkRateOfChange(name string) *Signal {
    buf := d.history[name]
    if len(buf) < 6 {
        return nil
    }

    n := len(buf)
    recent := avg(buf[n-3 : n])
    previous := avg(buf[n-6 : n-3])

    if previous == 0 {
        return nil
    }

    changeRate := (recent - previous) / previous
    if math.Abs(changeRate) > 0.25 {
        return &Signal{Kind: "rate_of_change", Value: changeRate}
    }
    return nil
}

func (d *Detector) CheckMulti(name string, value float64) *MultiAnomaly {
    var signals []Signal

    if a := d.Check(name, value); a != nil {
        signals = append(signals, Signal{Kind: "zscore", Value: a.ZScore})
    }
    if s := d.checkTrend(name); s != nil {
        signals = append(signals, *s)
    }
    if s := d.checkRateOfChange(name); s != nil {
        signals = append(signals, *s)
    }

    if len(signals) == 0 {
        return nil
    }

    return &MultiAnomaly{
        MetricName: name,
        Value:      value,
        Signals:    signals,
        Severity:   deriveSeverity(signals),
    }
}

// avg returns the mean of a slice
func avg(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// linearSlope computes the slope of a best-fit line through the values
// hint: use the formula slope = (n*Σxy - Σx*Σy) / (n*Σx² - (Σx)²)
// where x is the index (0,1,2...) and y is the value
func linearSlope(values []float64) float64 {
	n := float64(len(values))
	if n < 2 {
		return 0
	}
	sumX := 0.0
	sumY := 0.0
	sumXY := 0.0
	sumX2 := 0.0
	for i, y := range values {
		x := float64(i)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}
	denom := n*sumX2 - sumX*sumX
	if denom == 0 {
		return 0
	}
	return (n*sumXY - sumX*sumY) / denom
}

// deriveSeverity picks severity based on how many signals fired
func deriveSeverity(signals []Signal) string {
    if len(signals) >= 3 {
        return "high"
    } else if len(signals) == 2 {
        return "medium"
    }
    return "low"
}