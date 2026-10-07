package pluginsystem

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

// Manager coordinates local plugin discovery with the installed_plugins cache.
// It is intentionally small: request hot paths should read cached manifests,
// while list/cron entrypoints refresh the cache from data/plugins first.
type Manager struct {
	App core.App
	Dir string
}

// PluginInfo is the UI-facing view of an installed plugin. It combines the
// static manifest with runtime availability and embedded icon data.
type PluginInfo struct {
	ID             string               `json:"id"`
	Type           string               `json:"type"`
	Name           string               `json:"name"`
	DisplayName    string               `json:"displayName,omitempty"`
	Description    string               `json:"description,omitempty"`
	Icon           string               `json:"icon,omitempty"`
	IconDark       string               `json:"iconDark,omitempty"`
	Version        string               `json:"version"`
	Runtime        string               `json:"runtime"`
	Path           string               `json:"path"`
	Capabilities   []string             `json:"capabilities"`
	Status         string               `json:"status"`
	SetupErrorCode PluginSetupErrorCode `json:"setupErrorCode,omitempty"`
	Error          string               `json:"error,omitempty"`
	Manifest       Manifest             `json:"manifest"`
}

// NewManager creates a manager for the configured plugin directory. Tests can
// pass a custom dir; production callers use the resolved runtime plugin
// directory.
func NewManager(app core.App, dir string) *Manager {
	if dir == "" {
		dir = PluginDir()
	}
	return &Manager{App: app, Dir: dir}
}

// ListLocalPlugins returns installed plugins in the shape consumed by the
// settings UI. It reads from installed_plugins first so listing does not need to
// parse every manifest from disk after the cache has been refreshed.
func (m *Manager) ListLocalPlugins(context.Context) ([]PluginInfo, error) {
	plugins, err := LoadInstalledPlugins(m.App, m.Dir)
	if err != nil {
		return nil, err
	}
	infos := make([]PluginInfo, 0, len(plugins))
	infoByPath := map[string]int{}
	for _, plugin := range plugins {
		status := "available"
		record, _ := m.App.FindFirstRecordByFilter(
			"installed_plugins",
			"plugin_id={:plugin_id}",
			dbx.Params{"plugin_id": plugin.Manifest.ID},
		)
		if record != nil && record.GetString("status") != "" {
			status = record.GetString("status")
		}
		errorMessage := ""
		if record != nil {
			errorMessage = record.GetString("error")
		}
		icon, iconDark := pluginIcons(plugin)
		infos = append(infos, PluginInfo{
			ID:           plugin.Manifest.ID,
			Type:         plugin.Manifest.Type,
			Name:         plugin.Manifest.Name,
			DisplayName:  stringMetadata(plugin.Manifest.Metadata, "displayName"),
			Description:  plugin.Manifest.Description,
			Icon:         icon,
			IconDark:     iconDark,
			Version:      plugin.Manifest.Version,
			Runtime:      plugin.Manifest.Runtime.Type,
			Path:         plugin.Dir,
			Capabilities: capabilityNames(plugin.Manifest.Capabilities),
			Status:       status,
			Error:        errorMessage,
			Manifest:     plugin.Manifest,
		})
		infoByPath[filepath.Clean(plugin.Dir)] = len(infos) - 1
	}

	_, issues, err := DiscoverLocalPlugins(m.Dir)
	if err != nil {
		return nil, err
	}
	for i := range issues {
		issue := &issues[i]
		if _, ok := infoByPath[filepath.Clean(issue.Dir)]; ok {
			continue
		}
		record, err := m.App.FindFirstRecordByData("installed_plugins", "path", issue.Dir)
		if err != nil {
			continue
		}
		info := PluginInfo{ID: record.GetString("plugin_id"), Name: record.GetString("name"), Type: PluginTypeTrails,
			Path: issue.Dir, Status: "error", Runtime: record.GetString("runtime"), Version: record.GetString("version")}
		_ = record.UnmarshalJSONField("manifest", &info.Manifest)
		infoByPath[filepath.Clean(issue.Dir)] = len(infos)
		infos = append(infos, info)
	}
	return mergePluginIssues(infos, infoByPath, issues), nil
}

