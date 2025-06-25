package mirror

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
	"net/http"
	"net/http/httputil"
	"net/url"
)

var (
	apiCounter  metric.Int64Counter
	apiDuration metric.Int64Histogram
	meter       = otel.Meter("flexcity.energy/plc-connector-go/handler/reverseproxy")
)

func init() {
	var err error
	apiCounter, err = meter.Int64Counter(
		"api.counter",
		metric.WithDescription("Number of API calls."),
		metric.WithUnit("{call}"),
	)
	if err != nil {
		panic(err)
	}
	apiDuration, err = meter.Int64Histogram(
		"api.duration",
		metric.WithDescription("Duration of API calls."),
		metric.WithUnit("s"),
	)
	if err != nil {
		panic(err)
	}
}

func NewProxy(target *url.URL) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ModifyResponse = func(resp *http.Response) error {
		attributes := metric.WithAttributes(
			semconv.HTTPResponseStatusCode(resp.StatusCode),
			semconv.HTTPRequestMethodKey.String(resp.Request.Method),
			semconv.URLPath(resp.Request.URL.Path),
			semconv.URLDomain(resp.Request.URL.Host),
		)
		apiCounter.Add(resp.Request.Context(), 1, attributes)
		apiDuration.Record(resp.Request.Context(), 1, attributes)
		return nil
	}
	return proxy
}
