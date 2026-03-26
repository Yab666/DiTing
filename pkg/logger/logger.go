package logger

// Logger 定义了统一的日志接口，方便后续替换为 Zap 或 Logrus。
type Logger interface {
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
}