// mergePluginIssues overlays current discovery failures on cached plugin
// information. Setup error codes are intentionally derived from this fresh
// discovery pass rather than reconstructed from cached raw diagnostics.
func mergePluginIssues(
	infos []PluginInfo,
	infoByPath map[string]int,
	issues []LocalPluginIssue,
) []PluginInfo {
	for _, issue := range issues {
		if index, ok := infoByPath[filepath.Clean(issue.Dir)]; ok {
			infos[index].Status = "error"
			infos[index].SetupErrorCode = issue.SetupErrorCode
			infos[index].Error = issue.Error
			continue
		}
		infos = append(infos, PluginInfo{
			ID:             issue.ID,
			Type:           PluginTypeTrails,
			Name:           issue.Name,
			Path:           issue.Dir,
			Status:         "error",
			SetupErrorCode: issue.SetupErrorCode,
			Error:          issue.Error,
			Runtime:        RuntimeWASM,
			Manifest: Manifest{
				ID:   issue.ID,
				Type: PluginTypeTrails,
				Name: issue.Name,
				Runtime: RuntimeManifest{
					Type: RuntimeWASM,
				},
			},
		})
	}
	return infos
}

// pluginIcons embeds optional light/dark icon files from the plugin bundle as
// data URLs so the frontend does not need direct filesystem access.
func pluginIcons(plugin LocalPlugin) (string, string) {
	icons, _ := plugin.Manifest.Metadata["icons"].(map[string]any)
	return pluginIcon(plugin.Dir, stringMetadata(icons, "light")), pluginIcon(plugin.Dir, stringMetadata(icons, "dark"))
}

func stringMetadata(values map[string]any, key string) string {
	value, _ := values[key].(string)
	return value
}

func pluginIcon(pluginDir string, iconPath string) string {
	iconPath = strings.TrimSpace(iconPath)
	if iconPath == "" {
		return ""
	}
	cleanPath := filepath.Clean(iconPath)
	if filepath.IsAbs(cleanPath) || cleanPath == ".." || strings.HasPrefix(cleanPath, ".."+string(filepath.Separator)) {
		return ""
	}
	fullPath := filepath.Join(pluginDir, cleanPath)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return ""
	}
	contentType := "image/svg+xml"
	switch strings.ToLower(filepath.Ext(fullPath)) {
	case ".png":
		contentType = "image/png"
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".webp":
		contentType = "image/webp"
	}
	return "data:" + contentType + ";base64," + base64.StdEncoding.EncodeToString(data)
}

// SyncInstalledPlugins scans the runtime plugin directory and upserts
// installed_plugins records.
// This keeps the manifest snapshot available even when later code paths should
// avoid repeated disk IO.
func (m *Manager) SyncInstalledPlugins(ctx context.Context) error {
	plugins, issues, err := DiscoverLocalPlugins(m.Dir)
	if err != nil {
		return err
	}
	// An unchanged listing is read-only. Only a real cache change acquires the
	// serialized writer, where we rebuild the plan against the current data.
	plan, err := m.planInstalledPluginCache(ctx, plugins, issues)
	if err != nil || !plan.changed() {
		return err
	}
	return m.App.RunInTransaction(func(txApp core.App) error {
		manager := Manager{App: txApp, Dir: m.Dir}
		fresh, err := manager.planInstalledPluginCache(ctx, plugins, issues)
		if err != nil {
			return err
		}
		return manager.applyInstalledPluginCache(ctx, fresh)
	})
}

type pluginCacheEntry struct {
	record       *core.Record
	values       map[string]any
	remove       bool
	issue        *LocalPluginIssue
	diagnosticID string
}

type pluginCachePlan []*pluginCacheEntry

func (plan pluginCachePlan) changed() bool {
	for _, entry := range plan {
		if entry.remove || entry.record.IsNew() || !cacheValuesEqual(entry.record, entry.values) {
			return true
		}
	}
	return false
}

func cacheValuesEqual(record *core.Record, values map[string]any) bool {
	for key, wanted := range values {
		// PocketBase JSON fields are JsonRaw; compare decoded values rather than
		// their representation, including recursively ordered object keys.
		a, errA := json.Marshal(record.Get(key))
		b, errB := json.Marshal(wanted)
		var current, target any
		if errA != nil || errB != nil || json.Unmarshal(a, &current) != nil || json.Unmarshal(b, &target) != nil || !reflect.DeepEqual(current, target) {
			return false
		}
	}
	return true
}

