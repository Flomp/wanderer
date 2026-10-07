package pluginsystem

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"sync"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/security"
)

var (
	ErrConnectionObsolete = errors.New("plugin connection obsolete")
	ErrConnectionConflict = fmt.Errorf("plugin connection changed: %w", ErrConnectionObsolete)
	ErrInstanceDisabled   = fmt.Errorf("plugin instance disabled: %w", ErrConnectionObsolete)
	ErrInstanceBusy       = errors.New("plugin instance busy")
	ErrOAuthFlowConflict  = errors.New("plugin oauth flow changed")
	ErrInstanceExists     = errors.New("plugin instance exists")
)

// ConnectionSnapshot is immutable operation identity. It deliberately contains
// neither secrets nor token/flow revisions: those have their own writer guards.
type ConnectionSnapshot struct {
	InstanceID     string
	UserID         string
	PluginID       string
	AuthContext    string
	Generation     string
	ProviderUserID string
}

func CaptureConnection(record *core.Record) ConnectionSnapshot {
	if record == nil {
		return ConnectionSnapshot{}
	}
	providerUserID, _ := JSONMapFromRecord(record, "auth")["providerUserId"].(string)
	return ConnectionSnapshot{InstanceID: record.Id, UserID: record.GetString("user"), PluginID: record.GetString("plugin_id"), AuthContext: record.GetString("auth_context"), Generation: record.GetString("connection_generation"), ProviderUserID: providerUserID}
}

func (c ConnectionSnapshot) Matches(record *core.Record, requireEnabled bool) bool {
	if record == nil || c.Generation == "" || record.GetString("oauth_token_revision") == "" {
		return false
	}
	fresh := CaptureConnection(record)
	return c.InstanceID == fresh.InstanceID && c.UserID == fresh.UserID && c.PluginID == fresh.PluginID && c.AuthContext == fresh.AuthContext && c.Generation == fresh.Generation && (c.ProviderUserID == "" || c.ProviderUserID == fresh.ProviderUserID) && (!requireEnabled || record.GetBool("enabled"))
}

func CheckConnection(app core.App, c ConnectionSnapshot, requireEnabled bool) (*core.Record, error) {
	if app == nil || c.InstanceID == "" {
		return nil, ErrConnectionConflict
	}
	record, err := app.FindRecordById("plugin_instances", c.InstanceID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrConnectionConflict
	}
	if err != nil {
		return nil, err
	}
	if !c.Matches(record, false) {
		return nil, ErrConnectionConflict
	}
	if requireEnabled && !record.GetBool("enabled") {
		return nil, ErrInstanceDisabled
	}
	return record, nil
}

// PatchConnection reloads under the database transaction and saves only the
// fresh record modified by patch. Remote work must finish before calling it.
func PatchConnection(app core.App, c ConnectionSnapshot, requireEnabled bool, patch func(*core.Record) error) (*core.Record, error) {
	var saved *core.Record
	err := app.RunInTransaction(func(tx core.App) error {
		record, err := CheckConnection(tx, c, requireEnabled)
		if err != nil {
			return err
		}
		if err := patch(record); err != nil {
			return err
		}
		if err := tx.Save(record); err != nil {
			return err
		}
		saved = record.Fresh()
		return nil
	})
	return saved, err
}

func NewConnectionRevision() string { return security.RandomString(40) }

var connectionLocks = struct {
	sync.Mutex
	held map[string]bool
}{held: map[string]bool{}}

// TryLockInstance describes only work held in this backend process. It has no
// waiters and no database side effects; transaction guards provide DB safety.
func TryLockInstance(id string) (func(), bool) {
	connectionLocks.Lock()
	defer connectionLocks.Unlock()
	if id == "" || connectionLocks.held[id] {
		return func() {}, false
	}
	connectionLocks.held[id] = true
	var once sync.Once
	return func() {
		once.Do(func() { connectionLocks.Lock(); delete(connectionLocks.held, id); connectionLocks.Unlock() })
	}, true
}

func InstanceBusy(id string) bool {
	connectionLocks.Lock()
	defer connectionLocks.Unlock()
	return connectionLocks.held[id]
}

func ConnectionAPIError(err error) error {
	code, message := "plugin_connection_conflict", "The plugin connection changed. Reload the instance."
	switch {
	case errors.Is(err, ErrInstanceBusy):
		code, message = "plugin_instance_busy", "Another plugin operation is running. Try again later."
	case errors.Is(err, ErrInstanceDisabled):
		code, message = "plugin_instance_disabled", "The plugin instance is disabled. Reload the instance."
	case errors.Is(err, ErrOAuthFlowConflict):
		code, message = "plugin_oauth_flow_conflict", "The authorization flow changed. Start a new connection."
	case errors.Is(err, ErrInstanceExists):
		code, message = "plugin_instance_exists", "An instance of this plugin already exists."
	}
	return apis.NewApiError(http.StatusConflict, message, validation.Errors{"operation": validation.NewError(code, message)})
}

func ConnectionGenerationRequiredError() error {
	return apis.NewBadRequestError("An expected connection generation is required.", validation.Errors{"operation": validation.NewError("connection_generation_required", "Reload the instance before changing its connection.")})
}
