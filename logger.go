package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
)

// LogLevel represents logging verbosity
type LogLevel int

const (
	LogLevelError LogLevel = iota // Always shown
	LogLevelWarn                  // Always shown
	LogLevelInfo                  // Normal mode
	LogLevelDebug                 // Verbose mode only
)

// Logger provides structured, leveled logging with color support
type Logger struct {
	level     LogLevel
	useColors bool
	errorLog  *log.Logger
	warnLog   *log.Logger
	infoLog   *log.Logger
	debugLog  *log.Logger
}

// NewLogger creates a logger with specified level
func NewLogger(verbose bool) *Logger {
	level := LogLevelInfo
	if verbose {
		level = LogLevelDebug
	}

	return &Logger{
		level:     level,
		useColors: isTerminal(),
		errorLog:  log.New(os.Stderr, "", 0),
		warnLog:   log.New(os.Stdout, "", 0),
		infoLog:   log.New(os.Stdout, "", 0),
		debugLog:  log.New(os.Stdout, "", 0),
	}
}

// SetOutput sets the output for all loggers
func (l *Logger) SetOutput(w io.Writer) {
	l.errorLog.SetOutput(w)
	l.warnLog.SetOutput(w)
	l.infoLog.SetOutput(w)
	l.debugLog.SetOutput(w)
}

// colorize applies color formatting if colors are enabled
func (l *Logger) colorize(color, text string) string {
	if !l.useColors {
		return text
	}
	return color + text + colorReset
}

// Info logs informational messages (always visible in normal mode)
func (l *Logger) Info(format string, args ...interface{}) {
	if l.level >= LogLevelInfo {
		msg := fmt.Sprintf(format, args...)
		l.infoLog.Println(msg)
	}
}

// InfoSuccess logs success with green checkmark
func (l *Logger) InfoSuccess(format string, args ...interface{}) {
	if l.level >= LogLevelInfo {
		icon := l.colorize(colorGreen, "✓")
		msg := fmt.Sprintf(format, args...)
		l.infoLog.Printf("%s %s", icon, msg)
	}
}

// InfoUpdate logs an update operation
func (l *Logger) InfoUpdate(title, detail string) {
	if l.level >= LogLevelInfo {
		icon := l.colorize(colorGreen, "✓")
		titleColored := l.colorize(colorCyan, title)
		l.infoLog.Printf("%s Updated: %s %s", icon, titleColored, detail)
	}
}

// Warn logs warnings (always visible)
func (l *Logger) Warn(format string, args ...interface{}) {
	if l.level >= LogLevelWarn {
		icon := l.colorize(colorYellow, "⚠")
		msg := fmt.Sprintf(format, args...)
		l.warnLog.Printf("%s %s", icon, msg)
	}
}

// Error logs errors (always visible)
func (l *Logger) Error(format string, args ...interface{}) {
	if l.level >= LogLevelError {
		icon := l.colorize(colorRed, "✗")
		msg := fmt.Sprintf(format, args...)
		l.errorLog.Printf("%s %s", icon, msg)
	}
}

// Debug logs debug information (verbose mode only)
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.level >= LogLevelDebug {
		msg := fmt.Sprintf(format, args...)
		l.debugLog.Printf("[DEBUG] %s", msg)
	}
}

// DebugDecision logs decision logic in branches (verbose mode only)
func (l *Logger) DebugDecision(format string, args ...interface{}) {
	if l.level >= LogLevelDebug {
		msg := fmt.Sprintf(format, args...)
		l.debugLog.Printf("[DECISION] %s", msg)
	}
}

// DebugHTTP logs HTTP requests and responses (verbose mode only)
func (l *Logger) DebugHTTP(format string, args ...interface{}) {
	if l.level >= LogLevelDebug {
		msg := fmt.Sprintf(format, args...)
		l.debugLog.Printf("[HTTP] %s", msg)
	}
}

