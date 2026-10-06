package sdk

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestEventLogEntryPreservesSafeMetadataAndDiscardsText(t *testing.T) {
	entry := EventLogEntry(LogLevelWarn, LogEventRequestFailed,
		PluginFailure{PluginError{Code: "rate_limited", Message: "provider-secret", Operation: "list", HTTPStatus: 429}})
	if entry.Error == nil || entry.Error.Code != "rate_limited" || entry.Error.HTTPStatus != 429 || entry.Error.Operation != "list" {
		t.Fatalf("lost event diagnostics: %#v", entry)
	}
	for _, entry := range []HostLogEntry{entry, EventLogEntry(LogLevelError, "https://secret.invalid", errors.New("provider-body-secret"))} {
		data, err := json.Marshal(entry)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(data), "secret") || strings.Contains(string(data), "https://") {
			t.Fatalf("untrusted text in log event: %s", data)
		}
	}
}
