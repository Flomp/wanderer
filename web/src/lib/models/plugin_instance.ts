export interface PluginInstance {
    id?: string;
    user: string;
    plugin_id: string;
    enabled: boolean;
    auth?: Record<string, string>;
    auth_context?: string;
    connection_generation?: string;
    busy?: boolean;
    config?: Record<string, unknown>;
    state?: Record<string, unknown>;
    status:
        | "configured"
        | "needs_auth"
        | "needs_reauth"
        | "rate_limited"
        | "unavailable"
        | "unsupported_protocol"
        | "error"
        | "disabled";
    last_error?: {
        code?: string;
        message?: string;
        authContext?: string;
        connectionGeneration?: string;
    };
    last_sync_at?: string;
    retry_not_before?: string;
}

// Host-owned fields and OAuth metadata never belong in a settings request.
export interface PluginInstanceSettings {
    enabled?: boolean;
    auth?: Record<string, string | null>;
    authContext?: string;
    config?: Record<string, unknown>;
    expected_connection_generation?: string;
}

export interface PluginInstanceDraft extends PluginInstanceSettings {
    id?: string;
    plugin_id: string;
}
