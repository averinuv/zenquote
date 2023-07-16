package logger

import (
	"fmt"
	"zenquote/internal/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger = *zap.Logger

func New(config config.Config) (Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.DisableCaller = true
	cfg.Sampling.Initial = 50
	cfg.Sampling.Thereafter = 50
	cfg.Encoding = config.Logger.Encoding
	cfg.OutputPaths = []string{"stderr"}
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.DisableStacktrace = true

	var lvl zapcore.Level
	if err := lvl.UnmarshalText([]byte(config.Logger.Level)); err != nil {
		return nil, fmt.Errorf("unmarshal config logger level failed: %w", err)
	}

	cfg.Level.SetLevel(lvl)

	logger, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("cfg build failed: %w", err)
	}

	logger.With(
		zap.Strings("tags", config.Logger.Tags),
	)

	return logger, nil
}
