package routes

import (
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pocketbase/pocketbase/apis"

	"pocketbase/pluginsystem"
)

// PocketBase preserves codes only for SafeErrorItems (validation.Error). Raw
// scalar maps are rewritten to validation_invalid_value. Keep safe metadata in
// params so proxies and the UI can distinguish errors without parsing text.
func pluginOperationAPIError(err error) error {
	safe := pluginsystem.PublicPluginError(err)
	params := map[string]any{}
	if safe.Operation != "" {
		params["operation"] = safe.Operation
	}
	if safe.HTTPStatus != 0 {
		params["httpStatus"] = safe.HTTPStatus
	}
	if safe.RetryAfterSeconds != nil {
		params["retryAfterSeconds"] = *safe.RetryAfterSeconds
	}
	return apis.NewBadRequestError("Plugin operation failed.", validation.Errors{
		"operation": validation.NewError(safe.Code, safe.Message).SetParams(params),
	})
}

func firstProviderHeader(headers map[string][]string, name string) string {
	for key, values := range headers {
		if strings.EqualFold(key, name) && len(values) > 0 {
			return values[0]
		}
	}
	return ""
}
