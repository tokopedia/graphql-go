package log

import (
	"context"
	"fmt"
	"log"
	"runtime"
)

// Logger is the interface used to log panics that occur during query execution. It is settable via graphql.ParseSchema
type Logger interface {
	LogPanic(ctx context.Context, value interface{}, info string)
}

// DefaultLogger is the default logger used to log panics that occur during query execution
type DefaultLogger struct{}

// LogPanic is used to log recovered panic values that occur during query execution
func (l *DefaultLogger) LogPanic(_ context.Context, value interface{}, info string) {
	const size = 64 << 10
	buf := make([]byte, size)
	buf = buf[:runtime.Stack(buf, false)]
		log.Printf("graphql: panic occurred: %v\n%s", value, buf)
}

// CustomLogger is the custom logger for use with gqlserver custome config to log panics that occur during query execution
type CustomLogger struct{
	VerbosePanicLog bool
}

// LogPanic is used to log recovered panic values that occur during query execution
func (l *CustomLogger) LogPanic(_ context.Context, value interface{}, info string) {
	const size = 64 << 10
	buf := make([]byte, size)
	buf = buf[:runtime.Stack(buf, false)]
	msg := fmt.Sprintf("graphql: panic occurred: %v\n%s", value, buf)
	if l.VerbosePanicLog {
		msg = fmt.Sprintf("%s\n\n%s", msg, info)
	}

	log.Printf(msg)
}

