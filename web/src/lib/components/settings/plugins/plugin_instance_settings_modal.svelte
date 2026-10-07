<script lang="ts">
    import Datepicker from "$lib/components/base/datepicker.svelte";
    import Modal from "$lib/components/base/modal.svelte";
    import Select, { type SelectItem } from "$lib/components/base/select.svelte";
    import SingleSelect from "$lib/components/base/single_select.svelte";
    import TextField from "$lib/components/base/text_field.svelte";
    import Toggle from "$lib/components/base/toggle.svelte";
    import PluginMergeSettings from "$lib/components/settings/plugins/plugin_merge_settings.svelte";
    import CategoryPicker from "$lib/components/trail/category_picker.svelte";
    import type { Category } from "$lib/models/category";
    import type { PluginInstance, PluginInstanceDraft } from "$lib/models/plugin_instance";
    import type { ConfigField, PluginAuthContext, PluginProvider } from "$lib/models/plugin_provider";
    import type { Subcategory } from "$lib/models/subcategory";
    import { plugin_auth_validate, plugin_oauth_start, plugin_oauth_revoke, plugin_instances_show } from "$lib/stores/plugin_instance_store";
    import { show_toast } from "$lib/stores/toast_store.svelte";
    import {
        categoryMappingTargetFromUnknown,
        categoryMappingTargetToPickerValue,
        type CategoryMappingTarget,
    } from "$lib/util/category_util";
    import { resolveProviderAuth, resolvePluginAuthContext, editableAuthFields, hasPluginConnectionMaterial, pluginSecretFields } from "$lib/util/plugin_auth_context";
    import { pluginSettingsConfigPatch, hasProviderConfigChanges } from "$lib/util/plugin_settings_patch";
    import { mergePluginConfig } from "$lib/util/plugin_config";
    import { translatePluginAPIError, pluginOperationNeedsReload } from "$lib/util/plugin_error_i18n";
    import {
        configFieldDescription,
        configFieldLabel,
        configFieldOptionLabel,
        pluginTitle as localizedPluginTitle,
        providerCategoryLabel,
    } from "$lib/util/plugin_i18n";
    import { tick } from "svelte";
    import { _, locale } from "svelte-i18n";

    interface CategoryMappingRow {
        providerCategory: string;
        category: string;
    }

    type PluginInstanceForm = PluginInstanceDraft & {
        mappingChanged?: boolean;
    };

    interface Props {
        plugin: PluginProvider;
        categories?: Category[];
        subcategories?: Subcategory[];
        instance?: PluginInstance;
        onbeforecategorymappingsave?: (instance: PluginInstanceForm) => Promise<boolean> | boolean;
        onsave?: (instance: PluginInstanceDraft) => Promise<PluginInstance | void> | PluginInstance | void;
    }

    let { plugin, categories = [], subcategories = [], instance, onbeforecategorymappingsave, onsave }: Props = $props();

    let modal: Modal;
    let openingInstance: PluginInstance | undefined = $state();
    let snapshotTaken = $state(false);
    let draftInstance = $derived(snapshotTaken ? openingInstance : instance);
    let draftStale = $derived(snapshotTaken && (openingInstance?.id !== instance?.id ||
        openingInstance?.connection_generation !== instance?.connection_generation));
    let authContext = $state(initialAuthContext());
    let clearedAuth: Record<string, boolean> = $state({});
    let openingConfig: Record<string, unknown> = {};
    let auth: Record<string, string> = $state(initialAuth());
    let planned = $state(true);
    let completed = $state(true);
    let mergeEnabled = $state(false);
    let privacy = $state("original");
    let categoryMappingRows: CategoryMappingRow[] = $state(initialCategoryMappingRows());
    let categoryMappingList: HTMLDivElement | undefined = $state();
    let resolvedAuth = $derived(resolveAuth());
    let authFields = $derived(editableAuthFields(resolvedAuth.auth));
    let secretFields = $derived(pluginSecretFields(plugin));
    let isOAuthPlugin = $derived(resolvedAuth.auth.type === "oauth2");
    let isSessionPlugin = $derived(resolvedAuth.auth.type === "session");
    let isConnected = $derived(hasPluginConnectionMaterial(plugin, draftInstance
        ? { ...draftInstance, auth_context: authContext } : undefined));
    let isSaving = $state(false);
    let plugin_id = $derived(plugin.id);
    let configSchema = $derived(plugin.configSchema ?? []);
    let visibleConfigSchema = $derived(configSchema.filter((field) => !field.hidden));
    let extraConfig: Record<string, any> = $state(initialExtraConfig());
    let configErrors: Record<string, string> = $state({});
    let supportsPlanned = $derived(
        plugin.capabilities?.includes("list_routes.v1") ?? false,
    );
    let supportsCompleted = $derived(
        plugin.capabilities?.includes("list_activities.v1") ?? false,
    );
    let hasTourKindChoice = $derived(supportsPlanned && supportsCompleted);
    let supportsSourcePrivacy = $derived(
        plugin.capabilities?.includes("source_privacy") ?? false,
    );
    let mergeAvailable = $derived((hostConfig().merge as any)?.available !== false);
    let supportsCategoryMapping = $derived(supportsPlanned || supportsCompleted);
    let providerCategorySelectItems: SelectItem[] = $derived(providerCategoryItems());
    let canAddCategoryMappingRow = $derived(
        categoryMappingRows.every((row) => row.providerCategory && row.category) &&
        providerCategorySelectItems.some(
            (item) => !categoryMappingRows.some((row) => row.providerCategory === item.value),
        ),
    );

    const configLabels: Record<string, string> = {
        after: "ignore-trails-before-date",
    };
    const configHints: Record<string, string> = {
        after: "plugin-after-date-hint",
    };
    const privacySelectItems: SelectItem[] = [
        { text: $_("keep-original"), value: "original" },
        { text: $_("apply-user-settings"), value: "settings" },
    ];
    const authLabels: Record<string, string> = {
        email: $_("email"),
        password: $_("password"),
        clientId: "Client ID",
        clientSecret: "Client Secret",
    };

    function initialAuthContext() {
        return draftInstance?.auth_context ?? (!draftInstance ? plugin.authContext : "") ?? "";
    }

    function resolveAuth(): { name?: string; auth: PluginAuthContext; error?: Error } {
        try {
            if (plugin.authContexts) {
                if (!authContext && draftInstance && Object.keys(plugin.authContexts).length) {
                    throw new Error("Plugin authentication context is missing");
                }
                return resolvePluginAuthContext(plugin.authContexts, authContext || undefined);
            }
            return resolveProviderAuth(plugin, draftInstance);
        } catch (error) {
            return { auth: { type: "unresolved" }, error: error as Error };
        }
    }

    function initialAuth() {
        const fields = [...new Set(Object.values(plugin.authContexts ?? { active: plugin.auth })
            .flatMap(editableAuthFields))];
        return Object.fromEntries(
            fields.map((field) => [
                field,
                draftInstance?.auth?.[field] == null ? "" : String(draftInstance.auth[field]),
            ]),
        );
    }

    function initialExtraConfig(): Record<string, any> {
        const config = pluginConfig();
        return Object.fromEntries(
            configSchema.map((field) => {
                const saved = config[field.key];
                if (saved !== undefined) return [field.key, saved];
                if (field.type === "boolean") {
                    return [field.key, booleanDefault(field.default, false)];
                }
                if (field.default !== undefined && field.default !== null) {
                    return [field.key, String(field.default)];
                }
                if (field.type === "select" && field.options?.length) {
                    return [field.key, field.options[0].value];
                }
                if (field.type === "select") {
                    return [field.key, ""];
                }
                if (field.type === "text" || field.type === "url") {
                    return [field.key, ""];
                }
                return [field.key, undefined];
            }),
        );
    }

    function initialCategoryMappingRows(): CategoryMappingRow[] {
        return Object.entries(categoryMapping())
            .filter(([, target]) => !isBlankCategoryMappingTarget(target))
            .map(([providerCategory, target]) => ({
                providerCategory,
                category: categoryPickerValue(target),
            }));
    }

    function manifestCategoryMapping(): Record<string, CategoryMappingTarget> {
        const raw = plugin.hostConfig?.categoryMapping;
        if (!raw || typeof raw !== "object" || Array.isArray(raw)) {
            return {};
        }
        return categoryTargetMapping(raw as Record<string, unknown>);
    }

    function categoryMapping(): Record<string, CategoryMappingTarget> {
        const raw = hostConfig().categoryMapping;
        if (!raw || typeof raw !== "object" || Array.isArray(raw)) {
            return {};
        }
        return categoryTargetMapping(raw as Record<string, unknown>);
    }

    function categoryTargetMapping(raw: Record<string, unknown>): Record<string, CategoryMappingTarget> {
        return Object.fromEntries(
            Object.entries(raw)
                .map(([key, value]) => [
                    key,
                    categoryMappingTargetFromUnknown(value),
                ])
                .filter(([, value]) => value !== undefined),
        );
    }

    function isBlankCategoryMappingTarget(target: CategoryMappingTarget): boolean {
        if (typeof target === "string") {
            return target.trim() === "";
        }
        return !target.category?.trim() && !target.subcategory?.trim();
    }

    function categoryPickerValue(target: CategoryMappingTarget): string {
        return categoryMappingTargetToPickerValue(
            target,
            categories,
            subcategories,
        );
    }

    function categoryMappingTargetFromPickerValue(value: string): CategoryMappingTarget {
        if (value.startsWith("subcategory:")) {
            const subcategoryId = value.replace("subcategory:", "");
            const subcategory = subcategories.find((candidate) => candidate.id === subcategoryId);
            return {
                category: subcategory?.category ?? "",
                subcategory: subcategoryId,
            };
        }
        if (value.startsWith("category:")) {
            return {
                category: value.replace("category:", ""),
                subcategory: "",
            };
        }
        return "";
    }

    function categoryPickerCurrentCategoryId(value: string): string | null {
        if (value.startsWith("category:")) {
            return value.replace("category:", "");
        }
        if (value.startsWith("subcategory:")) {
            const subcategoryId = value.replace("subcategory:", "");
            return subcategories.find((candidate) => candidate.id === subcategoryId)?.category ?? null;
        }
        return null;
    }

    function providerCategoryItems(): SelectItem[] {
        const values = new Set<string>([
            ...Object.keys(manifestCategoryMapping()),
            ...Object.keys(categoryMapping()),
            ...categoryMappingRows.map((row) => row.providerCategory).filter(Boolean),
        ]);
        return [...values]
            .map((value) => ({
                text: providerCategoryLabel(plugin, value, $locale),
                value,
            }))
            .sort((a, b) => a.text.localeCompare(b.text, $locale ?? undefined));
    }

    function providerCategoryItemsForRow(index: number): SelectItem[] {
        const currentValue = categoryMappingRows[index]?.providerCategory;
        const assignedValues = new Set(
            categoryMappingRows
                .filter((_, i) => i !== index)
                .map((row) => row.providerCategory)
                .filter(Boolean),
        );
        return providerCategorySelectItems.filter(
            (item) => item.value === currentValue || !assignedValues.has(item.value),
        );
    }

    function booleanDefault(value: unknown, fallback: boolean): boolean {
        if (typeof value === "boolean") return value;
        if (typeof value === "string") return value.trim().toLowerCase() === "true";
        return fallback;
    }

    function selectItems(field: ConfigField): SelectItem[] {
        return (field.options ?? []).map((o) => ({
            text: configFieldOptionLabel(o, $locale, $_(o.value)),
            value: o.value,
        }));
    }

    function fieldLabel(field: ConfigField): string {
        return configFieldLabel(field, $locale, $_(configLabels[field.key] ?? field.key));
    }

    function fieldHint(field: ConfigField): string | undefined {
        const description = configFieldDescription(field, $locale);
        if (description) return description;
        const hint = configHints[field.key];
        return hint ? $_(hint) : undefined;
    }

    function fieldError(field: ConfigField): string {
        return configErrors[field.key] ?? "";
    }

    async function addCategoryMappingRow() {
        const newIndex = categoryMappingRows.length;
        categoryMappingRows = [
            ...categoryMappingRows,
            {
                providerCategory: "",
                category: "",
            },
        ];
        await tick();
        categoryMappingList?.querySelector<HTMLElement>(`[data-category-mapping-row="${newIndex}"]`)
            ?.scrollIntoView({ block: "nearest", behavior: "smooth" });
    }

    function removeCategoryMappingRow(index: number) {
        categoryMappingRows = categoryMappingRows.filter((_, i) => i !== index);
    }

    function validateConfig(): boolean {
        if (resolvedAuth.error) {
            show_toast({
                text: $_("error-setting-up-plugin", { values: { provider: pluginTitle() } }),
                icon: "close",
                type: "error",
            });
            return false;
        }
        const errors: Record<string, string> = {};
        const hiddenMissing: ConfigField[] = [];
        for (const field of configSchema) {
            if (!field.required) continue;
            const value = extraConfig[field.key];
            if (value === undefined || value === null || value === "") {
                if (field.hidden) {
                    hiddenMissing.push(field);
                    continue;
                }
                errors[field.key] = $_("required");
            }
        }
        configErrors = errors;
        if (hiddenMissing.length > 0) {
            show_toast({
                text: `${hiddenMissing.map((field) => fieldLabel(field)).join(", ")}: ${$_("required")}`,
                icon: "close",
                type: "error",
            });
        }
        return Object.keys(errors).length === 0 && hiddenMissing.length === 0;
    }

    function authLabel(field: string): string {
        return authLabels[field] ?? field;
    }

    function authFieldType(field: string): "password" | "text" {
        return secretFields.has(field) ? "password" : "text";
    }

    function configSection(key: string): Record<string, any> {
        const section = draftInstance?.config?.[key];
        if (section && typeof section === "object" && !Array.isArray(section)) {
            return section as Record<string, any>;
        }
        return {};
    }

    function pluginConfig(): Record<string, any> {
        return configSection("plugin");
    }

    function hostConfig(): Record<string, any> {
        return mergePluginConfig(plugin.hostConfig ?? {}, configSection("host"));
    }

    function authChanged() {
        if (authContext !== (draftInstance?.auth_context ?? (!draftInstance ? plugin.authContext : "") ?? "")) return true;
        if (Object.values(clearedAuth).some(Boolean)) return true;
        for (const field of authFields) {
            const value = auth[field] ?? "";
            if (secretFields.has(field)) {
                if (value !== "") {
                    return true;
                }
                continue;
            }
            if (value !== ((draftInstance?.auth?.[field] as string | undefined) ?? "")) {
                return true;
            }
        }
        return false;
    }

    export function openModal() {
        openingInstance = instance ? JSON.parse(JSON.stringify(instance)) : undefined;
        snapshotTaken = true;
        authContext = draftInstance?.auth_context ?? (!draftInstance ? plugin.authContext : "") ?? "";
        clearedAuth = {};
        auth = initialAuth();
        const config = hostConfig();
        planned = supportsPlanned && (!hasTourKindChoice || ((config.planned as boolean | undefined) ?? true));
        completed =
            supportsCompleted && (!hasTourKindChoice || ((config.completed as boolean | undefined) ?? true));
        mergeEnabled = mergeAvailable && Boolean((config.merge as any)?.enabled);
        privacy = (config.privacy as string | undefined) ?? "original";
        extraConfig = initialExtraConfig();
        categoryMappingRows = initialCategoryMappingRows();
        configErrors = {};
        openingConfig = formConfig();
        modal.openModal();
    }

    function formConfig(): Record<string, unknown> {
        const pluginRuntimeConfig: Record<string, unknown> = { ...pluginConfig() };
        const pluginHostConfig: Record<string, unknown> = { ...hostConfig() };
        if (supportsPlanned) {
            pluginHostConfig.planned = hasTourKindChoice ? planned : true;
        } else {
            delete pluginHostConfig.planned;
        }
        if (supportsCompleted) {
            pluginHostConfig.completed = hasTourKindChoice ? completed : true;
        } else {
            delete pluginHostConfig.completed;
        }
        if (supportsSourcePrivacy) {
            pluginHostConfig.privacy = privacy;
        }
        const currentMergeConfig =
            pluginHostConfig.merge && typeof pluginHostConfig.merge === "object"
                ? (pluginHostConfig.merge as Record<string, unknown>)
                : {};
        pluginHostConfig.merge = {
            ...currentMergeConfig,
            enabled: mergeAvailable && mergeEnabled,
        };
        const categoryMappingConfig: Record<string, CategoryMappingTarget> = {};
        const assignedProviderCategories = new Set<string>();
        for (const row of categoryMappingRows) {
            const providerCategory = row.providerCategory.trim();
            if (!providerCategory || !row.category) {
                continue;
            }
            assignedProviderCategories.add(providerCategory);
            categoryMappingConfig[providerCategory] = categoryMappingTargetFromPickerValue(row.category);
        }
        for (const providerCategory of [
            ...Object.keys(manifestCategoryMapping()),
            ...Object.keys(categoryMapping()),
        ]) {
            if (!assignedProviderCategories.has(providerCategory)) {
                categoryMappingConfig[providerCategory] = "";
            }
        }
        pluginHostConfig.categoryMapping = categoryMappingConfig;
        for (const field of configSchema) {
            const val = extraConfig[field.key];
            if (val !== undefined && val !== "") {
                pluginRuntimeConfig[field.key] = val;
            } else {
                delete pluginRuntimeConfig[field.key];
            }
        }
        const config: Record<string, unknown> = {
            plugin: pluginRuntimeConfig,
            host: pluginHostConfig,
        };

        return config;
    }

    function pluginInstanceFromForm(): PluginInstanceForm {
        const candidate: PluginInstanceForm = { id: draftInstance?.id, plugin_id };
        const submittedAuth: Record<string, string | null> = {};
        for (const field of authFields) {
            const value = auth[field] ?? "";
            if (clearedAuth[field]) submittedAuth[field] = null;
            else if (secretFields.has(field)) {
                if (value) submittedAuth[field] = value;
            } else if (!draftInstance || value !== (draftInstance.auth?.[field] ?? "")) {
                submittedAuth[field] = value;
            }
        }
        if (Object.keys(submittedAuth).length) candidate.auth = submittedAuth;
        if (authContext && authContext !== draftInstance?.auth_context) candidate.authContext = authContext;
        if (draftInstance?.id && authChanged()) candidate.expected_connection_generation = draftInstance.connection_generation;
        const fullConfig = formConfig();
        const patch = draftInstance ? pluginSettingsConfigPatch(openingConfig, fullConfig) : fullConfig;
        if (Object.keys(patch).length) candidate.config = patch;
        const hostPatch = patch.host as Record<string, unknown> | undefined;
        candidate.mappingChanged = !!hostPatch && Object.hasOwn(hostPatch, "categoryMapping");
        if (candidate.mappingChanged && hostPatch) hostPatch.categoryMappingUpdatedAt = new Date().toISOString();
        return candidate;
    }

    async function reloadAfterConflict(error: unknown) {
        if (draftInstance?.id && pluginOperationNeedsReload(error)) {
            try { await plugin_instances_show(draftInstance.id); } catch { /* Keep the action error. */ }
        }
    }

    async function submit(validate = false) {
        const candidate = pluginInstanceFromForm();
        let persistenceStarted = false;
        if ((!draftInstance || hasProviderConfigChanges(candidate.config) || validate) && !validateConfig()) return;
        try {
            isSaving = true;
            if (validate && isSessionPlugin) {
                await plugin_auth_validate({
                    pluginId: plugin.id, instanceId: draftInstance?.id, authContext: resolvedAuth.name,
                    auth: candidate.auth ?? {},
                    config: candidate.config,
                    expected_connection_generation: draftInstance?.connection_generation,
                });
            }
            if (candidate.mappingChanged) {
                const shouldSave = await onbeforecategorymappingsave?.(candidate);
                if (shouldSave === false) return;
            }
            delete candidate.mappingChanged;
            persistenceStarted = true;
            await onsave?.(candidate);
            modal.closeModal();
        } catch (e) {
            await reloadAfterConflict(e);
            // Persistence errors are reported by onsave; validation errors are
            // reported here before persistence has been attempted.
            if (validate && !persistenceStarted) show_toast({
                text: translatePluginAPIError(e, $_("error-setting-up-plugin", { values: { provider: pluginTitle() } })),
                icon: "close", type: "error",
            });
        } finally {
            isSaving = false;
            if (draftInstance?.id) await plugin_instances_show(draftInstance.id).catch(() => {});
        }
    }

    function pluginTitle() {
        return localizedPluginTitle(plugin, $locale);
    }

    async function disconnect() {
        if (!draftInstance?.id || !draftInstance.connection_generation) return;
        isSaving = true;
        try {
            await plugin_oauth_revoke(draftInstance.id, draftInstance.connection_generation);
            modal.closeModal();
        } catch (e) {
            await reloadAfterConflict(e);
            show_toast({ text: translatePluginAPIError(e, $_("error-completing-oauth")), icon: "close", type: "error" });
        } finally { isSaving = false; }
    }

    async function startOAuth() {
        const candidate = pluginInstanceFromForm();
        const connectionChanges = authChanged();
        const repairUnreadableGrant = draftInstance?.last_error?.code === "credentials_unreadable" &&
            authContext === draftInstance.auth_context && isConnected;
        if ((!draftInstance || hasProviderConfigChanges(candidate.config)) && !validateConfig()) return;
        isSaving = true;
        try {
            delete candidate.mappingChanged;
            const saved = await onsave?.(candidate);
            const instanceId = saved?.id ?? draftInstance?.id;
            let expectedGeneration = !draftInstance || connectionChanges
                ? saved?.connection_generation : draftInstance.connection_generation;
            if (!instanceId || !expectedGeneration) throw new Error("Missing saved connection revision");
            if (repairUnreadableGrant) {
                const revoked = await plugin_oauth_revoke(instanceId, expectedGeneration);
                expectedGeneration = revoked.connection_generation;
                if (!expectedGeneration) throw new Error("Missing revoked connection revision");
            }
            const redirectUri = `${window.location.origin}/settings/plugins/oauth/callback`;
            const result = await plugin_oauth_start({
                pluginId: plugin.id, instanceId, authContext: resolvedAuth.name, redirectUri,
                expected_connection_generation: expectedGeneration, activate_on_success: true,
            });
            sessionStorage.setItem(`wanderer_plugin_oauth_${result.state}`, result.instanceId);
            window.location.href = result.url;
        } catch (e) {
            await reloadAfterConflict(e);
            show_toast({ text: translatePluginAPIError(e, $_("error-starting-oauth")), icon: "close", type: "error" });
        } finally { isSaving = false; }
    }

