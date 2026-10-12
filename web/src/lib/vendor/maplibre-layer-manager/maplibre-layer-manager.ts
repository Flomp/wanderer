import * as M from "maplibre-gl";
import { DebugLayer } from "./debug-layer";
import { baseMapStyles, defaultMapState, type BaseLayer, type MapState } from "./layers";
import { OverlayLayer } from "./overlay-layer";
import { OverpassLayer, type OverpassPopupActionFactory } from "./overpass-layer";

const DEFAULT_GLYPHS = "https://tiles.openfreemap.org/fonts/{fontstack}/{range}.pbf";

export type ViewerPrefsScopeKind = 'account' | 'trail' | 'list';

export interface ViewerPrefsScope {
    getUrl: string;
    putUrl: string;
    kind?: ViewerPrefsScopeKind;
}

export function viewerPrefsScope(trailIds: Array<string | undefined>, listId?: string | null): ViewerPrefsScope {
    const ids = trailIds.filter((id): id is string => !!id);
    if (ids.length === 1) {
        const trailUrl = `/map-viewer-prefs/trail/${ids[0]}`;
        if (listId) {
            return { getUrl: `${trailUrl}?list=${listId}`, putUrl: trailUrl, kind: 'trail' };
        }
        return { getUrl: trailUrl, putUrl: trailUrl, kind: 'trail' };
    }
    if (listId) {
        const listUrl = `/map-viewer-prefs/list/${listId}`;
        return { getUrl: listUrl, putUrl: listUrl, kind: 'list' };
    }
    return { ...ACCOUNT_SCOPE };
}

function inferScopeKind(scope: ViewerPrefsScope): ViewerPrefsScopeKind {
    if (scope.kind) {
        return scope.kind;
    }
    if (scope.putUrl === ACCOUNT_SCOPE.putUrl) {
        return 'account';
    }
    if (scope.putUrl.startsWith('/map-viewer-prefs/trail/')) {
        return 'trail';
    }
    if (scope.putUrl.startsWith('/map-viewer-prefs/list/')) {
        return 'list';
    }
    return 'account';
}

const ACCOUNT_SCOPE: ViewerPrefsScope = {
    getUrl: '/map-viewer-prefs/account',
    putUrl: '/map-viewer-prefs/account'
};

function hasStoredPrefs(state: unknown): boolean {
    if (!state || typeof state !== 'object') {
        return false;
    }
    const prefs = state as Partial<MapState>;
    return prefs.base !== undefined || prefs.overlays !== undefined || prefs.pois !== undefined;
}

export class LayerManager {
    private map: M.Map;
    state!: MapState;
    layers: Record<string, BaseLayer> = {};
    private addedListeners: Set<string> = new Set();
    private overpassActionFactory?: OverpassPopupActionFactory;
    private scope: ViewerPrefsScope = ACCOUNT_SCOPE;
    private prefsGeneration = 0;

    constructor(map: M.Map, options?: { overpassActionFactory?: OverpassPopupActionFactory, scope?: ViewerPrefsScope }) {
        this.map = map;
        this.overpassActionFactory = options?.overpassActionFactory;
        if (options?.scope) {
            this.scope = options.scope;
        }

        const storedMapState = localStorage.getItem("map-state")
        if (storedMapState) {
            try {
                this.state = JSON.parse(storedMapState)
            } catch {
                this.state = defaultMapState
            }

            // Fix for spelling correction introduced in v0.18.4
            if (this.state && 'ammenity' in this.state.pois) {
                this.state.pois.amenity = this.state.pois.ammenity;
                delete this.state.pois.ammenity;

                localStorage.setItem('map-state', JSON.stringify(this.state));
            }
        } else {
            this.state = defaultMapState
        }
    }

    init() {
        this.map.on('moveend', () => { });

        this.map.on('styledata', () => {
            this.restoreLayers();
        })
        try {
            this.update(this.state, true);

            const overpassLayer = new OverpassLayer(this.map, this.overpassActionFactory)
            const debugLayer = new DebugLayer()

            this.addLayer("overpass", overpassLayer)
            this.addLayer("debug", debugLayer)

            this.map.on('moveend', this.updateOverpassLayerAfterMapMoveBinded);
        } catch (e) {
            console.error(e)
            // map is probably not initialized yet
        }
    }

    async update(newState: MapState, initialize: boolean = false): Promise<void> {
        if (!initialize) {
            this.prefsGeneration++;
        }
        const oldState = this.state;

        if (oldState.base != newState.base || initialize) {
            this.updateBaseLayer(baseMapStyles[newState.base])
        }

        for (const [name, active] of Object.entries(newState.overlays)) {
            const oldOverlayActive = oldState.overlays[name]

            if (active && (!oldOverlayActive || initialize)) {
                this.addLayer(name, new OverlayLayer(name))
            } else if (oldOverlayActive && !active) {
                this.removeLayer(name)
            }
        }

        this.updateOverpassLayer(newState);

        this.state = newState
        localStorage.setItem("map-state", JSON.stringify(this.state));

        if (!initialize) {
            try {
                await fetch(this.scope.putUrl, {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    credentials: 'same-origin',
                    body: JSON.stringify(newState)
                });
            } catch {
                // Local state is already updated; a failed save retries on the next change.
            }
        }
    }

