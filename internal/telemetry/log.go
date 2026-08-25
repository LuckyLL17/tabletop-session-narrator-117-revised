package telemetry

import (
	"log"
	"time"
)

type Logger struct{ base *log.Logger }

func NewLogger() *Logger                    { return &Logger{base: log.Default()} }
func (l *Logger) Event(kind, detail string) { l.base.Printf("event kind=%s detail=%s", kind, detail) }
func (l *Logger) Request(method, path string, started time.Time) {
	l.base.Printf("http method=%s path=%s duration=%s", method, path, time.Since(started))
}
func (l *Logger) Error(kind string, err error) {
	if err != nil {
		l.base.Printf("error kind=%s message=%s", kind, err)
	}
}
