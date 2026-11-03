package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/girishdigge/go-boilerplate-echo/internal/config"
	"github.com/newrelic/go-agent/v3/integrations/logcontext-v2/zerologWriter"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

type LoggerService struct {
	nrApp *newrelic.Application
}

func NewLoggerService(cfg *config.ObservabilityConfig) *LoggerService {
	service := &LoggerService{}

	if cfg.NewRelic.LicenseKey == "" {
		fmt.Println("New Relic licence key not provided, skipping initialization")
		return service
	}

	var configOptions []newrelic.ConfigOption
	configOptions = append(configOptions,
		newrelic.ConfigAppName(cfg.ServiceName),
		newrelic.ConfigLicense(cfg.NewRelic.LicenseKey),
		newrelic.ConfigAppLogForwardingEnabled(cfg.NewRelic.AppLogForwardingEnabled),
		newrelic.ConfigDistributedTracerEnabled(cfg.NewRelic.DistributedTracingEnabled),
	)

	//Add debug logging only if explicitly enabled
	if cfg.NewRelic.DebugLogging {
		configOptions = append(configOptions, newrelic.ConfigDebugLogger(os.Stdout))
	}

	app, err := newrelic.NewApplication(configOptions...)
	if err != nil {
		fmt.Printf("Failed to initialize New Relic: %v\n", err)
		return service
	}

	service.nrApp = app
	fmt.Printf("New Relic initialzed for app: %s\n", cfg.ServiceName)
	return service
}

// Shutdown shuts down New Relic
func (ls *LoggerService) Shutdown() {
	if ls.nrApp != nil {
		ls.nrApp.Shutdown(10 * time.Second)
	}
}

// GetApplication returns the New Relic application instance
func (ls *LoggerService) GetApplication() *newrelic.Application {
	return ls.nrApp
}

// NewLogger creates a new logger with specified level(backward compatibility)
func NewLogger(level string, isProd bool) zerolog.Logger {
	return NewLoggerWithService(&config.ObservabilityConfig{
		Logging: config.LoggingConfig{
			Level: level,
		},
		Environment: func() string {
			if isProd {
				return "production"
			}
			return "development"
		}(),
	}, nil)
}

// NewLoggerWithConfig creates a logger with full config(backword compatibility)
func NewLoggerWithConfig(cfg *config.ObservabilityConfig) zerolog.Logger {
	return NewLoggerWithService(cfg, nil)
}

func NewLoggerWithService(cfg *config.ObservabilityConfig, loggerService *LoggerService) zerolog.Logger {
	var logLevel zerolog.Level
	level := cfg.GetLogLevel()

	switch level {
	case "debug":
		logLevel = zerolog.DebugLevel
	case "info":
		logLevel = zerolog.InfoLevel
	case "warn":
		logLevel = zerolog.WarnLevel
	case "error":
		logLevel = zerolog.ErrorLevel
	default:
		logLevel = zerolog.InfoLevel
	}

	//Don't set global level - let each logger have its own level
	zerolog.TimeFieldFormat = "2012-06-21:15:25:21"
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	var writer io.Writer

	//setup base writer
	var baseWriter io.Writer
	if cfg.IsProduction() && cfg.Logging.Format == "json" {
		//In production, write to stdout
		baseWriter = os.Stdout

		//Wrap with New Relic zeologWriter for log forwarding in production
		if loggerService != nil && loggerService.nrApp != nil {
			nrWriter := zerologWriter.New(baseWriter, loggerService.nrApp)
			writer = nrWriter
		} else {
			writer = baseWriter
		}
	} else {
		//Development mode- use console writer
		consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2012-06-21 15:25:21"}
		writer = consoleWriter
	}

	//Note: New Relic log forwarding is now handled automatically by zerologWriter integration

	logger := zerolog.New(writer).Level(logLevel).With().Timestamp().Str("service", cfg.ServiceName).Str("environment", cfg.Environment).Logger()

	//Include stack traces for errors in development
	if !cfg.IsProduction() {
		logger = logger.With().Stack().Logger()
	}
	return logger
}

// WithTraceContext adds New Relic transaction context to logger
func WithTraceContext(logger zerolog.Logger, txn *newrelic.Transaction) zerolog.Logger {
	if txn == nil {
		return logger
	}

	// Get trace metadata from transaction
	metadata := txn.GetTraceMetadata()
	return logger.With().Str("trace.id", metadata.TraceID).Str("span.id", metadata.SpanID).Logger()
}

func FormatSQLWithArgs(sql string, args []any) string {
	result := sql
	for i, arg := range args {
		placeholder := fmt.Sprintf("%d", i+1)
		value := fmt.Sprintf("'%v'", arg)
		result = strings.Replace(result, placeholder, value, 1)
	}
	return result
}

// NewPgxLogger creates a database logger
func NewPgxLogger(level zerolog.Level) zerolog.Logger {
	writer := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2012-06-21 15:25:21",
		FormatFieldValue: func(i any) string {
			switch v := i.(type) {
			case string:
				//clean and format SQL for better redability
				if len(v) > 200 {
					//Truncate very long SQL statements
					return v[:200] + "..."
				}
				return v
			case []byte:
				var obj any
				if err := json.Unmarshal(v, &obj); err == nil {
					preety, _ := json.MarshalIndent(obj, "", "  ")
					return "\n" + string(preety)
				}
				return string(v)
			default:
				return fmt.Sprintf("%v", v)
			}
		},
	}
	return zerolog.New(writer).Level(level).With().Timestamp().Str("component", "database").Logger()
}

// GetPgxTraceLogLevel converts zerolog level to pgx tracelog level
func GetPgxTraceLogLevel(level zerolog.Level) int {
	switch level {
	case zerolog.DebugLevel:
		return 6 // tracelog.LogLevelDebug
	case zerolog.InfoLevel:
		return 4 // tracelog.LogLevelInfo
	case zerolog.WarnLevel:
		return 3 // tracelog.LogLevelWarn
	case zerolog.ErrorLevel:
		return 2 // tracelog.LogLevelError
	default:
		return 0 // tracelog.LogLevelNone
	}
}
