package base

// HistBucket is a single bucket in a histogram
type HistBucket struct {
	Lower float64 `json:"lower" yaml:"lower"`
	Upper float64 `json:"upper" yaml:"upper"`
	Count uint64  `json:"count" yaml:"count"`
}

// MetricType identifies the type of the metric.
type MetricType string

// AggregationType identifies the type of the metric aggregator.
type AggregationType string

const (
	// CounterMetricType corresponds to Prometheus Counter metric type
	CounterMetricType MetricType = "Counter"
	// GaugeMetricType corresponds to Prometheus Gauge metric type
	GaugeMetricType MetricType = "Gauge"
	// HistogramMetricType corresponds to a Histogram metric type
	HistogramMetricType MetricType = "Histogram"
	// SampleMetricType corresponds to a Sample metric type
	SampleMetricType MetricType = "Sample"

	// decimalRegex is the regex used to identify percentiles
	decimalRegex = `^([\d]+(\.[\d]*)?|\.[\d]+)$`

	// CountAggregator corresponds to aggregation of type count
	CountAggregator AggregationType = "count"
	// MeanAggregator corresponds to aggregation of type mean
	MeanAggregator AggregationType = "mean"
	// StddevAggregator corresponds to aggregation of type stddev
	StdDevAggregator AggregationType = "stddev"
	// MinAggregator corresponds to aggregation of type min
	MinAggregator AggregationType = "min"
	// MaxAggregator corresponds to aggregation of type max
	MaxAggregator AggregationType = "max"
	// PercentileAggregator corresponds to aggregation of type max
	PercentileAggregator AggregationType = "percentile"
	// PercentileAggregatorPrefix corresponds to prefix for percentiles
	PercentileAggregatorPrefix = "p"
)

// // AuthType identifies the type of authentication used in the HTTP request
// type AuthType string

// const (
// 	// BasicAuthType corresponds to authentication with basic auth
// 	BasicAuthType AuthType = "Basic"

// 	// BearerAuthType corresponds to authentication with bearer token
// 	BearerAuthType AuthType = "Bearer"

// 	// APIKeyAuthType corresponds to authentication with API keys
// 	APIKeyAuthType AuthType = "APIKey"
// )

// // MethodType identifies the HTTP request method (aka verb) used in the HTTP request
// type MethodType string

// const (
// 	// GETMethodType corresponds to HTTP GET method
// 	GETMethodType MethodType = "GET"

// 	// POSTMethodType corresponds to HTTP POST method
// 	POSTMethodType MethodType = "POST"
// )

// // BackendInfo is map of backends
// type BackendInfo map[string]Backend

// // Backend describes how to query the backend and the list of metrics supported by the backend
// type Backend struct {
// 	// Description of the backend
// 	Description *string `json:"description,omitempty" yaml:"description,omitempty"`

// 	// BackendRequest provides detailed information about how to query the backend
// 	BackendRequest `json:",inline" yaml:",inline"`

// 	// Metrics is a list of metrics available from this backend
// 	Metrics []Metric `json:"metrics,omitempty" yaml:"metrics,omitempty"`
// }

// // BackendRequests provides details about how to query a backend
// type BackendRequest struct {
// 	// AuthType is the type of authentication used in the HTTP request
// 	AuthType *AuthType `json:"authType,omitempty" yaml:"authType,omitempty"`

// 	// Method is the HTTP method used in the HTTP request
// 	Method *MethodType `json:"method,omitempty" yaml:"method,omitempty"`

// 	// Provider identifies the type of metric database. Used for informational purposes.
// 	Provider *string `json:"provider,omitempty" yaml:"provider,omitempty"`

// 	// JQExpression defines the jq expression used by Iter8 to extract the metric value from the (JSON) response returned by the HTTP URL queried by Iter8.
// 	// An empty string is a valid jq expression.
// 	JQExpression *string `json:"jqExpression,omitempty" yaml:"jqExpression,omitempty"`

// 	// Secret is a reference to the Kubernetes secret.
// 	// Secret contains data used for HTTP authentication.
// 	// Secret may also contain data used for placeholder substitution in HeaderTemplates.
// 	Secret *string `json:"secret,omitempty" yaml:"secret,omitempty"`

// 	// Headers are key/value pairs corresponding to HTTP request headers and their values.
// 	// Value may be templated, in which Iter8 will attempt to substitute placeholders in the template at query time using Secret.
// 	// Placeholder substitution will be attempted only when Secret != nil.
// 	Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`

// 	// URL is HTTP URL of the metrics backend
// 	URL *string `json:"url,omitempty" yaml:"url,omitempty"`
// }

// // Metric object provides info about a specific metric provided by the backend
// type Metric struct {
// 	// Name is the name of the metric
// 	Name string `json:"name" yaml:"name"`

// 	// Text description of the metric
// 	Description *string `json:"description,omitempty" yaml:"description,omitempty"`

// 	// Params are key/value pairs corresponding to HTTP request parameters
// 	// Value may contain place holders, to be substituted by Iter8 with version-specific values.
// 	Params map[string]string `json:"params,omitempty" yaml:"params,omitempty"`

// 	// Units of the metric. Used for informational purposes.
// 	Units *string `json:"units,omitempty" yaml:"units,omitempty"`

// 	// Type of the metric
// 	Type *MetricType `json:"type,omitempty" yaml:"type,omitempty"`

// 	// Body is the string used to construct the (json) body of the HTTP request
// 	// Body may contain place holders, to be substituted by Iter8 with version-specific values.
// 	Body *string `json:"body,omitempty" yaml:"body,omitempty"`
// }
