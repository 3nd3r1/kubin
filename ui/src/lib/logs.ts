import type { LogFile, Manifest } from "./api";
import type { ContainerStatus, Pod } from "./k8s";

export type ContainerLogs = {
    container: string;
    current?: LogFile;
    previous?: LogFile;
};

/** The pod's captured logs per container, in the order of the pod spec. */
export function podLogs(manifest: Manifest, pod: Pod): ContainerLogs[] {
    const ns = pod.metadata?.namespace;
    const name = pod.metadata?.name;
    const byContainer = new Map<string, ContainerLogs>();
    for (const f of manifest.logs?.files ?? []) {
        if (f.namespace !== ns || f.pod !== name) continue;
        const c = byContainer.get(f.container) ?? { container: f.container };
        if (f.previous) c.previous = f;
        else c.current = f;
        byContainer.set(f.container, c);
    }
    const order = [
        ...(pod.spec?.initContainers ?? []),
        ...(pod.spec?.containers ?? []),
    ].map((c) => c.name);
    const rank = (c: string) => {
        const i = order.indexOf(c);
        return i < 0 ? order.length : i;
    };
    return [...byContainer.values()].sort(
        (a, b) =>
            rank(a.container) - rank(b.container) ||
            a.container.localeCompare(b.container)
    );
}

/** The last run of the container crashed and it hasn't come back up. */
function crashing(s: ContainerStatus): boolean {
    if (s.state?.waiting?.reason === "CrashLoopBackOff") return true;
    const exit = s.state?.terminated?.exitCode;
    return exit !== undefined && exit !== 0;
}

/** Running again after a run that failed. */
function recovered(s: ContainerStatus): boolean {
    const exit = s.lastState?.terminated?.exitCode;
    return !!s.state?.running && exit !== undefined && exit !== 0;
}

/**
 * The log to open first. For a crash-looping container that is the log of
 * its last run, where the crash is: the kubelet keeps only the newest dead
 * container, so that is the current log, and the previous one only when it
 * is all that was captured. For a container that failed and is running
 * again, the crash is in the previous log. Otherwise the first container's
 * log. crashing says the pod is crash-looping now.
 */
export function defaultLog(
    pod: Pod,
    logs: ContainerLogs[]
): { file: LogFile; crashing: boolean } | undefined {
    const statuses = [
        ...(pod.status?.initContainerStatuses ?? []),
        ...(pod.status?.containerStatuses ?? []),
    ];
    const status = (c: ContainerLogs) =>
        statuses.find((s) => s.name === c.container);
    for (const c of logs) {
        const s = status(c);
        const file = c.current ?? c.previous;
        if (s && file && crashing(s)) return { file, crashing: true };
    }
    for (const c of logs) {
        const s = status(c);
        const file = c.previous ?? c.current;
        if (s && file && recovered(s)) return { file, crashing: false };
    }
    const first = logs.find((c) => c.current ?? c.previous);
    const file = first && (first.current ?? first.previous);
    return file ? { file, crashing: false } : undefined;
}

/** Lines that read like an error, to highlight them in a log. */
export const errorLine =
    /\b(error|err|fatal|panic|exception|failed|failure|traceback)\b/i;

/** The log chosen in the pod URL: `?container={name}&previous=1&line={n}`. */
export type LogParams = {
    container?: string;
    previous: boolean;
    line?: number;
};

export function parseLogParams(params: URLSearchParams): LogParams {
    const raw = params.get("line");
    const n = raw && /^\d+$/.test(raw) ? Number(raw) : 0;
    return {
        container: params.get("container") || undefined,
        previous: params.get("previous") === "1",
        line: n >= 1 ? n : undefined,
    };
}

/** The params with the log selection replaced; other params are kept. */
export function formatLogParams(
    params: URLSearchParams,
    p: { container: string; previous: boolean; line?: number }
): URLSearchParams {
    const next = new URLSearchParams(params);
    next.set("container", p.container);
    if (p.previous) next.set("previous", "1");
    else next.delete("previous");
    if (p.line !== undefined && p.line >= 1) next.set("line", String(p.line));
    else next.delete("line");
    return next;
}

/**
 * The lines to render for a scrolled viewport of fixed-height lines: the
 * visible ones plus `overscan` on each side. start is inclusive, end is
 * exclusive.
 */
export function visibleRange(
    scrollTop: number,
    viewportHeight: number,
    lineHeight: number,
    lineCount: number,
    overscan: number
): { start: number; end: number } {
    if (lineCount <= 0 || lineHeight <= 0) return { start: 0, end: 0 };
    const first = Math.floor(Math.max(0, scrollTop) / lineHeight);
    const last = Math.ceil(
        (Math.max(0, scrollTop) + viewportHeight) / lineHeight
    );
    const start = Math.min(Math.max(0, first - overscan), lineCount);
    const end = Math.min(lineCount, Math.max(start, last + overscan));
    return { start, end };
}