// Stage logs a high-level stage (e.g., "Authenticating...")
func (l *Logger) Stage(format string, args ...interface{}) {
	if l.level >= LogLevelInfo {
		msg := fmt.Sprintf(format, args...)
		colored := l.colorize(colorBold+colorCyan, msg)
		l.infoLog.Println(colored)
	}
}

// Progress logs sync progress (overwrites previous line in TTY, single line otherwise)
func (l *Logger) Progress(current, total int, status string) {
	if l.level < LogLevelInfo {
		return
	}

	if !l.useColors {
		// Non-terminal (file/pipe): only show final message
		if current == total {
			l.infoLog.Printf("Processed %d items\n", total)
		}
		return
	}

	// Terminal: overwrite line with carriage return
	msg := fmt.Sprintf("[%d/%d] Processing %s...", current, total, status)
	l.infoLog.Print("\r\033[K" + msg) // \033[K clears to end of line
	if current == total {
		l.infoLog.Println()
	}
}

// isTerminal checks if output is a terminal (for color support)
func isTerminal() bool {
	// Simple check - if stdout is a terminal
	fileInfo, _ := os.Stdout.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const loggerKey contextKey = "logger"

// WithContext returns a new context with the logger embedded
func (l *Logger) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

// LoggerFromContext retrieves the logger from the context
// Returns nil if no logger is set in the context
func LoggerFromContext(ctx context.Context) *Logger {
	if logger, ok := ctx.Value(loggerKey).(*Logger); ok {
		return logger
	}
	return nil
}

// LogInfo logs an informational message using the logger from context
func LogInfo(ctx context.Context, format string, args ...interface{}) {
	if logger := LoggerFromContext(ctx); logger != nil {
		logger.Info(format, args...)
	}
}

// LogInfoSuccess logs a success message using the logger from context
func LogInfoSuccess(ctx context.Context, format string, args ...interface{}) {
	if logger := LoggerFromContext(ctx); logger != nil {
		logger.InfoSuccess(format, args...)
	}
}

// LogInfoUpdate logs an update operation using the logger from context
func LogInfoUpdate(ctx context.Context, title, detail string) {
	if logger := LoggerFromContext(ctx); logger != nil {
		logger.InfoUpdate(title, detail)
	}
}

// LogWarn logs a warning using the logger from context
func LogWarn(ctx context.Context, format string, args ...interface{}) {
	if logger := LoggerFromContext(ctx); logger != nil {
		logger.Warn(format, args...)
	}
}

// LogError logs an error using the logger from context
func LogError(ctx context.Context, format string, args ...interface{}) {
	if logger := LoggerFromContext(ctx); logger != nil {
		logger.Error(format, args...)
	}
}

// LogDebug logs debug information using the logger from context
func LogDebug(ctx context.Context, format string, args ...interface{}) {
	if logger := LoggerFromContext(ctx); logger != nil {
		logger.Debug(format, args...)
	}
}

// LogDebugDecision logs decision logic using the logger from context
func LogDebugDecision(ctx context.Context, format string, args ...interface{}) {
	if logger := LoggerFromContext(ctx); logger != nil {
		logger.DebugDecision(format, args...)
	}
}

// LogDebugHTTP logs HTTP requests/responses using the logger from context
func LogDebugHTTP(ctx context.Context, format string, args ...interface{}) {
	if logger := LoggerFromContext(ctx); logger != nil {
		logger.DebugHTTP(format, args...)
	}
}

// LogStage logs a high-level stage using the logger from context
func LogStage(ctx context.Context, format string, args ...interface{}) {
	if logger := LoggerFromContext(ctx); logger != nil {
		logger.Stage(format, args...)
	}
}

// LogProgress logs sync progress using the logger from context
func LogProgress(ctx context.Context, current, total int, status string) {
	if logger := LoggerFromContext(ctx); logger != nil {
		logger.Progress(current, total, status)
	}
}
