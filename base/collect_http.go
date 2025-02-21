package base

import (
	"errors"
	"fmt"
	"io"
	"time"

	"fortio.org/fortio/fhttp"
	fortioLog "fortio.org/fortio/log"
	"fortio.org/fortio/periodic"
	"fortio.org/fortio/stats"
	log "github.com/iter8-tools/iter8/base/log"
)

// versionHTTP contains header and url information needed to send requests to each version.
type versionHTTP struct {
	// HTTP headers to use in the query for this version; optional
	Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
	// URL to use for querying this version
	URL string `json:"url" yaml:"url"`
}

// errorRange has lower and upper limits for HTTP status codes. HTTP status code within this range is considered an error
type errorRange struct {
	// Lower end of the range
	Lower *int `json:"lower" yaml:"lower"`
	// Upper end of the range
	Upper *int `json:"upper" yaml:"upper"`
}

// collectHTTPInputs contain the inputs to the metrics collection task to be executed.
type collectHTTPInputs struct {
	// NumRequests is the number of requests to be sent to each version. Default value is 100.
	NumRequests *int64 `json:"numRequests" yaml:"numRequests"`
	// Duration of this task. Specified in the Go duration string format (example, 5s). If both duration and numQueries are specified, then duration is ignored.
	Duration *string `json:"duration" yaml:"duration"`
	// QPS is the number of requests per second sent to each version. Default value is 8.0.
	QPS *float32 `json:"qps" yaml:"qps"`
	// Connections is the number of number of parallel connections used to send load. Default value is 4.
	Connections *int `json:"connections" yaml:"connections"`
	// PayloadStr is the string data to be sent as payload. If this field is specified, Iter8 will send HTTP POST requests to versions using this string as the payload.
	PayloadStr *string `json:"payloadStr" yaml:"payloadStr"`
	// PayloadURL is the URL of payload. If this field is specified, Iter8 will send HTTP POST requests to versions using data downloaded from this URL as the payload. If both `payloadStr` and `payloadURL` is specified, the former is ignored.
	PayloadURL *string `json:"payloadURL" yaml:"payloadURL"`
	// ContentType is the type of the payload. Indicated using the Content-Type HTTP header value. This is intended to be used in conjunction with one of the `payload*` fields above. If this field is specified, Iter8 will send HTTP POST requests to versions using this content type header value.
	ContentType *string `json:"contentType" yaml:"contentType"`
	// ErrorRanges is a list of errorRange values. Each range specifies an upper and/or lower limit on HTTP status codes. HTTP responses that fall within these error ranges are considered error. Default value is {{lower: 400},} - i.e., HTTP status codes >= 400 are considered as error.
	ErrorRanges []errorRange `json:"errorRanges" yaml:"errorRanges"`
	// Percentiles are the latency percentiles collected by this task. Percentile values have a single digit precision (i.e., rounded to one decimal place). Default value is {50.0, 75.0, 90.0, 95.0, 99.0, 99.9,}.
	Percentiles []float64 `json:"percentiles" yaml:"percentiles"`
	// VersionInfo is a non-empty list of version values.
	VersionInfo []*versionHTTP `json:"versionInfo" yaml:"versionInfo"`
	// Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
}

const (
	// CollectHTTPTaskName is the name of this task which performs load generation and metrics collection.
	CollectHTTPTaskName                = "gen-load-and-collect-metrics-http"
	defaultQPS                         = float32(8)
	defaultHTTPNumRequests             = int64(100)
	defaultHTTPConnections             = 4
	httpMetricPrefix                   = "http"
	builtInHTTPRequestCountId          = "request-count"
	builtInHTTPErrorCountId            = "error-count"
	builtInHTTPErrorRateId             = "error-rate"
	builtInHTTPLatencyMeanId           = "latency-mean"
	builtInHTTPLatencyStdDevId         = "latency-stddev"
	builtInHTTPLatencyMinId            = "latency-min"
	builtInHTTPLatencyMaxId            = "latency-max"
	builtInHTTPLatencyHistId           = "latency"
	builtInHTTPLatencyPercentilePrefix = "latency-p"
	builtInActualQPSId 								 = "actual-qps"
	builtInActualDurationId 					 = "actual-duration"
)

var (
	defaultErrorRanges = []errorRange{{Lower: intPointer(400)}}
	defaultPercentiles = [...]float64{50.0, 75.0, 90.0, 95.0, 99.0, 99.9}
)

// errorCode checks if a given code is an error code
func (t *collectHTTPTask) errorCode(code int) bool {
	for _, lims := range t.With.ErrorRanges {
		// if no lower limit (check upper)
		if lims.Lower == nil && code <= *lims.Upper {
			return true
		}
		// if no upper limit (check lower)
		if lims.Upper == nil && code >= *lims.Lower {
			return true
		}
		// if both limits are present (check both)
		if lims.Upper != nil && lims.Lower != nil && code <= *lims.Upper && code >= *lims.Lower {
			return true
		}
	}
	return false
}