</script>

<Modal
    id="{plugin_id}-plugin-settings-modal"
    size="md:min-w-xl lg:min-w-2xl"
    title={pluginTitle() + " " + $_("settings")}
    bind:this={modal}
>
    {#snippet content()}
        <form
            id="{plugin_id}-plugin-settings-form"
            class="space-y-3"
            novalidate
            onsubmit={(event) => {
                event.preventDefault();
                submit();
            }}
        >
            {#if draftStale}
                <p class="text-sm text-red-400">{$_("plugin-error-connection-conflict")}</p>
                <button class="btn-secondary" type="button" onclick={openModal} disabled={isSaving}
                    >{$_("plugin-reload-settings")}</button>
            {/if}
            {#if Object.keys(plugin.authContexts ?? {}).length > 1 || resolvedAuth.error}
                <Select name="pluginAuthContext" label={$_("plugin-auth-context")} items={[
                    { text: $_("select"), value: "" },
                    ...Object.keys(plugin.authContexts ?? {}).map(name => ({ text: name, value: name })),
                ]} bind:value={authContext} />
            {/if}
            {#if draftInstance?.last_error?.code === "credentials_unreadable"}
                <p class="text-sm text-red-400">{$_("plugin-error-credentials-unreadable")}</p>
            {/if}
            {#each authFields as field}
                <TextField
                    label={authLabel(field)}
                    placeholder={secretFields.has(field)
                        ? draftInstance?.id
                            ? `(${$_("unchanged")})`
                            : ""
                        : field == "email"
                          ? "user@example.com"
                          : ""}
                    bind:value={auth[field]}
                    name={field}
                    type={authFieldType(field)}
                ></TextField>
                {#if draftInstance && secretFields.has(field) && Object.hasOwn(draftInstance.auth ?? {}, field)}
                    <label class="flex items-center gap-2 text-xs text-gray-500">
                        <input name="clear-{field}" type="checkbox" bind:checked={clearedAuth[field]} />
                        {$_("plugin-clear-credential", { values: { field: authLabel(field) } })}
                    </label>
                {/if}
            {/each}

            {#if isOAuthPlugin}
                <p class="text-xs text-gray-500 max-w-lg">
                    {#if isConnected}
                        {$_("plugin-oauth-material-present-hint")}
                    {:else}
                        {$_("plugin-oauth-needs-connect-hint")}
                    {/if}
                </p>
            {/if}

            {#if hasTourKindChoice}
                <div class="flex flex-wrap gap-x-4">
                    <Toggle
                        bind:value={planned}
                        label={$_("planned-tours", { values: { n: 2 } })}
                    ></Toggle>
                    <Toggle
                        bind:value={completed}
                        label={$_("completed-tours", { values: { n: 2 } })}
                    ></Toggle>
                </div>
            {/if}

            {#if supportsSourcePrivacy}
                <Select
                    label={$_("privacy")}
                    items={privacySelectItems}
                    bind:value={privacy}
                ></Select>
                <p class="text-xs text-gray-500 max-w-lg">
                    {#if privacy == "original"}
                        {$_("plugin-privacy-hint-original")}
                    {:else}
                        {$_("plugin-privacy-hint-user")}
                    {/if}
                </p>
            {/if}

            {#each visibleConfigSchema as field}
                {#if field.type === "select"}
                    <Select
                        label={fieldLabel(field)}
                        items={selectItems(field)}
                        bind:value={extraConfig[field.key] as string}
                        error={fieldError(field)}
                    ></Select>
                {:else if field.type === "boolean"}
                    <Toggle
                        bind:value={extraConfig[field.key]}
                        label={fieldLabel(field)}
                        error={fieldError(field)}
                    ></Toggle>
                    {@const hint = fieldHint(field)}
                    {#if hint}
                        <p class="text-xs text-gray-500 max-w-lg">{hint}</p>
                    {/if}
                {:else if field.type === "date"}
                    {@const hint = fieldHint(field)}
                    {#if hint}
                        <p
                            class="text-xs text-gray-500 max-w-lg pt-4 pb-1 border-t border-input-border"
                        >
                            {hint}
                        </p>
                    {/if}
                    <div class="flex items-end relative gap-x-2">
                        <Datepicker
                            label={fieldLabel(field)}
                            bind:value={extraConfig[field.key]}
                            error={fieldError(field)}
                        ></Datepicker>
                        <button
                            class="btn-icon mb-[10px]"
                            type="button"
                            onclick={() => {
                                extraConfig[field.key] = undefined;
                            }}
                            aria-label={$_("clear")}
                        ><i class="fa fa-close"></i></button>
                    </div>
                {:else if field.type === "text" || field.type === "url"}
                    <TextField
                        label={fieldLabel(field)}
                        bind:value={extraConfig[field.key]}
                        name={field.key}
                        type={field.type === "url" ? "url" : "text"}
                        error={fieldError(field)}
                    ></TextField>
                {/if}
            {/each}

            {#if supportsCategoryMapping && categories.length > 0}
                <div class="space-y-2 pt-4 border-t border-input-border">
                    <div class="flex items-center justify-between gap-3">
                        <div>
                            <h4 class="text-sm font-medium">{$_("category-mapping")}</h4>
                            <p class="text-xs text-secondary mt-1">
                                {$_("category-mapping-help")}
                            </p>
                        </div>
                        <button
                            class="btn-primary text-sm"
                            class:btn-disabled={!canAddCategoryMappingRow}
                            type="button"
                            disabled={!canAddCategoryMappingRow}
                            onclick={addCategoryMappingRow}
                        >{$_("add-entry")}</button>
                    </div>
                    {#if categoryMappingRows.length > 0}
                        <div class="hidden md:grid grid-cols-[minmax(0,1.2fr)_minmax(14rem,1fr)_2.75rem] gap-3 text-sm font-medium">
                            <span>{$_("provider-category")}</span>
                            <span>{$_("category")} / {$_("subcategory")}</span>
                            <span></span>
                        </div>
                        <div
                            bind:this={categoryMappingList}
                            class="space-y-3 pr-1 scroll-smooth"
                            class:max-h-[300px]={categoryMappingRows.length > 6}
                            class:overflow-y-auto={categoryMappingRows.length > 6}
                            class:overflow-y-visible={categoryMappingRows.length <= 6}
                        >
                            {#each categoryMappingRows as row, i}
                                {@const providerItems = providerCategoryItemsForRow(i)}
                                <div
                                    class="grid grid-cols-1 md:grid-cols-[minmax(0,1.2fr)_minmax(14rem,1fr)_2.75rem] items-end gap-3"
                                    data-category-mapping-row={i}
                                >
                                    <SingleSelect
                                        ariaLabel={$_("provider-category")}
                                        placeholder={$_("select-provider-category")}
                                        items={providerItems}
                                        bind:value={row.providerCategory}
                                        disabled={row.providerCategory !== "" && providerItems.length <= 1}
                                    ></SingleSelect>
                                    <CategoryPicker
                                        value={row.category}
                                        label=""
                                        placeholder={$_("select-category")}
                                        currentCategoryId={categoryPickerCurrentCategoryId(row.category)}
                                        fixedDropdown
                                        onchange={(selection) => {
                                            row.category = selection.subcategory
                                                ? `subcategory:${selection.subcategory}`
                                                : `category:${selection.category}`;
                                        }}
                                    ></CategoryPicker>
                                    <button
                                        class="btn-icon h-10"
                                        type="button"
                                        onclick={() => removeCategoryMappingRow(i)}
                                        aria-label={$_("remove")}
                                    ><i class="fa fa-close"></i></button>
                                </div>
                            {/each}
                        </div>
                    {/if}
                </div>
            {/if}

            {#if mergeAvailable}
                <PluginMergeSettings prefix="merge" bind:value={mergeEnabled} />
            {/if}
        </form>
    {/snippet}
    {#snippet footer()}
        <div class="flex items-center gap-4">
            <button class="btn-secondary" onclick={() => modal.closeModal()} disabled={isSaving}
                >{$_("cancel")}</button
            >
            {#if isOAuthPlugin}
                <button class="btn-secondary" type="button" onclick={startOAuth} disabled={isSaving}
                    >{isConnected ? $_("save-and-reconnect") : $_("save-and-connect")}</button>
                {#if isConnected && authContext === draftInstance?.auth_context}
                    <button class="btn-secondary" type="button" onclick={disconnect} disabled={isSaving}
                        >{$_("plugin-disconnect")}</button>
                {/if}
                <button class="btn-primary" form="{plugin_id}-plugin-settings-form" type="submit"
                    name="save" disabled={isSaving}>{$_("save")}</button>
            {:else}
                {#if isSessionPlugin}
                    <button class="btn-secondary" type="button" name="save-validate" onclick={() => submit(true)} disabled={isSaving}
                        >{$_("save-and-validate")}</button>
                {/if}
                <button class="btn-primary" form="{plugin_id}-plugin-settings-form" type="submit"
                    name="save" disabled={isSaving}>{$_("save")}</button>
            {/if}

        </div>
    {/snippet}
</Modal>
