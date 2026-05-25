package logging

import (
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Init configures zerolog with a console writer for stderr and a rotating
// JSON file in logDir. Callers can adjust level at runtime via Logger.Level().
func Init(level zerolog.Level, logDir string) (zerolog.Logger, error) {
	if err := os.MkdirAll(logDir, 0o700); err != nil {
		return zerolog.Nop(), err
	}
	rotator := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, "pidisk.log"),
		MaxSize:    20,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   true,
	}
	console := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
	writer := zerolog.MultiLevelWriter(console, rotator)
	logger := zerolog.New(writer).
		Level(level).
		With().
		Timestamp().
		Caller().
		Str("app", "pidisk").
		Logger()
	return logger, nil
}

func ParseLevel(s string) zerolog.Level {
	lvl, err := zerolog.ParseLevel(s)
	if err != nil || lvl == zerolog.NoLevel {
		return zerolog.InfoLevel
	}
	return lvl
}