func (m *Manager) planInstalledPluginCache(ctx context.Context, plugins []LocalPlugin, issues []LocalPluginIssue) (pluginCachePlan, error) {
	collection, err := m.App.FindCollectionByNameOrId("installed_plugins")
	if err != nil {
		return nil, err
	}
	records, err := m.App.FindRecordsByFilter("installed_plugins", "", "id", -1, 0)
	if err != nil {
		return nil, err
	}
	paths := activePluginPaths(plugins, issues)
	brokenPaths := map[string]bool{}
	healthyIDs := map[string]bool{}
	for _, plugin := range plugins {
		healthyIDs[plugin.Manifest.ID] = true
	}
	for _, issue := range issues {
		brokenPaths[filepath.Clean(issue.Dir)] = true
	}
	plan := make(pluginCachePlan, 0, len(records)+len(plugins)+len(issues))
	byID, byPath := map[string]*pluginCacheEntry{}, map[string]*pluginCacheEntry{}
	for _, record := range records {
		entry := &pluginCacheEntry{record: record, values: map[string]any{}}
		path := filepath.Clean(record.GetString("path"))
		entry.remove = record.GetString("path") == "" || !paths[path]
		plan = append(plan, entry)
		if !entry.remove {
			byID[record.GetString("plugin_id")] = entry
			byPath[path] = entry
		}
	}
	newEntry := func(id, path string) *pluginCacheEntry {
		entry := &pluginCacheEntry{record: core.NewRecord(collection), values: map[string]any{"plugin_id": id, "path": path}}
		plan = append(plan, entry)
		byID[id], byPath[filepath.Clean(path)] = entry, entry
		return entry
	}
	issueID := func(issue LocalPluginIssue, owner *pluginCacheEntry, collision bool) string {
		id := pluginIssueRecordID(issue)
		if collision || healthyIDs[id] || byID[id] != nil && byID[id] != owner {
			id = collisionPluginIssueID(issue)
		}
		base := id
		for suffix := 1; healthyIDs[id] || byID[id] != nil && byID[id] != owner; suffix++ {
			ending := fmt.Sprintf("-%d", suffix)
			prefix := base
			if len(prefix)+len(ending) > 128 {
				prefix = prefix[:128-len(ending)]
			}
			id = strings.TrimRight(prefix, "-_") + ending
		}
		return id
	}
	claimed := map[*pluginCacheEntry]bool{}
	// Valid manifest identities get priority. An unrelated broken directory
	// must never claim their cache row through its directory name.
	for _, plugin := range plugins {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		path := filepath.Clean(plugin.Dir)
		entry := byID[plugin.Manifest.ID]
		pathEntry := byPath[path]
		var inheritedConfig map[string]any
		if entry != nil && entry != pathEntry && filepath.Clean(entry.record.GetString("path")) != path && brokenPaths[filepath.Clean(entry.record.GetString("path"))] {
			var cached Manifest
			if entry.record.UnmarshalJSONField("manifest", &cached) == nil && ValidateManifest(cached) == nil && cached.ID == plugin.Manifest.ID {
				inheritedConfig = JSONMapFromRecord(entry.record, "config")
			}
			// A failed directory remains its own cache entry, including when it
			// once provided this identity. It cannot overwrite the healthy row.
			issue := LocalPluginIssue{ID: filepath.Base(entry.record.GetString("path")), Dir: entry.record.GetString("path")}
			newID := issueID(issue, entry, true)
			entry.values["plugin_id"] = newID
			entry.diagnosticID = newID
			delete(byID, plugin.Manifest.ID)
			byID[newID] = entry
			entry = nil
		}
		if entry == nil {
			// A path may still refer to a row already moved by an earlier
			// identity, or reserved for a healthy identity processed later.
			// Neither may be repurposed for this newly discovered identity.
			if pathEntry != nil && !claimed[pathEntry] && !healthyIDs[pathEntry.record.GetString("plugin_id")] {
				entry = pathEntry
			}
		}
		if entry == nil {
			entry = newEntry(plugin.Manifest.ID, plugin.Dir)
		}
		claimed[entry] = true
		if pathEntry != nil && pathEntry != entry && !claimed[pathEntry] && !healthyIDs[pathEntry.record.GetString("plugin_id")] {
			// Another healthy identity may move to this row's former path later
			// in the same discovery. Keep its record/config for that relocation.
			pathEntry.remove = true
		}
		manifest, err := marshalManifest(plugin.Manifest)
		if err != nil {
			return nil, err
		}
		config := JSONMapFromRecord(entry.record, "config")
		if len(config) == 0 && inheritedConfig != nil {
			config = inheritedConfig
		}
		entry.values = map[string]any{
			"plugin_id": plugin.Manifest.ID, "name": plugin.Manifest.Name,
			"type": plugin.Manifest.Type, "version": plugin.Manifest.Version,
			"runtime": plugin.Manifest.Runtime.Type, "path": plugin.Dir,
			"status": "available", "error": "", "manifest": manifest,
			"config": mergeDefaultConfig(defaultConfig(plugin.Manifest), config),
		}
		byID[plugin.Manifest.ID], byPath[path] = entry, entry
	}
	for _, issue := range issues {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		entry := byPath[filepath.Clean(issue.Dir)]
		if entry == nil {
			id := issueID(issue, nil, false)
			entry = newEntry(id, issue.Dir)
			if id != pluginIssueRecordID(issue) {
				entry.diagnosticID = id
			}
		}
		id := entry.record.GetString("plugin_id")
		if changedID, ok := entry.values["plugin_id"].(string); ok {
			id = changedID
		}
		// Reconstruct the shadow label after reloading a minimal error manifest,
		// so a changed failure cause still belongs to this diagnostic entry.
		if id == issueID(issue, entry, true) {
			entry.diagnosticID = id
		}
		entry.values = map[string]any{
			"plugin_id": id, "name": issue.Name, "type": PluginTypeTrails,
			"version": "unknown", "runtime": RuntimeWASM, "path": issue.Dir,
			"manifest": map[string]any{"id": id, "type": PluginTypeTrails, "name": issue.Name},
			"status":   "error", "error": issue.Error,
		}
		// Preserve validated identity/metadata for diagnostics and the UI. The
		// explicit error status still prevents this manifest from executing.
		var cached Manifest
		if entry.record.UnmarshalJSONField("manifest", &cached) == nil && ValidateManifest(cached) == nil {
			if cached.ID != id {
				entry.diagnosticID = id
			}
			for _, key := range []string{"name", "type", "version", "runtime", "manifest"} {
				entry.values[key] = entry.record.Get(key)
			}
		}
		issueCopy := issue
		entry.issue = &issueCopy
	}
	return plan, nil
}

