//go:build tinygo

package sdk

import (
	"encoding/base64"
	"encoding/json"

	"github.com/extism/go-pdk"
)

//go:wasmimport wanderer http_request
func wandererHTTPRequest(uint64) uint64

//go:wasmimport wanderer log
func wandererLog(uint64)

func Log(level LogLevel, message string) {
	logEntry(HostLogEntry{Level: level, Message: message})
}

// LogEvent emits a safe structured event. Use errors from ResponseError or
// PluginFailure to retain operation and HTTP status, without provider text.
func LogEvent(level LogLevel, event LogEventName, err error) {
	logEntry(EventLogEntry(level, event, err))
}

func logEntry(entry HostLogEntry) {
	memory, err := pdk.AllocateJSON(entry)
	if err != nil {
		return
	}
	defer memory.Free()
	wandererLog(memory.Offset())
}

func LogDebug(message string) {
	Log(LogLevelDebug, message)
}

func LogInfo(message string) {
	Log(LogLevelInfo, message)
}

func LogWarn(message string) {
	Log(LogLevelWarn, message)
}

func LogError(message string) {
	Log(LogLevelError, message)
}

func HostRequest(spec HostRequestSpec) (HostResponse, []byte, error) {
	requestMemory, err := pdk.AllocateJSON(spec)
	if err != nil {
		return HostResponse{}, nil, PluginFailure{PluginError{Code: "invalid_request", Operation: "host_http"}}
	}
	defer requestMemory.Free()

	responsePointer := wandererHTTPRequest(requestMemory.Offset())
	if responsePointer == 0 {
		return HostResponse{}, nil, PluginFailure{PluginError{Code: "runtime_error", Operation: "host_http"}}
	}
	responseMemory := pdk.FindMemory(responsePointer)
	var response HostResponse
	if err := json.Unmarshal(responseMemory.ReadBytes(), &response); err != nil {
		return HostResponse{}, nil, PluginFailure{PluginError{Code: "invalid_output", Operation: "host_http"}}
	}
	if response.Error != nil {
		return response, nil, PluginFailure{NormalizePluginError(*response.Error)}
	}
	body, err := base64.StdEncoding.DecodeString(response.BodyBase64)
	if err != nil {
		return response, nil, PluginFailure{PluginError{Code: "invalid_output", Operation: "host_http"}}
	}
	return response, body, nil
}

func ConnectorRequest(method string, connector string, path string, query []QueryParam, headers map[string]string, expect ResponseExpect) (HostResponse, []byte, error) {
	return HostRequest(HostRequestSpec{
		Method: method,
		Target: RequestTarget{
			Type:      "connector",
			Connector: connector,
			Path:      path,
			Query:     query,
		},
		Headers: headers,
		Expect:  expect,
	})
}

func Get(connector string, path string, query []QueryParam, headers map[string]string, expect ResponseExpect) (HostResponse, []byte, error) {
	return ConnectorRequest("GET", connector, path, query, headers, expect)
}

func PostJSON(connector string, path string, query []QueryParam, headers map[string]string, body any, expect ResponseExpect) (HostResponse, []byte, error) {
	return HostRequest(HostRequestSpec{
		Method: "POST",
		Target: RequestTarget{
			Type:      "connector",
			Connector: connector,
			Path:      path,
			Query:     query,
		},
		Headers: headers,
		Body: &HostRequestBody{
			Type: HostRequestBodyTypeJSON,
			JSON: body,
		},
		Expect: expect,
	})
}

func PostForm(connector string, path string, query []QueryParam, headers map[string]string, form []FormField, expect ResponseExpect) (HostResponse, []byte, error) {
	return HostRequest(HostRequestSpec{
		Method: "POST",
		Target: RequestTarget{
			Type:      "connector",
			Connector: connector,
			Path:      path,
			Query:     query,
		},
		Headers: headers,
		Body: &HostRequestBody{
			Type: HostRequestBodyTypeForm,
			Form: form,
		},
		Expect: expect,
	})
}

func Bool(value bool) *bool {
	return &value
}
