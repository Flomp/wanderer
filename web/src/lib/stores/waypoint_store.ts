import { Waypoint } from "$lib/models/waypoint";
import { APIError } from "$lib/util/api_util";
import { get, writable, type Writable } from "svelte/store";
import { currentUser } from "./user_store";
import type { AuthRecord } from "pocketbase";
import { assets_attach_to_target, assets_delete_removed, has_asset_attachments, type CompletedAssetAttachments, type AssetImportOmissionHandler } from "./asset_store";
import { assetIdFromPhotoURL, enrichWaypointAssetExpands, shownAssetLinks } from "$lib/util/asset_link_util";

export const waypoint: Writable<Waypoint> = writable(new Waypoint(0, 0));

export async function waypoints_create(waypoint: Waypoint, f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch, user?: AuthRecord, onOmitted?: AssetImportOmissionHandler) {
    user ??= get(currentUser)
    if (!user) {
        throw Error("Unauthenticated")
    }

    waypoint.author = user.actor

    let r = await f('/api/v1/waypoint', {
        method: 'PUT',
        body: JSON.stringify({ ...waypoint, photos: [], _photos: undefined, _assetCandidates: undefined, _assetLinks: undefined, _assetPluginLinks: undefined }),
    })

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }


    let model: Waypoint = await r.json();
    const attachments = {
        files: waypoint._photos,
        assetIds: shownAssetLinks(waypoint._assetLinks, waypoint.photos),
        pluginLinks: waypoint._assetPluginLinks,
        target: {
            trail: model.trail ?? waypoint.trail,
            waypoint: model.id,
        },
        existingPhotos: model.photos,
        onOmitted,
        f,
    };
    try {
        model.photos = await assets_attach_to_target(attachments);
        if (has_asset_attachments(attachments) && !model.photos.length) {
            throw new APIError(422, "asset_import_failed");
        }
    } catch (importError) {
        // An earlier attachment request may already have saved a photo. Only
        // remove a newly created waypoint after confirming it has no links.
        if (importError instanceof APIError && importError.detail?.attachedPhotos) {
            model.photos = importError.detail.attachedPhotos;
        }
        const remainingAttachments = remainingWaypointAttachments(waypoint, importError);
        let savedWaypoint: Waypoint | undefined;
        try {
            savedWaypoint = await removeEmptyCreatedWaypoint(model, f);
        } catch (cleanupError) {
            throw new APIError(422, "asset_import_cleanup_failed", {
                importError, cleanupError, savedWaypoint: model, remainingAttachments,
            });
        }
        if (savedWaypoint) {
            throw new APIError(422, "asset_import_partial_failure", {
                importError, savedWaypoint, remainingAttachments,
            });
        }
        throw new APIError(422, "asset_import_failed", { importError });
    }

    return model;

}

// The attached photos start with those the form already had. The ones linked
// by ID are saved only once linking completed: listed earlier, dropping one
// before the retry would send a removal, and that deletes a library photo
// without other links.
function savedAttachedPhotos(attachedPhotos: string[], waypoint: Waypoint, error: unknown): string[] {
    const linked = error instanceof APIError && error.detail?.completedAttachments?.assetIds;
    const unlinked = new Set(linked ? [] : waypoint._assetLinks);
    return attachedPhotos.filter((photo) => !unlinked.has(assetIdFromPhotoURL(photo) ?? ""));
}

function remainingWaypointAttachments(waypoint: Waypoint, error: unknown) {
    const completed: CompletedAssetAttachments | undefined = error instanceof APIError
        ? error.detail?.completedAttachments
        : undefined;
    return {
        _photos: completed?.files ? undefined : waypoint._photos,
        _assetLinks: completed?.assetIds ? undefined : waypoint._assetLinks,
        _assetPluginLinks: waypoint._assetPluginLinks?.filter((link) => !completed?.pluginLinks.includes(link)),
    };
}

