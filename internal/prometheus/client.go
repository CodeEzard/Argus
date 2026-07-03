package prometheus

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type QueryResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string   `json:"resultType"`
		Result     []Result `json:"result"`
	} `json:"data"`
}

type Result struct {
	Metric map[string]string `json:"metric"`
	Value  []interface{}     `json:"value"`
}

// MetricSample is our clean internal representation
type MetricSample struct {
	Labels    map[string]string
	Value     float64
	Timestamp time.Time
}

func (c *Client) Query(promql string) ([]MetricSample, error) {
	endpoint := fmt.Sprintf("%s/api/v1/query", c.BaseURL)
	params := url.Values{}
	params.Set("query", promql)
	fullURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	resp, err := c.HTTPClient.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("failed to reach Prometheus at %s: %w", c.BaseURL, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var qr QueryResponse
	if err := json.Unmarshal(body, &qr); err != nil {
		return nil, fmt.Errorf("failed to parse Prometheus response: %w", err)
	}

	if qr.Status != "success" {
		return nil, fmt.Errorf("prometheus returned non-success status: %s", qr.Status)
	}

	samples := make([]MetricSample, 0, len(qr.Data.Result))
	for _, r := range qr.Data.Result {
		sample, err := parseResult(r)
		if err != nil {
			continue
		}
		samples = append(samples, sample)
	}

	return samples, nil
}

func parseResult(r Result) (MetricSample, error) {
	if len(r.Value) != 2 {
		return MetricSample{}, fmt.Errorf("unexpected value length: %d", len(r.Value))
	}

	tsFloat, ok := r.Value[0].(float64)
	if !ok {
		return MetricSample{}, fmt.Errorf("timestamp is not a float64")
	}

	valStr, ok := r.Value[1].(string)
	if !ok {
		return MetricSample{}, fmt.Errorf("value is not a string")
	}

	val, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return MetricSample{}, fmt.Errorf("failed to parse value %q: %w", valStr, err)
	}

	return MetricSample{
		Labels:    r.Metric,
		Value:     val,
		Timestamp: time.Unix(int64(tsFloat), 0),
	}, nil
}