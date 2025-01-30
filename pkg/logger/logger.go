package logger

import (
	"context"
	"fmt"
	stdLog "log"
	"os"
	"strings"
)

type Logger interface {
	Error(ctx context.Context, args ...interface{})
	Info(ctx context.Context, args ...interface{})
}

// TODO (KB): Use alternative logger
type log struct {
	l *stdLog.Logger
}

func (l *log) Error(ctx context.Context, args ...interface{}) {
	var result strings.Builder

	for _, arg := range append(args) {
		if _, err := result.WriteString(fmt.Sprintf(" %v ", arg)); err != nil {
			panic(err)
		}
	}

	l.l.Println(result.String())
}

func (l *log) Info(ctx context.Context, args ...interface{}) {
	l.Error(ctx, args...)
}

func New(prefix string) Logger {
	return &log{
		l: stdLog.New(os.Stdout, fmt.Sprintf("%s: ", prefix), stdLog.LstdFlags),
	}
}
