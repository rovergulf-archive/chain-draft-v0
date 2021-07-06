package traceutil

import (
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/spf13/viper"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-lib/metrics/prometheus"
	"io"
)

const JaegerTraceConfigKey = "jaeger_trace"

var ErrCollectorUrlNotSpecified = fmt.Errorf("jaeger collector url not specified")

type Tracer interface {
	opentracing.Tracer
	Close()
	CollectorUrl() string
}

type jaegerTracer struct {
	opentracing.Tracer
	io.Closer

	collectorUrl string
}

func (t *jaegerTracer) Close() {
	t.Closer.Close()
}

// CollectorUrl returns current Jaeger Collector agent url
func (t *jaegerTracer) CollectorUrl() string {
	return t.collectorUrl
}

func NewTracerFromViperConfig() (Tracer, error) {
	jaegerAddr := viper.GetString(JaegerTraceConfigKey)
	fmt.Println("jaegerAddr", jaegerAddr, len(jaegerAddr) > 0)
	if len(jaegerAddr) > 0 {
		return NewTracer(jaegerAddr)
	} else {
		return nil, ErrCollectorUrlNotSpecified
	}
}

func NewTracer(address string) (Tracer, error) {
	metrics := prometheus.New()

	traceTransport, err := jaeger.NewUDPTransport(address, 0)
	if err != nil {
		return nil, err
	}

	tracer, closer, err := config.Configuration{
		ServiceName: "rbn",
	}.NewTracer(
		config.Sampler(jaeger.NewConstSampler(true)),
		config.Reporter(jaeger.NewRemoteReporter(
			traceTransport,
			jaeger.ReporterOptions.Logger(jaeger.StdLogger)),
		),
		config.Metrics(metrics),
	)
	if err != nil {
		return nil, err
	}

	return &jaegerTracer{
		Tracer:       tracer,
		Closer:       closer,
		collectorUrl: address,
	}, nil
}
