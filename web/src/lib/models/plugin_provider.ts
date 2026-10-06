export type LocalizedTextMap = Record<string, string>;

export interface ConfigFieldOption {
    value: string;
    label?: string;
    labels?: LocalizedTextMap;
}

export interface ConfigField {
    key: string;
    type: "boolean" | "date" | "select" | "text" | "url";
    label?: string;
    labels?: LocalizedTextMap;
    description?: string;
    descriptions?: LocalizedTextMap;
    options?: ConfigFieldOption[];
    default?: unknown;
    required?: boolean;
    hidden?: boolean;
}

export interface PluginAuthContext {
    type: string;
    fields?: string[];
    secretFields?: string[];
    secretField?: string;
    authorizationUrl?: string;
    tokenUrl?: string;
    tokenRequestFormat?: "json" | "form";
    scopes?: string[];
    scopeSeparator?: string;
    authorizationParams?: Record<string, string>;
    pkce?: boolean;
    tokenAuth?: string;
}

export interface PluginProvider {
    id: string;
    type: "trails";
    name: string;
    displayName?: string;
    displayNames?: LocalizedTextMap;
    description?: string;
    descriptions?: LocalizedTextMap;
    information?: LocalizedTextMap;
    homepageUrl?: string;
    donationUrl?: string;
    icon?: string;
    iconDark?: string;
    version?: string;
    protocolVersion?: string;
    risk?: string;
    auth: PluginAuthContext;
    authContext?: string;
    authContextError?: string;
    authContexts?: Record<string, PluginAuthContext>;
    configSchema?: ConfigField[];
    hostConfig?: Record<string, unknown>;
    metadata?: Record<string, unknown>;
    capabilities?: string[];
    limits?: {
        recommendedBatchSize?: number;
    };
    status: "available" | "disabled" | "error";
    setupErrorCode?: string;
    error?: string;
}