    async refresh(): Promise<void> {
        const generation = ++this.prefsGeneration;
        const requestUrl = this.scope.getUrl;
        let response: Response;
        try {
            response = await fetch(requestUrl, { credentials: 'same-origin' });
        } catch {
            // Network failure: stay local; a later refresh retries.
            return;
        }
        if (!response.ok) {
            // Anonymous or denied: browser-local behavior.
            return;
        }
        let server: Partial<MapState> | null = null;
        try {
            const parsed: unknown = await response.json();
            if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
                server = parsed as Partial<MapState>;
            }
            // An unreadable or misshapen body falls through as a missing scope.
        } catch {
            // Unreadable body falls through as a missing scope.
        }
        if (generation !== this.prefsGeneration || requestUrl !== this.scope.getUrl) {
            return;
        }
        if (!hasStoredPrefs(server)) {
            // A scope with nothing saved opens from viewer defaults. Browser
            // values are never read back and the scope stays unsaved until its
            // next save; the assembled prefs still refresh the browser cache.
            this.update(defaultMapState, true);
            return;
        }
        const merged: MapState = structuredClone(defaultMapState);
        if (server.base !== undefined) merged.base = server.base;
        Object.assign(merged.overlays, server.overlays ?? {});
        const mergedPois = merged.pois as Record<string, Record<string, boolean>>;
        for (const [category, toggles] of Object.entries(server.pois ?? {})) {
            Object.assign(mergedPois[category] ??= {}, toggles);
        }
        this.update(merged, true);
    }

    updateOverpassLayerAfterMapMoveBinded = this.updateOverpassLayerAfterMapMove.bind(this);
    updateOverpassLayerAfterMapMove() {
        this.updateOverpassLayer(this.state)
    }

    private async updateOverpassLayer(newState: MapState) {
        const overpassLayer = this.layers.overpass;
        if (overpassLayer && this.map.getSource('overpass')) {
            const castedOverpassLayer = overpassLayer as OverpassLayer;
            overpassLayer.filter = await castedOverpassLayer.updateLayerIfNeeded(newState, this.map.getBounds());
            (this.map.getSource('overpass') as M.GeoJSONSource).setData(castedOverpassLayer.data);
        }
    }

    private updateBaseLayer(layer: string | M.StyleSpecification) {
        this.map.setStyle(layer);
    }


    addLayer(id: string, layer: BaseLayer) {
        if (!layer.spec) {
            return;
        }
        if (this.layers[id] && this.map.getLayer(id)) {
            // update sources and return
            for (const [sourceId, s] of Object.entries(layer.spec.sources)) {
                if (s.type != "geojson" || !this.map.getSource(sourceId)) {
                    continue;
                }

                const source = this.map.getSource(sourceId) as M.GeoJSONSource
                source.setData(s.data)
            }
            this.layers[id] = layer
            return;
        }
        for (const [id, s] of Object.entries(layer.spec.sources)) {
            if (!this.map.getSource(id)) {
                this.map.addSource(id, s)
            }
        }

        const style = this.map.getStyle();
        if (
            !style.glyphs &&
            layer.spec.layers.some((l) => l.type === "symbol" && l.layout && "text-field" in l.layout)
        ) {
            this.map.setGlyphs(layer.spec.glyphs ?? DEFAULT_GLYPHS);
        }

        for (const l of layer.spec.layers) {
            if (!this.map.getLayer(l.id)) {
                this.map.addLayer(l)
            }
        }

        if (layer.listeners) {
            for (const [id, listener] of Object.entries(layer.listeners)) {
                if (listener.onEnter && !this.addedListeners.has("mouseenter-" + id)) {
                    this.addedListeners.add("mouseenter-" + id)
                    this.map.on('mouseenter', id, listener.onEnter);
                }

                if (listener.onLeave && !this.addedListeners.has("onleave-" + id)) {
                    this.addedListeners.add("mouseleave-" + id)
                    this.map.on('mouseleave', id, listener.onLeave);
                }

                if (listener.onMouseDown && !this.addedListeners.has("mousedown-" + id)) {
                    this.addedListeners.add("mousedown-" + id)
                    this.map.on('mousedown', id, listener.onMouseDown);
                }

                if (listener.onMouseUp && !this.addedListeners.has("mouseup-" + id)) {
                    this.addedListeners.add("mouseup-" + id)
                    this.map.on('mouseup', id, listener.onMouseUp);
                }

                if (listener.onMouseMove && !this.addedListeners.has("mousemove-" + id)) {
                    this.addedListeners.add("mousemove-" + id)
                    this.map.on('mousemove', id, listener.onMouseMove);
                }
            }
        }

        if (layer.filter && !this.map.getFilter(id)) {
            this.map.setFilter(id, layer.filter)
        }

        this.layers[id] = layer
    }

    removeLayer(id: string) {
        const layer = this.layers[id];
        if (!layer) {
            return;
        }
        for (const l of layer.spec.layers) {
            if (this.map.getLayer(l.id)) {
                this.map.removeLayer(l.id)
            }
        }

        for (const [id, _] of Object.entries(layer.spec.sources)) {
            if (this.map.getSource(id)) {
                this.map.removeSource(id)
            }
        }

        if (layer.listeners) {
            for (const [id, listener] of Object.entries(layer.listeners)) {
                if (listener.onEnter) {
                    this.addedListeners.delete("mouseenter-" + id)
                    this.map.off('mouseenter', id, listener.onEnter);
                }

                if (listener.onLeave) {
                    this.addedListeners.delete("mouseleave-" + id)
                    this.map.off('mouseleave', id, listener.onLeave);
                }
                if (listener.onMouseDown) {
                    this.addedListeners.delete("click-" + id)
                    this.map.off('click', id, listener.onMouseDown);
                }
                if (listener.onMouseMove) {
                    this.addedListeners.delete("mousemove-" + id)
                    this.map.off('mousemove', id, listener.onMouseMove);
                }
            }
        }

        delete this.layers[id]

    }

    private restoreLayers() {
        for (const [id, layer] of Object.entries(this.layers)) {
            this.addLayer(id, layer)
        }
    }
}