func (m *Manager) applyInstalledPluginCache(ctx context.Context, plan pluginCachePlan) error {
	// Remove obsolete rows first to free unique plugin IDs, then write changes.
	for _, entry := range plan {
		if !entry.remove {
			continue
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := m.App.Delete(entry.record); err != nil {
			return err
		}
	}
	// Free every changed identity before assigning final IDs. Dependency
	// chains and swaps must not depend on PocketBase record-ID order.
	occupied := map[string]bool{}
	for _, entry := range plan {
		occupied[entry.record.GetString("plugin_id")] = true
		if target, ok := entry.values["plugin_id"].(string); ok {
			occupied[target] = true
		}
	}
	var renamed []*pluginCacheEntry
	for _, entry := range plan {
		if !entry.remove && !entry.record.IsNew() && entry.values["plugin_id"] != nil && entry.values["plugin_id"] != entry.record.GetString("plugin_id") {
			if err := ctx.Err(); err != nil {
				return err
			}
			var temporary string
			for suffix := 0; ; suffix++ {
				temporary = fmt.Sprintf("cache-transition-%s-%d", entry.record.Id, suffix)
				if !occupied[temporary] {
					occupied[temporary] = true
					break
				}
			}
			entry.record.Set("plugin_id", temporary)
			if err := m.App.Save(entry.record); err != nil {
				return err
			}
			renamed = append(renamed, entry)
		}
	}
	for _, entry := range renamed {
		entry.record.Set("plugin_id", entry.values["plugin_id"])
		if err := m.App.Save(entry.record); err != nil {
			return err
		}
	}
	for _, entry := range plan {
		if entry.remove || (!entry.record.IsNew() && cacheValuesEqual(entry.record, entry.values)) {
			continue
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		changedIssue := entry.issue != nil && (entry.record.GetString("status") != "error" || entry.record.GetString("error") != entry.issue.Error)
		diagnosticID := ""
		if changedIssue {
			diagnosticID = pluginSetupDiagnosticID(entry.record, *entry.issue)
			if entry.diagnosticID != "" {
				diagnosticID = entry.diagnosticID
			}
		}
		for key, value := range entry.values {
			entry.record.Set(key, value)
		}
		if err := m.App.Save(entry.record); err != nil {
			return err
		}
		if changedIssue {
			code := PublicPluginSetupErrorCode(entry.issue.SetupErrorCode)
			m.App.TxInfo().OnComplete(func(txErr error) error {
				if txErr == nil {
					m.App.Logger().Warn("plugin setup error", "plugin", diagnosticID, "setup_code", code)
				}
				return nil
			})
		}
	}
	return nil
}

func collisionPluginIssueID(issue LocalPluginIssue) string {
	id := pluginIssueRecordID(issue)
	suffix := "-setup-" + pluginIssueHash(issue)
	if len(id)+len(suffix) > 128 {
		id = id[:128-len(suffix)]
	}
	return strings.TrimRight(id, "-_") + suffix
}

func activePluginPaths(plugins []LocalPlugin, issues []LocalPluginIssue) map[string]bool {
	paths := make(map[string]bool, len(plugins)+len(issues))
	for _, plugin := range plugins {
		paths[filepath.Clean(plugin.Dir)] = true
	}
	for _, issue := range issues {
		paths[filepath.Clean(issue.Dir)] = true
	}
	return paths
}

// Prefer an already validated manifest identity; a discovery directory name
// is otherwise accepted only as a bounded plugin identifier. Invalid names
// become a deterministic bounded label instead of a transliterated path/URL.
// This uses the existing non-cryptographic issue hash solely as an identifier.
func pluginSetupDiagnosticID(record *core.Record, issue LocalPluginIssue) string {
	var manifest Manifest
	if record.UnmarshalJSONField("manifest", &manifest) == nil && ValidateManifest(manifest) == nil &&
		len(manifest.ID) <= 128 && pluginIDPattern.MatchString(manifest.ID) {
		return manifest.ID
	}
	if len(issue.ID) <= 128 && pluginIDPattern.MatchString(issue.ID) {
		return issue.ID
	}
	return "plugin-setup-error-" + pluginIssueHash(issue)
}

func pluginIssueRecordID(issue LocalPluginIssue) string {
	originalID := strings.TrimSpace(issue.ID)
	id := strings.ToLower(originalID)
	var builder strings.Builder
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
		case r == '_' || r == '-':
			builder.WriteRune(r)
		default:
			builder.WriteRune('-')
		}
	}
	result := strings.Trim(builder.String(), "-_")
	if result == "" {
		result = "plugin-setup-error"
	}
	if originalID != result || !pluginIDPattern.MatchString(result) {
		result = strings.Trim(result, "-_")
		if result == "" {
			result = "plugin-setup-error"
		}
		result = result + "-" + pluginIssueHash(issue)
	}
	if len(result) > 128 {
		hash := pluginIssueHash(issue)
		prefixLength := 128 - len(hash) - 1
		result = strings.Trim(result[:prefixLength], "-_") + "-" + hash
	}
	if pluginIDPattern.MatchString(result) {
		return result
	}
	return "plugin-setup-error-" + pluginIssueHash(issue)
}

