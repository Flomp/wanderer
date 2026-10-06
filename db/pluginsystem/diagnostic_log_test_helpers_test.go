package pluginsystem

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"log/slog"
	"testing"
)

// SetDefault also changes the standard logger. Restore both APIs so captured
// records do not change the diagnostic format of tests that run afterwards.
func setPluginTestDefaultLogger(t *testing.T, logger *slog.Logger) {
	t.Helper()
	previous, writer, flags, prefix := slog.Default(), log.Writer(), log.Flags(), log.Prefix()
	slog.SetDefault(logger)
	t.Cleanup(func() { slog.SetDefault(previous); log.SetOutput(writer); log.SetFlags(flags); log.SetPrefix(prefix) })
}
func setPluginTestStandardLogWriter(t *testing.T, writer io.Writer) {
	t.Helper()
	previous, flags, prefix := log.Writer(), log.Flags(), log.Prefix()
	log.SetOutput(writer)
	t.Cleanup(func() { log.SetOutput(previous); log.SetFlags(flags); log.SetPrefix(prefix) })
}
func singlePluginDiagnosticRecord(t *testing.T, output *bytes.Buffer) map[string]any {
	t.Helper()
	data := bytes.TrimSpace(output.Bytes())
	if len(data) == 0 {
		t.Fatal("guest boundary emitted no diagnostic")
	}
	lines := bytes.Split(data, []byte("\n"))
	if len(lines) != 1 {
		t.Fatalf("guest boundary emitted %d diagnostics: %s", len(lines), data)
	}
	var record map[string]any
	if err := json.Unmarshal(lines[0], &record); err != nil {
		t.Fatalf("invalid structured diagnostic: %s: %v", data, err)
	}
	return record
}