// collectHTTPTask enables load testing of HTTP services.
type collectHTTPTask struct {
	TaskMeta
	With collectHTTPInputs `json:"with" yaml:"with"`
}

// initializeDefaults sets default values for the collect task
func (t *collectHTTPTask) initializeDefaults() {
	if t.With.NumRequests == nil && t.With.Duration == nil {
		t.With.NumRequests = int64Pointer(defaultHTTPNumRequests)
	}
	if t.With.QPS == nil {
		t.With.QPS = float32Pointer(defaultQPS)
	}
	if t.With.Connections == nil {
		t.With.Connections = intPointer(defaultHTTPConnections)
	}
	if t.With.ErrorRanges == nil {
		t.With.ErrorRanges = defaultErrorRanges
	}
	// default percentiles are always collected
	// if other percentiles are specified, they are collected as well
	for _, p := range defaultPercentiles {
		t.With.Percentiles = append(t.With.Percentiles, p)
	}
	tmp := Uniq(t.With.Percentiles)
	t.With.Percentiles = []float64{}
	for _, val := range tmp {
		t.With.Percentiles = append(t.With.Percentiles, val.(float64))
	}
}

//validateInputs for this task
func (t *collectHTTPTask) validateInputs() error {
	return nil
}

// getFortioOptions constructs Fortio's HTTP runner options based on collect task inputs
func (t *collectHTTPTask) getFortioOptions(j int) (*fhttp.HTTPRunnerOptions, error) {
	fortioLog.SetOutput(io.Discard)
	// basic runner
	fo := &fhttp.HTTPRunnerOptions{
		RunnerOptions: periodic.RunnerOptions{
			RunType:     "Iter8 load test",
			QPS:         float64(*t.With.QPS),
			NumThreads:  *t.With.Connections,
			Percentiles: t.With.Percentiles,
			Out:         io.Discard,
		},
		HTTPOptions: fhttp.HTTPOptions{
			URL: t.With.VersionInfo[j].URL,
		},
	}

	//num requests
	if t.With.NumRequests != nil {
		fo.RunnerOptions.Exactly = *t.With.NumRequests
	}

	// add duration
	var duration time.Duration
	var err error
	if t.With.Duration != nil {
		duration, err = time.ParseDuration(*t.With.Duration)
		if err == nil {
			fo.RunnerOptions.Duration = duration
		} else {
			log.Logger.WithStackTrace(err.Error()).Error("unable to parse duration")
			return nil, err
		}
	}

	// content type & payload
	if t.With.ContentType != nil {
		fo.ContentType = *t.With.ContentType
	}
	if t.With.PayloadStr != nil {
		fo.Payload = []byte(*t.With.PayloadStr)
	}
	if t.With.PayloadURL != nil {
		b, err := getPayloadBytes(*t.With.PayloadURL)
		if err != nil {
			return nil, err
		} else {
			fo.Payload = b
		}
	}

	// headers
	log.Logger.Trace("attempting to add headers")
	for key, value := range t.With.VersionInfo[j].Headers {
		log.Logger.Trace(key + ":" + value)
		fo.AddAndValidateExtraHeader(key + ":" + value)
	}
	// for key, value := range t.With.Headers {
	// 	log.Logger.Trace(key + ":" + value)
	// 	fo.AddAndValidateExtraHeader(key + ":" + value)
	// }

	return fo, nil
}

// resultForVersion collects Fortio result for a given version
func (t *collectHTTPTask) resultForVersion(j int) (*fhttp.HTTPRunnerResults, error) {
	// the main idea is to run Fortio with proper options

	fo, err := t.getFortioOptions(j)
	if err != nil {
		return nil, err
	}
	log.Logger.Trace("got fortio options")
	log.Logger.Trace("URL: ", fo.URL)
	ifr, err := fhttp.RunHTTPTest(fo)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("fortio failed")
		if ifr == nil {
			log.Logger.Error("failed to get results since fortio run was aborted")
		}
	}
	log.Logger.Trace("ran fortio http test")
	return ifr, err
}