func pluginIssueHash(issue LocalPluginIssue) string {
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(issue.Dir))
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write([]byte(issue.ID))
	return fmt.Sprintf("%08x", hash.Sum32())
}

func marshalManifest(manifest Manifest) (map[string]any, error) {
	data, err := json.Marshal(manifest)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func defaultConfig(manifest Manifest) map[string]any {
	hostConfig, _ := CloneJSONValue(manifest.HostConfig).(map[string]any)
	if hostConfig == nil {
		hostConfig = map[string]any{}
	}
	config := map[string]any{
		"host": hostConfig,
	}
	pluginConfig := map[string]any{}
	for _, field := range manifest.ConfigSchema {
		if field.Key == "" || field.Default == nil {
			continue
		}
		pluginConfig[field.Key] = CloneJSONValue(field.Default)
	}
	config["plugin"] = pluginConfig
	return config
}

func mergeDefaultConfig(defaults map[string]any, current map[string]any) map[string]any {
	if len(defaults) == 0 {
		return current
	}
	merged := CloneJSONMap(defaults)
	MergePluginConfig(merged, current)
	return merged
}

func MergePluginConfig(dst map[string]any, src map[string]any) {
	DeepMergeConfigWithReplaceKeys(dst, src, map[string]bool{
		"categoryMapping": true,
	})
}

func capabilityNames(capabilities []CapabilityManifest) []string {
	names := make([]string, 0, len(capabilities))
	for _, capability := range capabilities {
		names = append(names, capability.Name+"."+capability.Version)
	}
	return names
}
