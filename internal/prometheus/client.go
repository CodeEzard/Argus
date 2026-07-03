package prometheus

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

// Client holds config for talking to a Prometheus instance.
// Structs in Go = named collection of fields, like an object in other languages.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient constructs a Client. Convention in Go: constructors are named NewXxx.
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{},
	}
}

// These structs mirror the JSON shape Prometheus returns.
// `json:"field"` tells the decoder which JSON key maps to which struct field.
type queryResponse struct {
	Status string    `json:"status"`
	Data   queryData `json:"data"`
}

type queryData struct {
	ResultType string        `json:"resultType"`
	Result     []queryResult `json:"result"`
}

type queryResult struct {
	Metric map[string]string `json:"metric"` // labels e.g. {"job": "prometheus"}
	Value  [2]interface{}    `json:"value"`  // [timestamp, "value_as_string"]
}

// MetricResult is the clean type we expose outside this package.
type MetricResult struct {
	Labels map[string]string
	Value  float64
}

// Query runs a PromQL instant query and returns parsed results.
// Notice the Go error pattern: always return (value, error). Always check error.
func (c *Client) Query(promql string) ([]MetricResult, error) {
	// Build the URL with the query as a query param
	endpoint := fmt.Sprintf("%s/api/v1/query", c.BaseURL)
	params := url.Values{}
	params.Set("query", promql)
	fullURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	// Make the HTTP GET request
	resp, err := c.HTTPClient.Get(fullURL)
	if err != nil {
		// Wrap the error with context — makes debugging much easier
		return nil, fmt.Errorf("failed to reach Prometheus at %s: %w", c.BaseURL, err)
	}
	defer resp.Body.Close() // always close the body — this is Go idiom

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var parsed queryResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse Prometheus response: %w", err)
	}

	if parsed.Status != "success" {
		return nil, fmt.Errorf("Prometheus returned non-success status: %s", parsed.Status)
	}

	// Convert raw results into our clean MetricResult type
	results := make([]MetricResult, 0, len(parsed.Data.Result))
	for _, r := range parsed.Data.Result {
		// Value[1] is the metric value as a string — parse it to float64
		valStr, ok := r.Value[1].(string)
		if !ok {
			continue
		}
		val, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			continue
		}
		results = append(results, MetricResult{
			Labels: r.Metric,
			Value:  val,
		})
	}

	return results, nil
}
