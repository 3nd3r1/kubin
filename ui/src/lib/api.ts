// Types and calls for the server API, served on the same origin.

export type KindCount = {
    group: string;
    version: string;
    kind: string;
    count: number;
};

export type Manifest = {
    format_version: number;
    created_at: string;
    cli_version: string;
    cluster: { context: string; name: string; server_version: string };
};

export type Snapshot = {
    id: string;
    uploaded_at: string;
    manifest: Manifest;
    counts: KindCount[];
};

export class NotFoundError extends Error {}

async function getJSON<T>(path: string, signal?: AbortSignal): Promise<T> {
    const resp = await fetch(path, {
        signal,
        headers: { Accept: "application/json" },
    });
    if (resp.status === 404) throw new NotFoundError(path);
    if (!resp.ok) {
        const body = await resp.json().catch(() => null);
        throw new Error(body?.error ?? `${resp.status} ${resp.statusText}`);
    }
    return resp.json() as Promise<T>;
}

export function getSnapshot(id: string, signal?: AbortSignal) {
    return getJSON<Snapshot>(
        `/api/v1/snapshots/${encodeURIComponent(id)}`,
        signal
    );
}
