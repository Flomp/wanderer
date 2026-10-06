package sdk

// LogEventName is a closed diagnostic vocabulary, not a provider text field.
type LogEventName string

const (
	LogEventRequestStarted   LogEventName = "request_started"
	LogEventRequestCompleted LogEventName = "request_completed"
	LogEventRequestFailed    LogEventName = "request_failed"
	LogEventSyncStarted      LogEventName = "sync_started"
	LogEventSyncCompleted    LogEventName = "sync_completed"
	LogEventSyncFailed       LogEventName = "sync_failed"
)

// EventLogEntry can also be used by plugins supplying their own host bindings.
// Arbitrary event names/messages are never forwarded into diagnostic logs.
func EventLogEntry(level LogLevel, event LogEventName, err error) HostLogEntry {
	switch event {
	case LogEventRequestStarted, LogEventRequestCompleted, LogEventRequestFailed,
		LogEventSyncStarted, LogEventSyncCompleted, LogEventSyncFailed:
	default:
		event = "guest_message_redacted"
	}
	entry := HostLogEntry{Level: level, Event: event, Message: string(event)}
	if err != nil {
		e := SafeError(err, "")
		entry.Error = &e
	}
	return entry
}
