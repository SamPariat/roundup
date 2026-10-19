// Package logger constructs the zap logger used by every Go service.
// Call New once in main.go and inject the result into handlers and middleware.
package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New builds a console logger that tags every line with the service name and
// level, e.g. "12:00:00.000	[INFO]	[gateway]	starting gateway	{"port": "8080"}".
// Level must be a valid zapcore level string: debug, info, warn, or error.
// Panics at startup if the level is invalid.
func New(service, level string) *zap.Logger {
	lvl, err := zapcore.ParseLevel(level)
	if err != nil {
		panic(fmt.Sprintf("invalid log level %q: %v", level, err))
	}

	encoderCfg := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "service",
		MessageKey:     "msg",
		CallerKey:      "caller",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    bracketLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("15:04:05.000"),
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeName:     bracketNameEncoder,
	}

	core := zapcore.NewCore(zapcore.NewConsoleEncoder(encoderCfg), zapcore.AddSync(os.Stdout), lvl)

	return zap.New(core).Named(service)
}

// bracketLevelEncoder renders the log level as "[INFO]", "[ERROR]", etc.
func bracketLevelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString("[" + l.CapitalString() + "]")
}

// bracketNameEncoder renders the logger name (service tag) as "[gateway]", etc.
func bracketNameEncoder(name string, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString("[" + name + "]")
}
