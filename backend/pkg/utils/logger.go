package utils

import (
	"encoding/json"
	"log"
	"runtime"
	"time"
)

type LogLevel string

const (
	LogLevelInfo    LogLevel = "INFO"
	LogLevelWarning LogLevel = "WARNING"
	LogLevelError   LogLevel = "ERROR"
	LogLevelDebug   LogLevel = "DEBUG"
)

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

func LogInfo(message string, fields ...LogField) {
	logEntry := createLogEntry(LogLevelInfo, message, fields...)
	outputLog(logEntry)
}

func LogWarning(message string, fields ...LogField) {
	logEntry := createLogEntry(LogLevelWarning, message, fields...)
	outputLog(logEntry)
}

func LogError(message string, err error, fields ...LogField) {
	logEntry := createLogEntry(LogLevelError, message, fields...)
	if err != nil {
		logEntry.Error = err.Error()
	}
	logEntry.StackTrace = getStackTrace()
	outputLog(logEntry)
}

func LogDebug(message string, fields ...LogField) {
	logEntry := createLogEntry(LogLevelDebug, message, fields...)
	outputLog(logEntry)
}

type LogField func(*LogEntry)

func WithUserID(userID int) LogField {
	return func(entry *LogEntry) {
		entry.UserID = &userID
	}
}

func WithFeatureID(featureID int) LogField {
	return func(entry *LogEntry) {
		entry.FeatureID = &featureID
	}
}

func WithVoteCount(voteCount int) LogField {
	return func(entry *LogEntry) {
		entry.VoteCount = &voteCount
	}
}

func WithEmail(email string) LogField {
	return func(entry *LogEntry) {
		entry.Email = email
	}
}

func WithUsername(username string) LogField {
	return func(entry *LogEntry) {
		entry.Username = username
	}
}

func WithMethod(method string) LogField {
	return func(entry *LogEntry) {
		entry.Method = method
	}
}

func WithPath(path string) LogField {
	return func(entry *LogEntry) {
		entry.Path = path
	}
}

func WithStatusCode(statusCode int) LogField {
	return func(entry *LogEntry) {
		entry.StatusCode = &statusCode
	}
}

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