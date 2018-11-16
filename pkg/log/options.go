package log

import (
	"io"
	"unsafe"

	"github.com/sirupsen/logrus"
)

type key uint16

const (
	keyLogger key = iota
)

// Cfg creates a Logger using the logrus library
type Cfg struct {
	Hooks     logrus.LevelHooks
	Formatter logrus.Formatter
	Level     logrus.Level
	Out       io.Writer
}

func (cfg Cfg) mkLogger() Logger {
	if cfg.Level == 0 {
		return NoOp()
	}

	l := logrus.New()
	if cfg.Hooks != nil {
		l.Hooks = cfg.Hooks
	}
	if cfg.Level != 0 {
		l.SetLevel(cfg.Level)
	}
	if cfg.Formatter != nil {
		l.Formatter = cfg.Formatter
	}
	if cfg.Out != nil {
		l.Out = cfg.Out
	}
	return (*entry)(unsafe.Pointer(logrus.NewEntry(l)))
}

// Option for Logger
type Option func(*Cfg) Option
