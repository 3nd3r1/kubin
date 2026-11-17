import type { ContainerState, Pod } from "./k8s";

export type Problem = "error" | "warning" | null;

export type PodSummary = {
    namespace: string;
    name: string;
    /** What kubectl get pods shows in STATUS, like CrashLoopBackOff. */
    status: string;
    ready: string;
    restarts: number;
    /** Age at capture time, like 3d or 12m. Empty when unknown. */
    age: string;
    node: string;
    problem: Problem;
};

// Waiting and terminated reasons that mean a container can't run.
const errorReasons = new Set([
    "CrashLoopBackOff",
    "Error",
    "OOMKilled",
    "ErrImagePull",
    "ImagePullBackOff",
    "InvalidImageName",
    "CreateContainerConfigError",
    "CreateContainerError",
    "RunContainerError",
    "ContainerCannotRun",
    "StartError",
    "Evicted",
    "Failed",
    "Unknown",
]);

function stateReason(s: ContainerState | undefined): string | undefined {
    if (s?.waiting?.reason) return s.waiting.reason;
    const t = s?.terminated;
    if (t) {
        if (t.reason) return t.reason;
        if (t.signal) return `Signal:${t.signal}`;
        return `ExitCode:${t.exitCode ?? 0}`;
    }
    return undefined;
}

/**
 * Status follows the STATUS column of kubectl get pods: the pod reason or
 * phase, then a failing or waiting init container, then the last app
 * container that isn't running.
 */
export function podStatus(pod: Pod): string {
    const st = pod.status ?? {};
    let reason = st.reason || st.phase || "Unknown";

    const inits = pod.spec?.initContainers ?? [];
    const initStatuses = st.initContainerStatuses ?? [];
    for (let i = 0; i < initStatuses.length; i++) {
        const c = initStatuses[i];
        // Sidecars (restartPolicy Always) keep running; they don't block.
        if (inits.find((s) => s.name === c.name)?.restartPolicy === "Always")
            continue;
        const t = c.state?.terminated;
        if (t && t.exitCode === 0) continue;
        if (t) return `Init:${stateReason(c.state)}`;
        const w = c.state?.waiting?.reason;
        if (w && w !== "PodInitializing") return `Init:${w}`;
        return `Init:${i}/${inits.length}`;
    }

    let running = false;
    const statuses = st.containerStatuses ?? [];
    for (let i = statuses.length - 1; i >= 0; i--) {
        const c = statuses[i];
        if (c.state?.running && c.ready) running = true;
        const r = c.state?.running ? undefined : stateReason(c.state);
        if (r) {
            reason = r;
            break;
        }
    }
    if (reason === "Completed" && running) reason = "Running";

    if (pod.metadata?.deletionTimestamp) return "Terminating";
    return reason;
}

function readyCount(pod: Pod): string {
    const total = pod.spec?.containers?.length ?? 0;
    const ready = (pod.status?.containerStatuses ?? []).filter(
        (c) => c.ready
    ).length;
    return `${ready}/${total}`;
}

export function restarts(pod: Pod): number {
    return [
        ...(pod.status?.containerStatuses ?? []),
        ...(pod.status?.initContainerStatuses ?? []),
    ].reduce((n, c) => n + (c.restartCount ?? 0), 0);
}

/** Human age like kubectl: 45s, 12m, 3h, 5d, 2y. */
export function age(from: string | undefined, to: string): string {
    if (!from) return "";
    const s = Math.floor((Date.parse(to) - Date.parse(from)) / 1000);
    if (!Number.isFinite(s) || s < 0) return "";
    if (s < 120) return `${s}s`;
    if (s < 2 * 3600) return `${Math.floor(s / 60)}m`;
    if (s < 2 * 86400) return `${Math.floor(s / 3600)}h`;
    if (s < 2 * 365 * 86400) return `${Math.floor(s / 86400)}d`;
    return `${Math.floor(s / (365 * 86400))}y`;
}

export function podProblem(pod: Pod, status: string): Problem {
    const phase = pod.status?.phase;
    const reason = status.replace(/^Init:/, "");
    if (
        errorReasons.has(reason) ||
        reason.startsWith("ExitCode:") ||
        reason.startsWith("Signal:")
    )
        return "error";
    if (phase === "Succeeded" || status === "Completed") return null;
    if (status !== "Running") return "warning";
    const ready = pod.status?.conditions?.find((c) => c.type === "Ready");
    if (ready && ready.status !== "True") return "warning";
    if (restarts(pod) > 0) return "warning";
    return null;
}

export function summarize(pod: Pod, capturedAt: string): PodSummary {
    const status = podStatus(pod);
    return {
        namespace: pod.metadata?.namespace ?? "",
        name: pod.metadata?.name ?? "",
        status,
        ready: readyCount(pod),
        restarts: restarts(pod),
        age: age(pod.metadata?.creationTimestamp, capturedAt),
        node: pod.spec?.nodeName ?? "",
        problem: podProblem(pod, status),
    };
}

const rank = { error: 0, warning: 1 } as const;

/** Problem pods, errors first, then by namespace and name. */
export function problems(pods: PodSummary[]): PodSummary[] {
    return pods
        .filter((p) => p.problem)
        .sort(
            (a, b) =>
                rank[a.problem!] - rank[b.problem!] ||
                a.namespace.localeCompare(b.namespace) ||
                a.name.localeCompare(b.name)
        );
}

/** Pods grouped by namespace, namespaces sorted, problem pods first in each. */
export function byNamespace(pods: PodSummary[]): [string, PodSummary[]][] {
    const groups = new Map<string, PodSummary[]>();
    for (const p of pods) {
        const g = groups.get(p.namespace) ?? [];
        g.push(p);
        groups.set(p.namespace, g);
    }
    const order = (p: PodSummary) => (p.problem ? rank[p.problem] : 2);
    return [...groups.entries()]
        .sort(([a], [b]) => a.localeCompare(b))
        .map(([ns, ps]) => [
            ns,
            ps.sort(
                (a, b) => order(a) - order(b) || a.name.localeCompare(b.name)
            ),
        ]);
}
