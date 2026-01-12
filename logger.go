package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
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

// Progress logs sync progress (overwrites previous line)
func (l *Logger) Progress(current, total int, status string) {
	if l.level >= LogLevelInfo {
		msg := fmt.Sprintf("[%d/%d] Processing %s...", current, total, status)
		l.infoLog.Print("\r" + strings.Repeat(" ", 100) + "\r" + msg)
		if current == total {
			l.infoLog.Println() // New line at end
		}
	}
}

// isTerminal checks if output is a terminal (for color support)
func isTerminal() bool {
	// Simple check - if stdout is a terminal
	fileInfo, _ := os.Stdout.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}