// Run executes this task
func (t *collectHTTPTask) Run(exp *Experiment) error {
	err := t.validateInputs()
	if err != nil {
		return err
	}

	t.initializeDefaults()

	if len(t.With.VersionInfo) == 0 {
		log.Logger.Error("collect task must specify info for at least one version")
		return errors.New("collect task must specify info for at least one version")
	}

	fm := make([]*fhttp.HTTPRunnerResults, len(t.With.VersionInfo))

	// run fortio queries for each version sequentially
	for j := range t.With.VersionInfo {
		log.Logger.Trace("initiating fortio for version ", j)
		var data *fhttp.HTTPRunnerResults
		var err error
		if t.With.VersionInfo[j] == nil {
			data = nil
		} else {
			data, err = t.resultForVersion(j)
			if err == nil {
				fm[j] = data
			} else {
				return err
			}
		}
	}

	// this task creates insights for different versions
	// initialize num versions
	err = exp.Result.initInsightsWithNumVersions(len(t.With.VersionInfo))
	if err != nil {
		return err
	}
	in := exp.Result.Insights

	for i := range t.With.VersionInfo {
		if fm[i] != nil {
			// ActualQPS
			m := httpMetricPrefix + "/" + builtInActualQPSId
			mm := MetricMeta{
				Description: "Actual Queries per Second",
				Type: GaugeMetricType,
			}
			in.updateMetric(m, mm, i, float64(fm[i].ActualQPS))

			// ActualDuration
			m = httpMetricPrefix + "/" + builtInActualDurationId
			mm = MetricMeta{
				Description: "Actual Duration of Experiment",
				Type:        GaugeMetricType,
				Units:       StringPointer("msec"),
			}
			in.updateMetric(m, mm, i, float64(fm[i].ActualDuration.Milliseconds()))


			// request count
			m = httpMetricPrefix + "/" + builtInHTTPRequestCountId
			mm = MetricMeta{
				Description: "number of requests sent",
				Type:        CounterMetricType,
			}
			in.updateMetric(m, mm, i, float64(fm[i].DurationHistogram.Count))

			// error count & rate
			val := float64(0)
			for code, count := range fm[i].RetCodes {
				if t.errorCode(code) {
					val += float64(count)
				}
			}
			// error count
			m = httpMetricPrefix + "/" + builtInHTTPErrorCountId
			mm = MetricMeta{
				Description: "number of responses that were errors",
				Type:        CounterMetricType,
			}
			in.updateMetric(m, mm, i, val)

			// error-rate
			m = httpMetricPrefix + "/" + builtInHTTPErrorRateId
			rc := float64(fm[i].DurationHistogram.Count)
			if rc != 0 {
				mm = MetricMeta{
					Description: "fraction of responses that were errors",
					Type:        GaugeMetricType,
				}
				in.updateMetric(m, mm, i, val/rc)
			}

			// mean-latency
			m = httpMetricPrefix + "/" + builtInHTTPLatencyMeanId
			mm = MetricMeta{
				Description: "mean of observed latency values",
				Type:        GaugeMetricType,
				Units:       StringPointer("msec"),
			}
			in.updateMetric(m, mm, i, 1000.0*fm[i].DurationHistogram.Avg)

			// stddev-latency
			m = httpMetricPrefix + "/" + builtInHTTPLatencyStdDevId
			mm = MetricMeta{
				Description: "standard deviation of observed latency values",
				Type:        GaugeMetricType,
				Units:       StringPointer("msec"),
			}
			in.updateMetric(m, mm, i, 1000.0*fm[i].DurationHistogram.StdDev)

			// min-latency
			m = httpMetricPrefix + "/" + builtInHTTPLatencyMinId
			mm = MetricMeta{
				Description: "minimum of observed latency values",
				Type:        GaugeMetricType,
				Units:       StringPointer("msec"),
			}
			in.updateMetric(m, mm, i, 1000.0*fm[i].DurationHistogram.Min)

			// max-latency
			m = httpMetricPrefix + "/" + builtInHTTPLatencyMaxId
			mm = MetricMeta{
				Description: "maximum of observed latency values",
				Type:        GaugeMetricType,
				Units:       StringPointer("msec"),
			}
			in.updateMetric(m, mm, i, 1000.0*fm[i].DurationHistogram.Max)

			// percentiles
			for _, p := range fm[i].DurationHistogram.Percentiles {
				m = fmt.Sprintf("%v/%v%v", httpMetricPrefix, builtInHTTPLatencyPercentilePrefix, p.Percentile)
				mm = MetricMeta{
					Description: fmt.Sprintf("%v-th percentile of observed latency values", p.Percentile),
					Type:        GaugeMetricType,
					Units:       StringPointer("msec"),
				}
				in.updateMetric(m, mm, i, 1000.0*p.Value)
			}

			// latency histogram
			m = httpMetricPrefix + "/" + builtInHTTPLatencyHistId
			mm = MetricMeta{
				Description: "Latency Histogram",
				Type:        HistogramMetricType,
				Units:       StringPointer("msec"),
			}
			lh := latencyHist(fm[i].DurationHistogram)
			in.updateMetric(m, mm, i, lh)
		}
	}
	return nil
}

// compute latency histogram by resampling
func latencyHist(hd *stats.HistogramData) []HistBucket {
	buckets := []HistBucket{}
	for _, v := range hd.Data {
		buckets = append(buckets, HistBucket{
			Lower: v.Start * 1000.0, // sec to msec
			Upper: v.End * 1000.0,
			Count: uint64(v.Count),
		})
	}
	return buckets
}
