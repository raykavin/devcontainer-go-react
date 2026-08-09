package log

import (
	"fmt"
	"time"

	"my-app/internal/port"
	"github.com/raykavin/gobox/logger"
	glog "github.com/raykavin/gobox/logger"
)

// Wrapper adapts ZerologExtension to the logger interfaces
type Wrapper struct {
	*glog.Zerolog
}

// Ensure Wrapper implements required interfaces
var _ port.Logger = (*Wrapper)(nil)

// Basic logging methods (non-formatted)

// Print logs using Print
func (s *Wrapper) Print(args ...any) {
	s.Logger.Print(args...)
}

// Debug logs at debug level
func (s *Wrapper) Debug(args ...any) {
	s.Logger.Debug().Msg(fmt.Sprint(args...))
}

// Info logs at info level
func (s *Wrapper) Info(args ...any) {
	s.Logger.Info().Msg(fmt.Sprint(args...))
}

// Warn logs at warn level
func (s *Wrapper) Warn(args ...any) {
	s.Logger.Warn().Msg(fmt.Sprint(args...))
}

// Error logs at error level
func (s *Wrapper) Error(args ...any) {
	s.Logger.Error().Msg(fmt.Sprint(args...))
}

// Fatal logs at fatal level
func (s *Wrapper) Fatal(args ...any) {
	s.Logger.Fatal().Msg(fmt.Sprint(args...))
}

// Panic logs at panic level
func (s *Wrapper) Panic(args ...any) {
	s.Logger.Panic().Msg(fmt.Sprint(args...))
}

// Formatted logging methods

// Printf logs using formatted print
func (s *Wrapper) Printf(format string, args ...any) {
	s.Logger.Printf(format, args...)
}

// Debugf logs formatted message at debug level
func (s *Wrapper) Debugf(format string, args ...any) {
	s.Logger.Debug().Msgf(format, args...)
}

// Infof logs formatted message at info level
func (s *Wrapper) Infof(format string, args ...any) {
	s.Logger.Info().Msgf(format, args...)
}

// Warnf logs formatted message at warn level
func (s *Wrapper) Warnf(format string, args ...any) {
	s.Logger.Warn().Msgf(format, args...)
}

// Errorf logs formatted message at error level
func (s *Wrapper) Errorf(format string, args ...any) {
	s.Logger.Error().Msgf(format, args...)
}

// Fatalf logs formatted message at fatal level
func (s *Wrapper) Fatalf(format string, args ...any) {
	s.Logger.Fatal().Msgf(format, args...)
}

// Panicf logs formatted message at panic level
func (s *Wrapper) Panicf(format string, args ...any) {
	s.Logger.Panic().Msgf(format, args...)
}

// Contextual logging

// WithError attaches an error field to the logger
func (s *Wrapper) WithError(err error) port.Logger {
	newLogger := s.With().Err(err).Logger()

	return &Wrapper{
		Zerolog: &glog.Zerolog{
			Logger: &newLogger,
		},
	}
}

// WithField attaches a single structured field
func (s *Wrapper) WithField(key string, value any) port.Logger {
	newLogger := s.With().Interface(key, value).Logger()

	return &Wrapper{
		Zerolog: &glog.Zerolog{
			Logger: &newLogger,
		},
	}
}

// WithFields attaches multiple structured fields
func (s *Wrapper) WithFields(fields map[string]any) port.Logger {
	newLogger := s.With().Fields(fields).Logger()

	return &Wrapper{
		Zerolog: &glog.Zerolog{
			Logger: &newLogger,
		},
	}
}

func NewLogger(loglevel string) (port.Logger, error) {
	log, err := logger.New(&logger.Config{
		Level:          loglevel,
		DateTimeLayout: time.RFC3339,
		Colored:        true,
		JSONFormat:     false,
		UseEmoji:       false,
	})
	if err != nil {
		return nil, err
	}

	return &Wrapper{log}, nil
}