async function removeEmptyCreatedWaypoint(
    model: Waypoint,
    f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response>,
): Promise<Waypoint | undefined> {
    const url = `/api/v1/waypoint/${model.id}`;
    const response = await f(`${url}?expand=waypoint_assets_via_waypoint.asset`);
    if (response.status === 404) {
        return undefined;
    }
    if (!response.ok) {
        throw new APIError(response.status, "Could not check saved waypoint photos");
    }
    const saved: Waypoint = await response.json();
    const links = saved.expand?.waypoint_assets_via_waypoint;
    const hasPhotos = Boolean(saved.photos?.length || saved.expand?.assets_via_waypoint?.length);
    enrichWaypointAssetExpands(saved);
    if (links?.length || hasPhotos) {
        return saved;
    }
    const removed = await f(url, { method: "DELETE" });
    if (!removed.ok && removed.status !== 404) {
        throw new APIError(removed.status, "Could not remove empty photo waypoint");
    }
    return undefined;
}

export async function waypoints_update(oldWaypoint: Waypoint, newWaypoint: Waypoint, onOmitted?: AssetImportOmissionHandler) {
    const user = get(currentUser)
    if (!user) {
        throw Error("Unauthenticated")
    }
    newWaypoint.author = user.id

    let r = await fetch('/api/v1/waypoint/' + newWaypoint.id, {
        method: 'POST',
        body: JSON.stringify({ ...newWaypoint, photos: [], _photos: undefined, _assetCandidates: undefined, _assetLinks: undefined, _assetPluginLinks: undefined }),
    })

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }

    const model: Waypoint = await r.json();
    try {
        model.photos = await assets_attach_to_target({
            files: newWaypoint._photos,
            assetIds: shownAssetLinks(newWaypoint._assetLinks, newWaypoint.photos),
            pluginLinks: newWaypoint._assetPluginLinks,
            target: {
                trail: model.trail ?? newWaypoint.trail,
                waypoint: model.id,
            },
            existingPhotos: model.photos ?? newWaypoint.photos,
            onOmitted,
        });
    } catch (importError) {
        const attachedPhotos: string[] = importError instanceof APIError ? importError.detail?.attachedPhotos ?? [] : [];
        model.photos = [...new Set([...(oldWaypoint.photos ?? []), ...savedAttachedPhotos(attachedPhotos, newWaypoint, importError)])];
        throw new APIError(422, "asset_import_failed", {
            importError,
            savedWaypoint: model,
            remainingAttachments: remainingWaypointAttachments(newWaypoint, importError),
        });
    }
    const removedPhotos = new Set((oldWaypoint.photos ?? [])
        .filter((photo) => !newWaypoint.photos?.includes(photo)));
    const selectedPhotos = [...new Set([
        ...(newWaypoint.photos ?? []),
        ...model.photos.filter((photo) => !removedPhotos.has(photo)),
    ])];
    try {
        await assets_delete_removed(oldWaypoint.photos, selectedPhotos, {
            waypoint: model.id,
        });
    } catch (cause) {
        // All additions were saved. Preserve the pending removals separately
        // from the last known stored photos so retry cannot upload them again.
        const error = cause instanceof APIError ? cause : new APIError(502, "error-saving-trail", { cause });
        error.detail = {
            ...error.detail,
            savedWaypoint: {
                ...model,
                photos: [...new Set([...(oldWaypoint.photos ?? []), ...model.photos])],
            },
            remainingAttachments: { _photos: undefined, _assetLinks: undefined, _assetPluginLinks: undefined },
            remainingPhotos: selectedPhotos,
        };
        throw error;
    }
    model.photos = selectedPhotos;

    return model;

}

export async function waypoints_delete(waypoint: Waypoint) {
    const r = await fetch('/api/v1/waypoint/' + waypoint.id, {
        method: 'DELETE',
    })
    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }

    return await r.json();

}
