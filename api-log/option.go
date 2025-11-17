package log

import (
	"io"
	"os"

	"github.com/rs/zerolog"
)

// request id not to be created, instead it should be used from the header
// log all the path and query params
// log sercvice name, function name
// save people from explicitly calling getctxlogger instead let the logging library get it from gin context in each of .info .debug .error .warn .fatal

type options struct {
	ServiceName    string
	Level          zerolog.Level
	OutputWriter   io.Writer
	Version        string
	SamplingConfig *SamplingConfig
}

func defaultLoggerOptions() options {
	return options{
		ServiceName:    "dummy",
		Level:          zerolog.InfoLevel,
		OutputWriter:   os.Stdout,
		Version:        "",
		SamplingConfig: nil, // No sampling by default
	}
}

type loggerOption func(o *options)

func WithServiceName(n string) loggerOption {
	return func(o *options) {
		o.ServiceName = n
	}
}

func WithLevel(l zerolog.Level) loggerOption {
	return func(o *options) {
		o.Level = l
	}
}

func WithOutputWriter(w io.Writer) loggerOption {
	return func(o *options) {
		o.OutputWriter = w
	}
}

func WithVersion(n string) loggerOption {
	return func(o *options) {
		o.Version = n
	}
}

// WithSampling configures log sampling to reduce log volume.
// Pass nil to disable sampling (default behavior).
//
// Example:
//   samplingConfig := &log.SamplingConfig{
//       LevelRates: map[zerolog.Level]float64{
//           zerolog.DebugLevel: 0.1,  // Sample 10% of debug logs
//       },
//   }
//   factory.Create(log.WithSampling(samplingConfig))
func WithSampling(config *SamplingConfig) loggerOption {
	return func(o *options) {
		o.SamplingConfig = config
	}
}
