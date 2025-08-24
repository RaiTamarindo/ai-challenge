package logs

import (
	"encoding/json"
	"log"
	"runtime"
	"time"
)

// LogLevel represents the severity level of a log entry
type LogLevel string

const (
	LogLevelInfo    LogLevel = "INFO"
	LogLevelWarning LogLevel = "WARNING"
	LogLevelError   LogLevel = "ERROR"
	LogLevelDebug   LogLevel = "DEBUG"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp  time.Time              `json:"timestamp"`
	Level      LogLevel               `json:"level"`
	Message    string                 `json:"message"`
	UserID     *int                   `json:"user_id,omitempty"`
	FeatureID  *int                   `json:"feature_id,omitempty"`
	VoteCount  *int                   `json:"vote_count,omitempty"`
	Email      string                 `json:"email,omitempty"`
	Username   string                 `json:"username,omitempty"`
	Method     string                 `json:"method,omitempty"`
	Path       string                 `json:"path,omitempty"`
	StatusCode *int                   `json:"status_code,omitempty"`
	Error      string                 `json:"error,omitempty"`
	StackTrace string                 `json:"stack_trace,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// Logger defines the interface for logging operations
type Logger interface {
	Info(message string, fields ...LogField)
	Warning(message string, fields ...LogField)
	Error(message string, err error, fields ...LogField)
	Debug(message string, fields ...LogField)
}

// JSONLogger implements Logger interface with JSON structured logging
type JSONLogger struct{}

// NewJSONLogger creates a new JSON logger
func NewJSONLogger() *JSONLogger {
	return &JSONLogger{}
}

// Info logs an info message
func (l *JSONLogger) Info(message string, fields ...LogField) {
	logEntry := createLogEntry(LogLevelInfo, message, fields...)
	outputLog(logEntry)
}

// Warning logs a warning message
func (l *JSONLogger) Warning(message string, fields ...LogField) {
	logEntry := createLogEntry(LogLevelWarning, message, fields...)
	outputLog(logEntry)
}

// Error logs an error message with stack trace
func (l *JSONLogger) Error(message string, err error, fields ...LogField) {
	logEntry := createLogEntry(LogLevelError, message, fields...)
	if err != nil {
		logEntry.Error = err.Error()
	}
	logEntry.StackTrace = getStackTrace()
	outputLog(logEntry)
}

// Debug logs a debug message
func (l *JSONLogger) Debug(message string, fields ...LogField) {
	logEntry := createLogEntry(LogLevelDebug, message, fields...)
	outputLog(logEntry)
}

// LogField is a function that modifies a log entry
type LogField func(*LogEntry)

// WithUserID adds user ID to log entry
func WithUserID(userID int) LogField {
	return func(entry *LogEntry) {
		entry.UserID = &userID
	}
}

// WithFeatureID adds feature ID to log entry
func WithFeatureID(featureID int) LogField {
	return func(entry *LogEntry) {
		entry.FeatureID = &featureID
	}
}

// WithVoteCount adds vote count to log entry
func WithVoteCount(voteCount int) LogField {
	return func(entry *LogEntry) {
		entry.VoteCount = &voteCount
	}
}

// WithEmail adds email to log entry
func WithEmail(email string) LogField {
	return func(entry *LogEntry) {
		entry.Email = email
	}
}

// WithUsername adds username to log entry
func WithUsername(username string) LogField {
	return func(entry *LogEntry) {
		entry.Username = username
	}
}

// WithMethod adds HTTP method to log entry
func WithMethod(method string) LogField {
	return func(entry *LogEntry) {
		entry.Method = method
	}
}

// WithPath adds request path to log entry
func WithPath(path string) LogField {
	return func(entry *LogEntry) {
		entry.Path = path
	}
}

// WithStatusCode adds HTTP status code to log entry
func WithStatusCode(statusCode int) LogField {
	return func(entry *LogEntry) {
		entry.StatusCode = &statusCode
	}
}

// WithMetadata adds custom metadata to log entry
func WithMetadata(key string, value interface{}) LogField {
	return func(entry *LogEntry) {
		if entry.Metadata == nil {
			entry.Metadata = make(map[string]interface{})
		}
		entry.Metadata[key] = value
	}
}

func createLogEntry(level LogLevel, message string, fields ...LogField) *LogEntry {
	entry := &LogEntry{
		Timestamp: time.Now().UTC(),
		Level:     level,
		Message:   message,
	}

	for _, field := range fields {
		field(entry)
	}

	return entry
}

func outputLog(entry *LogEntry) {
	jsonBytes, err := json.Marshal(entry)
	if err != nil {
		log.Printf("Error marshalling log entry: %v", err)
		return
	}
	log.Println(string(jsonBytes))
}

func getStackTrace() string {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	stackTrace := ""
	for {
		frame, more := frames.Next()
		if frame.File != "" {
			stackTrace += frame.File + ":" + frame.Function + ":" + 
				runtime.FuncForPC(frame.PC).Name() + "\n"
		}
		if !more {
			break
		}
	}
	return stackTrace
}