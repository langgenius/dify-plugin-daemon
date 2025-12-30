package debugging_runtime

import (
	"fmt"

	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
)

// GnetLogger is a logger implementation for gnet that uses the custom log package
//
// It implements the gnet.logging.Logger interface
// See: https://pkg.go.dev/github.com/panjf2000/gnet/v2#WithLogger
//
// Gnet uses format strings for logging, but our custom log package uses structured logging
// reformat it using fmt.Sprintf here as a bridge
type GnetLogger struct{}

func (l GnetLogger) Printf(format string, args ...any) string {
	return fmt.Sprintf(format, args...)
}

func (l GnetLogger) Debugf(format string, args ...any) {
	log.Debug(l.Printf(format, args...))
}

func (l GnetLogger) Infof(format string, args ...any) {
	log.Info(l.Printf(format, args...))
}

func (l GnetLogger) Warnf(format string, args ...any) {
	log.Warn(l.Printf(format, args...))
}

func (l GnetLogger) Errorf(format string, args ...any) {
	log.Error(l.Printf(format, args...))
}

func (l GnetLogger) Fatalf(format string, args ...any) {
	log.Error(l.Printf(format, args...))
}
