package logger

import (
	"context"
	"log/slog"
	"os"
)

const levelFatal = slog.Level(12)

var levelNames = map[slog.Leveler]string{
	levelFatal: "FATAL",
}

var logger *slog.Logger

func InitLogger() error {
	fileName := "data.log"
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	opts := slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.LevelKey {
				level := a.Value.Any().(slog.Level)
				levelLabel, exists := levelNames[level]
				if !exists {
					levelLabel = level.String()
				}

				a.Value = slog.StringValue(levelLabel)
			}
			return a
		},
	}

	logger = slog.New(slog.NewJSONHandler(file, &opts))
	return nil
}

func Info(msg, methodPtr string, fields ...any) {
	fields = append(fields, "method_pointer", methodPtr)
	logger.Info(msg, fields...)

}

func Debug(msg, methodPtr string, fields ...any) {
	fields = append(fields, "method_pointer", methodPtr)
	logger.Debug(msg, fields...)
}

func Warn(msg, methodPtr string, fields ...any) {
	fields = append(fields, "method_pointer", methodPtr)
	logger.Warn(msg, fields...)
}

func Fatal(msg, methodPtr string, fields ...any) {
	fields = append(fields, "method_pointer", methodPtr)
	logger.Log(context.Background(), levelFatal, msg, fields...)
	os.Exit(1)
}

func Error(msg, methodPtr string, fields ...any) {
	fields = append(fields, "method_pointer", methodPtr)
	logger.Error(msg, fields...)
}
