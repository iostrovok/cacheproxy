package logger

import (
	"github.com/iostrovok/cacheproxy/plugins"
)

type Logger struct {
}

func New() plugins.ILogger {
	return &Logger{}
}

func (p *Logger) Printf(_ string, v ...interface{}) {}
