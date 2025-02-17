package logger

import (
	"io"
	"os"
	"time"

	"github.com/kiegroup/kogito-cloud-operator/pkg/util"

	"github.com/go-logr/logr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var (
	defaultOutput = os.Stderr
)

// Logger shared logger struct
type Logger struct {
	Logger        logr.Logger
	SugaredLogger *zap.SugaredLogger
}

// Opts describe logger options
type Opts struct {
	// Verbose will increase logging
	Verbose bool
	// Output specifies where to log
	Output io.Writer
	// Console logging doesn't display level nor timestamp and should be readable by humans
	Console bool
}

// GetLoggerWithOptions returns a custom named logger with given options
func GetLoggerWithOptions(name string, options *Opts) *zap.SugaredLogger {
	if options == nil {
		options = getDefaultOpts()
	} else if options.Output == nil {
		options.Output = defaultOutput
	}
	return getLogger(name, options)
}

// GetLogger returns a custom named logger
func GetLogger(name string) *zap.SugaredLogger {
	options := getDefaultOpts()
	return getLogger(name, options)
}

func getDefaultOpts() *Opts {
	return &Opts{
		Verbose: util.GetBoolEnv("DEBUG"),
		Output:  defaultOutput,
		Console: false,
	}
}

func getLogger(name string, options *Opts) *zap.SugaredLogger {
	// Set log level... override default w/ command-line variable if set.
	// The logger instantiated here can be changed to any logger
	// implementing the logr.Logger interface. This logger will
	// be propagated through the whole operator, generating
	// uniform and structured logs.
	logger := createLogger(options)
	logger.Logger = logf.Log.WithName(name)
	return logger.SugaredLogger.Named(name)
}

func createLogger(options *Opts) (logger Logger) {
	log := Logger{
		Logger:        logf.ZapLogger(options.Verbose),
		SugaredLogger: zapSugaredLogger(options),
	}
	defer log.SugaredLogger.Sync()

	logf.SetLogger(log.Logger)
	return log
}

// zapSugaredLogger is a Logger implementation.
// If development is true, a Zap development config will be used,
// otherwise a Zap production config will be used
// (stacktraces on errors, sampling).
func zapSugaredLogger(options *Opts) *zap.SugaredLogger {
	return zapSugaredLoggerTo(options)
}

// zapSugaredLoggerTo returns a new Logger implementation using Zap which logs
// to the given destination, instead of stderr.  It otherise behaves like
// ZapLogger.
func zapSugaredLoggerTo(options *Opts) *zap.SugaredLogger {
	// this basically mimics New<type>Config, but with a custom sink
	sink := zapcore.AddSync(options.Output)

	var enc zapcore.Encoder
	var lvl zap.AtomicLevel
	var opts []zap.Option

	if options.Console {
		encCfg := zap.NewDevelopmentEncoderConfig()
		encCfg.LevelKey = ""
		encCfg.TimeKey = ""
		encCfg.NameKey = ""
		encCfg.CallerKey = ""
		enc = zapcore.NewConsoleEncoder(encCfg)
		if options.Verbose {
			lvl = zap.NewAtomicLevelAt(zap.DebugLevel)
		} else {
			lvl = zap.NewAtomicLevelAt(zap.InfoLevel)
		}
		opts = append(opts, zap.Development(), zap.AddStacktrace(zap.ErrorLevel))
	} else {
		if options.Verbose {
			encCfg := zap.NewDevelopmentEncoderConfig()
			enc = zapcore.NewConsoleEncoder(encCfg)
			lvl = zap.NewAtomicLevelAt(zap.DebugLevel)
			opts = append(opts, zap.Development(), zap.AddStacktrace(zap.ErrorLevel))
		} else {
			encCfg := zap.NewProductionEncoderConfig()
			enc = zapcore.NewJSONEncoder(encCfg)
			lvl = zap.NewAtomicLevelAt(zap.InfoLevel)
			opts = append(opts, zap.WrapCore(func(core zapcore.Core) zapcore.Core {
				return zapcore.NewSampler(core, time.Second, 100, 100)
			}))
		}
	}

	opts = append(opts, zap.AddCallerSkip(1), zap.ErrorOutput(sink))
	log := zap.New(zapcore.NewCore(&logf.KubeAwareEncoder{Encoder: enc, Verbose: options.Verbose}, sink, lvl))
	log = log.WithOptions(opts...)

	return log.Sugar()
}
