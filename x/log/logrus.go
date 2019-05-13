package log

import (
	"github.com/sirupsen/logrus"
	"log"
	"sync"
)

type FieldLogger struct {
	logger logrus.FieldLogger
	mu     sync.Mutex
}

func New(l logrus.FieldLogger) *FieldLogger {
	return &FieldLogger{
		logger: l,
	}
}

func (b *FieldLogger) Write(p []byte) (n int, err error) {
	b.GetLogger().Printf("%s", string(p))
	return len(p), nil
}
func (b *FieldLogger) GetLogger() logrus.FieldLogger {
	if b == nil {
		return logrus.StandardLogger()
	}
	if b.logger != nil {
		return b.logger
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.logger == nil {
		b.logger = logrus.StandardLogger()
	}
	return b.logger
}
func (b *FieldLogger) SetStdLogger(l *log.Logger) {
	if l == nil {
		return
	}
	logger := logrus.New()
	logger.Out = l.Writer()
	b.SetLogger(logger)
}
func (b *FieldLogger) SetLogger(l logrus.FieldLogger) {
	if b == nil {
		return
	}
	if l == nil {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	b.logger = l
}