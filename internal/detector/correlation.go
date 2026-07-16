package detector

import(
	"time"
)

// CorrelatedPair represents two metrics that anomalied close together
type CorrelatedPair struct {
    MetricA    string
    MetricB    string
    TimeDelta  time.Duration // how far apart they spiked
    Relation   string        // "likely_cause", "likely_symptom", "correlated"
}

func (d *Detector) FindCorrelated(triggerName string, triggerTime time.Time) []CorrelatedPair {
    var pairs []CorrelatedPair

    for name, timestamps := range d.timestamps {
        if name == triggerName {
            continue // skip the trigger itself
        }

        // check if this metric had a recent observation within 60s of trigger
        for _, ts := range timestamps {
            delta := triggerTime.Sub(ts)
            if delta < 0 {
                delta = -delta
            }
            if delta <= 15*time.Second {
                relation := "correlated"
                if ts.Before(triggerTime) {
                    relation = "likely_cause"
                } else {
                    relation = "likely_symptom"
                }
                pairs = append(pairs, CorrelatedPair{
                    MetricA:   triggerName,
                    MetricB:   name,
                    TimeDelta: delta,
                    Relation:  relation,
                })
                break // one match per metric is enough
            }
        }
    }

    return pairs
}

