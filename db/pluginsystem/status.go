package pluginsystem

import (
	"time"
)

// PluginCapabilityError wraps a plugin-reported error returned inside a
// successful export response, so status mapping can treat it like runtime
// PluginCallError failures.
type PluginCapabilityError struct {
	Err *PluginError
}

func (e PluginCapabilityError) Error() string {
	if e.Err == nil {
		return "plugin_error"
	}
	return NormalizePluginError(*e.Err).Message
}

// InstanceStatusUpdate contains the normalized status fields that are written
// back to plugin_instances after a failed sync.
type InstanceStatusUpdate struct {
	Status         string
	Code           string
	Message        string
	Operation      string
	HTTPStatus     int
	RetryNotBefore *time.Time
}

// InstanceStatusForError converts sync/runtime errors into the persisted
// plugin_instances status fields used by the UI and cron backoff logic.
func InstanceStatusForError(err error, now time.Time) InstanceStatusUpdate {
	return InstanceStatusForPluginError(PublicPluginError(err), now)
}

// InstanceStatusForPluginError maps the stable plugin error codes from the ABI
// to host instance states. retryAfterSeconds wins over default retry windows.
func InstanceStatusForPluginError(pluginErr PluginError, now time.Time) InstanceStatusUpdate {
	pluginErr = NormalizePluginError(pluginErr)
	code := pluginErr.Code
	message := pluginErr.Message

	status := "error"
	switch code {
	case "auth_failed", "credentials_unreadable", "invalid_grant", "invalid_client", "unauthorized_client":
		status = "needs_reauth"
	case "unauthorized":
		// A rejected item/upload alone does not invalidate the stored grant.
		// Legacy guests without operation metadata retain their v1 behavior.
		if pluginErr.Operation != OperationDetail && pluginErr.Operation != OperationSend {
			status = "needs_reauth"
		}
	case "rate_limited":
		status = "rate_limited"
	case "provider_unavailable", "temporary_unavailable":
		status = "unavailable"
	}

	update := InstanceStatusUpdate{
		Status:     status,
		Code:       code,
		Message:    message,
		Operation:  pluginErr.Operation,
		HTTPStatus: pluginErr.HTTPStatus,
	}
	if pluginErr.RetryAfterSeconds != nil && *pluginErr.RetryAfterSeconds >= 0 {
		retryNotBefore := now.Add(time.Duration(*pluginErr.RetryAfterSeconds) * time.Second)
		update.RetryNotBefore = &retryNotBefore
	} else if code == "rate_limited" {
		retryNotBefore := now.Add(time.Hour)
		update.RetryNotBefore = &retryNotBefore
	}
	return update
}
