package routes

import (
	"errors"
	"reflect"
	"time"

	"github.com/pocketbase/pocketbase/core"
	"pocketbase/pluginsystem"
)

func pluginSyncSkipReason(err error) string {
	if errors.Is(err, pluginsystem.ErrInstanceBusy) {
		return "plugin_instance_busy"
	}
	if errors.Is(err, pluginsystem.ErrInstanceDisabled) {
		return "plugin_instance_disabled"
	}
	return "plugin_connection_obsolete"
}

func setPluginInstanceStatusForError(app core.App, instance *core.Record, err error) {
	if errors.Is(err, pluginsystem.ErrConnectionObsolete) || errors.Is(err, pluginsystem.ErrInstanceBusy) {
		return
	}
	update := pluginsystem.InstanceStatusForError(err, time.Now())
	// Detail/upload failures belong to the operation; they do not establish a
	// connection-wide fault. Host login/refresh failures have their own source.
	if update.Operation == pluginsystem.OperationDetail || update.Operation == pluginsystem.OperationSend {
		return
	}
	snapshot := pluginsystem.CaptureConnection(instance)
	tokenRevision := instance.GetString("oauth_token_revision")
	storedAuth := pluginsystem.JSONMapFromRecord(instance, "auth")
	oauth := update.Operation == pluginsystem.OperationOAuthRefresh || update.Operation == pluginsystem.OperationOAuthExchange
	if manifest, manifestErr := pluginsystem.InstalledPluginManifest(app, snapshot.PluginID); manifestErr == nil {
		_, context, contextErr := pluginsystem.EffectiveAuthContext(manifest, snapshot.AuthContext)
		oauth = oauth || contextErr == nil && context.Type == pluginsystem.AuthTypeOAuth2
	} else {
		// The provenance marker can only tighten the guard when a manifest is
		// unavailable; it does not select the active context.
		oauth = oauth || storedAuth[pluginsystem.AuthFieldOAuthContext] == snapshot.AuthContext && snapshot.AuthContext != ""
	}
	_, saveErr := pluginsystem.PatchConnection(app, snapshot, true, func(fresh *core.Record) error {
		if oauth && (tokenRevision == "" || fresh.GetString("oauth_token_revision") != tokenRevision) {
			return pluginsystem.ErrOAuthTokenRevisionConflict
		}
		diagnosis := map[string]any{"code": update.Code, "message": update.Message, "operation": update.Operation, "httpStatus": update.HTTPStatus, "authContext": snapshot.AuthContext, "connectionGeneration": snapshot.Generation}
		if oauth {
			diagnosis["oauthTokenRevision"] = tokenRevision
		}
		fresh.Set("status", update.Status)
		fresh.Set("last_error", diagnosis)
		if update.RetryNotBefore != nil {
			fresh.Set("retry_not_before", *update.RetryNotBefore)
		} else {
			fresh.Set("retry_not_before", "")
		}
		return nil
	})
	if saveErr != nil && !errors.Is(saveErr, pluginsystem.ErrConnectionObsolete) && !errors.Is(saveErr, pluginsystem.ErrOAuthTokenRevisionConflict) {
		pluginsystem.LogPluginFailure(app.Logger(), "failed to update plugin instance status", pluginsystem.WithDiagnosticCause(saveErr, "database_save"), "instance", instance.Id)
	}
}

func completePluginSync(app core.App, instance *core.Record, completedAt time.Time) error {
	snapshot := pluginsystem.CaptureConnection(instance)
	observedError := pluginsystem.JSONMapFromRecord(instance, "last_error")
	observedBackoff := instance.GetString("retry_not_before")
	observedState := pluginsystem.JSONMapFromRecord(instance, "state")
	operationTokenRevision := instance.GetString("oauth_token_revision")
	_, err := pluginsystem.PatchConnection(app, snapshot, true, func(fresh *core.Record) error {
		state := pluginsystem.JSONMapFromRecord(fresh, "state")
		matches := true
		for key, observed := range observedState {
			current, present := state[key]
			if !present || !reflect.DeepEqual(current, observed) {
				matches = false
				break
			}
		}
		if matches {
			for key := range observedState {
				delete(state, key)
			}
			fresh.Set("state", state)
		}
		if fresh.GetDateTime("last_sync_at").Time().Before(completedAt) {
			fresh.Set("last_sync_at", completedAt)
		}
		if operationTokenRevision != "" && fresh.GetString("oauth_token_revision") == operationTokenRevision && reflect.DeepEqual(pluginsystem.JSONMapFromRecord(fresh, "last_error"), observedError) && fresh.GetString("retry_not_before") == observedBackoff && syncCanClearDiagnosis(observedError, snapshot) {
			fresh.Set("last_error", map[string]any{})
			fresh.Set("retry_not_before", "")
			fresh.Set("status", "configured")
		}
		return nil
	})
	if err != nil && !errors.Is(err, pluginsystem.ErrConnectionObsolete) {
		return pluginsystem.WithDiagnosticCause(err, "database_save")
	}
	return err
}

func syncCanClearDiagnosis(diagnosis map[string]any, c pluginsystem.ConnectionSnapshot) bool {
	if len(diagnosis) == 0 {
		return true
	}
	context, contextPresent := diagnosis["authContext"]
	generation, generationPresent := diagnosis["connectionGeneration"]
	// Migration preserves historical diagnostics; a complete unchanged sync is
	// the sole runtime operation allowed to clear their unbound tuple.
	if !contextPresent && !generationPresent {
		return true
	}
	if context != c.AuthContext || generation != c.Generation {
		return false
	}
	return true
}

func pluginConnectionAPIErrorForEvent(e *core.RequestEvent, logContext pluginOperationLogContext, err error) error {
	if errors.Is(err, pluginsystem.ErrConnectionObsolete) || errors.Is(err, pluginsystem.ErrOAuthFlowConflict) || errors.Is(err, pluginsystem.ErrInstanceBusy) {
		return pluginsystem.ConnectionAPIError(err)
	}
	return pluginHostAPIErrorForEvent(e, logContext, "Failed to load plugin connection.", pluginsystem.WithDiagnosticCause(err, "database_lookup"))
}
