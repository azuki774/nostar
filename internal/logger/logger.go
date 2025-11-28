package logger

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Load() *zap.Logger {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = JSTTimeEncoder
	config.Level = zap.NewAtomicLevelAt(zap.DebugLevel) // FOR DEBUG
	l, _ := config.Build()

	l.WithOptions(zap.AddStacktrace(zap.ErrorLevel))
	zap.ReplaceGlobals(l)
	return l
}

func JSTTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	const layout = "2006-01-02T15:04:05+09:00"
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	enc.AppendString(t.In(jst).Format(layout))
}
