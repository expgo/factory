package factory

import (
	"fmt"
	"sync"
)

type Logger interface {
	Debugf(template string, args ...any)
}

type logger struct {
	mu   sync.Mutex
	hook func(msg string)
}

func (l *logger) Debugf(template string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	msg := fmt.Sprintf(template, args...)
	fmt.Println(msg)
	if l.hook != nil {
		l.hook(msg)
	}
}
