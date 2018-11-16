package log

import (
	"context"
	"unsafe"

	"github.com/sirupsen/logrus"
)

const (
	locusLabel = "locus"
)

// F is a set of fields
type F map[string]interface{}

// Logger provides observability
type Logger interface {
	Debug(...interface{})
	Debugf(string, ...interface{})
	Debugln(...interface{})

	Info(...interface{})
	Infof(string, ...interface{})
	Infoln(...interface{})

	Warn(...interface{})
	Warnf(string, ...interface{})
	Warnln(...interface{})

	Error(...interface{})
	Errorf(string, ...interface{})
	Errorln(...interface{})

	WithLocus(string) Logger
	WithError(error) Logger
	WithField(string, interface{}) Logger
	WithFields(F) Logger
}

// Get Logger
func Get(c context.Context) Logger {
	return c.Value(keyLogger).(Logger)
}

// Set logger
func Set(c context.Context, l Logger) context.Context {
	return context.WithValue(c, keyLogger, l)
}

type fieldLogger struct {
	log logrus.FieldLogger
}

// WrapLogrus is a convenience function
func WrapLogrus(l logrus.FieldLogger) Logger {
	return fieldLogger{log: l}
}

func (l fieldLogger) Debug(v ...interface{}) {
	l.log.Debug(v...)
}
func (l fieldLogger) Debugf(fmt string, v ...interface{}) {
	l.log.Debugf(fmt, v...)
}
func (l fieldLogger) Debugln(v ...interface{}) {
	l.log.Debugln(v...)
}
func (l fieldLogger) Info(v ...interface{}) {
	l.log.Info(v...)
}
func (l fieldLogger) Infof(fmt string, v ...interface{}) {
	l.log.Infof(fmt, v...)
}
func (l fieldLogger) Infoln(v ...interface{}) {
	l.log.Infoln(v...)
}
func (l fieldLogger) Warn(v ...interface{}) {
	l.log.Warn(v...)
}
func (l fieldLogger) Warnf(fmt string, v ...interface{}) {
	l.log.Warnf(fmt, v)
}
func (l fieldLogger) Warnln(v ...interface{}) {
	l.log.Warnln(v...)
}
func (l fieldLogger) Error(v ...interface{}) {
	l.log.Error(v...)
}
func (l fieldLogger) Errorf(fmt string, v ...interface{}) {
	l.log.Errorf(fmt, v...)
}
func (l fieldLogger) Errorln(v ...interface{}) {
	l.log.Errorln(v...)
}
func (l fieldLogger) WithLocus(locus string) Logger {
	return l.WithField(locusLabel, locus)
}
func (l fieldLogger) WithError(err error) Logger {
	return (*entry)(unsafe.Pointer(l.log.WithError(err)))
}
func (l fieldLogger) WithField(k string, v interface{}) Logger {
	return (*entry)(unsafe.Pointer(l.log.WithField(k, v)))
}
func (l fieldLogger) WithFields(f F) Logger {
	return (*entry)(unsafe.Pointer(l.log.WithFields(logrus.Fields(f))))
}

type entry logrus.Entry

func (e *entry) Debug(v ...interface{}) {
	(*logrus.Entry)(unsafe.Pointer(e)).Debug(v...)
}
func (e *entry) Debugf(fmt string, v ...interface{}) {
	(*logrus.Entry)(unsafe.Pointer(e)).Debugf(fmt, v...)
}
func (e *entry) Debugln(v ...interface{}) {
	(*logrus.Entry)(unsafe.Pointer(e)).Debugln(v...)
}

func (e *entry) Info(v ...interface{}) {
	(*logrus.Entry)(unsafe.Pointer(e)).Info(v...)
}
func (e *entry) Infof(fmt string, v ...interface{}) {
	(*logrus.Entry)(unsafe.Pointer(e)).Infof(fmt, v...)
}
func (e *entry) Infoln(v ...interface{}) {
	(*logrus.Entry)(unsafe.Pointer(e)).Infoln(v...)
}

func (e *entry) Warn(v ...interface{}) {
	(*logrus.Entry)(unsafe.Pointer(e)).Warn(v...)
}
func (e *entry) Warnf(fmt string, v ...interface{}) {
	(*logrus.Entry)(unsafe.Pointer(e)).Warnf(fmt, v...)
}
func (e *entry) Warnln(v ...interface{}) {
	(*logrus.Entry)(unsafe.Pointer(e)).Warnln(v...)
}

func (e *entry) Error(v ...interface{}) {
	(*logrus.Entry)(unsafe.Pointer(e)).Error(v...)
}
func (e *entry) Errorf(fmt string, v ...interface{}) {
	(*logrus.Entry)(unsafe.Pointer(e)).Errorf(fmt, v...)
}
func (e *entry) Errorln(v ...interface{}) {
	(*logrus.Entry)(unsafe.Pointer(e)).Errorln(v...)
}

func (e *entry) WithLocus(locus string) Logger {
	return e.WithField(locusLabel, locus)
}
func (e *entry) WithError(err error) Logger {
	return (*entry)(unsafe.Pointer(
		(*logrus.Entry)(unsafe.Pointer(e)).WithError(err),
	))
}
func (e *entry) WithField(k string, v interface{}) Logger {
	return (*entry)(unsafe.Pointer(
		(*logrus.Entry)(unsafe.Pointer(e)).WithField(k, v),
	))
}
func (e *entry) WithFields(f F) Logger {
	return (*entry)(unsafe.Pointer(
		(*logrus.Entry)(unsafe.Pointer(e)).WithFields(
			logrus.Fields(f),
		),
	))
}

type noop struct{}

// NoOp returns a Logger that does nothing
func NoOp() Logger { return noop{} }

func (noop) Debug(...interface{})                 {}
func (noop) Debugf(string, ...interface{})        {}
func (noop) Debugln(...interface{})               {}
func (noop) Info(...interface{})                  {}
func (noop) Infof(string, ...interface{})         {}
func (noop) Infoln(...interface{})                {}
func (noop) Warn(...interface{})                  {}
func (noop) Warnf(string, ...interface{})         {}
func (noop) Warnln(...interface{})                {}
func (noop) Error(...interface{})                 {}
func (noop) Errorf(string, ...interface{})        {}
func (noop) Errorln(...interface{})               {}
func (noop) WithLocus(string) Logger              { return noop{} }
func (noop) WithError(error) Logger               { return noop{} }
func (noop) WithField(string, interface{}) Logger { return noop{} }
func (noop) WithFields(F) Logger                  { return noop{} }

// New logger
func New(opt ...Option) Logger {
	cfg := new(Cfg)
	for _, fn := range opt {
		fn(cfg)
	}
	return cfg.mkLogger()
}
