// Types and calls for the server API, served on the same origin.

export type KindCount = {
    group: string;
    version: string;
    kind: string;
    count: number;
};

export type LogFile = {
    namespace: string;
    pod: string;
    container: string;
    previous: boolean;
    truncated: boolean;
};

export type Manifest = {
    format_version: number;
    created_at: string;
    cli_version: string;
    cluster: { context: string; name: string; server_version: string };
    /** Missing in format v1 snapshots, which have no logs. */
    logs?: { tail_lines: number; max_bytes: number; files: LogFile[] };
    /** The redaction rules that ran. Missing before format v3. */
    redactions?: string[];
};

export type Snapshot = {
    id: string;
    uploaded_at: string;
    expires_at?: string;
    manifest: Manifest;
    counts: KindCount[];
};

export class NotFoundError extends Error {}

export class ExpiredError extends Error {}

async function get(
    path: string,
    accept: string,
    signal?: AbortSignal
): Promise<Response> {
    const resp = await fetch(path, { signal, headers: { Accept: accept } });
    if (resp.status === 404) throw new NotFoundError(path);
    if (resp.status === 410) throw new ExpiredError(path);
    if (!resp.ok) {
        const body = await resp.json().catch(() => null);
        throw new Error(body?.error ?? `${resp.status} ${resp.statusText}`);
    }
    return resp;
}

async function getJSON<T>(path: string, signal?: AbortSignal): Promise<T> {
    const resp = await get(path, "application/json", signal);
    return resp.json() as Promise<T>;
}

export function getSnapshot(id: string, signal?: AbortSignal) {
    return getJSON<Snapshot>(
        `/api/v1/snapshots/${encodeURIComponent(id)}`,
        signal
    );
}

export type ResourceItem<T = unknown> = {
    group: string;
    version: string;
    kind: string;
    namespace: string;
    name: string;
    object: T;
};

const snapshotPath = (id: string) =>
    `/api/v1/snapshots/${encodeURIComponent(id)}`;

export async function listResources<T>(
    id: string,
    filter: { kind?: string; namespace?: string },
    signal?: AbortSignal
): Promise<ResourceItem<T>[]> {
    const q = new URLSearchParams();
    if (filter.kind) q.set("kind", filter.kind);
    if (filter.namespace) q.set("namespace", filter.namespace);
    const res = await getJSON<{ items: ResourceItem<T>[] }>(
        `${snapshotPath(id)}/resources?${q}`,
        signal
    );
    return res.items;
}

export function getResource<T>(
    id: string,
    kind: string,
    namespace: string,
    name: string,
    signal?: AbortSignal
): Promise<T> {
    const parts = [kind, namespace || "_cluster", name].map(encodeURIComponent);
    return getJSON<T>(
        `${snapshotPath(id)}/resources/${parts.join("/")}`,
        signal
    );
}

/** One container log as text. The CLI caps each log at 1 MiB. */
export async function getLog(
    id: string,
    f: Pick<LogFile, "namespace" | "pod" | "container" | "previous">,
    signal?: AbortSignal
): Promise<string> {
    const parts = [f.namespace, f.pod, f.container].map(encodeURIComponent);
    const q = f.previous ? "?previous=true" : "";
    const resp = await get(
        `${snapshotPath(id)}/logs/${parts.join("/")}${q}`,
        "text/plain",
        signal
    );
    return resp.text();
}
