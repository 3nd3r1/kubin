import { lastSeen } from "./events";
import type { AnyObject } from "./k8s";
import { type Problem, age } from "./pods";

export type Column = {
    header: string;
    /** The cell text. capturedAt is the manifest's created_at. */
    cell: (o: AnyObject, capturedAt: string) => string;
};

const none = "<none>";
const orNone = (s: string | undefined | null) => s || none;
const num = (n: number | undefined) => String(n ?? 0);
const join = (xs: string[]) => (xs.length ? xs.join(",") : none);

function ready(o: AnyObject, desired = o.spec?.replicas ?? 1): string {
    return `${num(o.status?.readyReplicas)}/${desired}`;
}

const accessShort: Record<string, string> = {
    ReadWriteOnce: "RWO",
    ReadOnlyMany: "ROX",
    ReadWriteMany: "RWX",
    ReadWriteOncePod: "RWOP",
};
const accessModes = (modes: string[] | undefined) =>
    join((modes ?? []).map((m) => accessShort[m] ?? m));

function condition(o: AnyObject, type: string): string | undefined {
    return o.status?.conditions?.find((c) => c.type === type)?.status;
}

function jobStatus(o: AnyObject): string {
    if (condition(o, "Complete") === "True") return "Complete";
    if (condition(o, "Failed") === "True") return "Failed";
    if (o.spec?.suspend) return "Suspended";
    return "Running";
}

function nodeStatus(o: AnyObject): string {
    const s = condition(o, "Ready") === "True" ? "Ready" : "NotReady";
    return o.spec?.unschedulable ? `${s},SchedulingDisabled` : s;
}

function nodeRoles(o: AnyObject): string {
    const prefix = "node-role.kubernetes.io/";
    return join(
        Object.keys(o.metadata?.labels ?? {})
            .filter((k) => k.startsWith(prefix))
            .map((k) => k.slice(prefix.length))
            .filter(Boolean)
            .sort()
    );
}

function duration(o: AnyObject, capturedAt: string): string {
    const start = o.status?.startTime;
    if (!start) return none;
    return age(start, o.status?.completionTime || capturedAt) || none;
}

function addresses(o: AnyObject): string[] {
    return (o.status?.loadBalancer?.ingress ?? [])
        .map((i) => i.ip || i.hostname || "")
        .filter(Boolean);
}

function externalIP(o: AnyObject): string {
    if (o.spec?.type === "ExternalName") return orNone(o.spec.externalName);
    const ips = [...(o.spec?.externalIPs ?? []), ...addresses(o)];
    if (ips.length) return ips.join(",");
    return o.spec?.type === "LoadBalancer" ? "<pending>" : none;
}

function servicePorts(o: AnyObject): string {
    return join(
        (o.spec?.ports ?? []).map(
            (p) =>
                `${p.port ?? 0}${p.nodePort ? `:${p.nodePort}` : ""}/${p.protocol ?? "TCP"}`
        )
    );
}

function ingressClass(o: AnyObject): string {
    return orNone(
        o.spec?.ingressClassName ||
            o.metadata?.annotations?.["kubernetes.io/ingress.class"]
    );
}

function ingressHosts(o: AnyObject): string {
    const hosts = (o.spec?.rules ?? []).map((r) => r.host || "*");
    return hosts.length ? hosts.join(",") : "*";
}

function sliceEndpoints(o: AnyObject): string {
    const all = (o.endpoints ?? []).flatMap((e) => e.addresses ?? []);
    if (all.length === 0) return "<unset>";
    const shown = all.slice(0, 3).join(",");
    return all.length > 3 ? `${shown} + ${all.length - 3} more...` : shown;
}

function nodeSelector(o: AnyObject): string {
    const sel = o.spec?.template?.spec?.nodeSelector ?? {};
    return join(
        Object.keys(sel)
            .sort()
            .map((k) => `${k}=${sel[k]}`)
    );
}

function eventObject(o: AnyObject): string {
    const r = o.involvedObject ?? o.regarding;
    return r?.kind && r.name ? `${r.kind.toLowerCase()}/${r.name}` : none;
}

const c = (header: string, cell: Column["cell"]): Column => ({ header, cell });

/** The kubectl get columns after Name and Namespace, by kind. */
const byKind: Record<string, Column[]> = {
    Deployment: [
        c("Ready", (o) => ready(o)),
        c("Up-to-date", (o) => num(o.status?.updatedReplicas)),
        c("Available", (o) => num(o.status?.availableReplicas)),
    ],
    ReplicaSet: [
        c("Desired", (o) => num(o.spec?.replicas ?? 1)),
        c("Current", (o) => num(o.status?.replicas)),
        c("Ready", (o) => num(o.status?.readyReplicas)),
    ],
    StatefulSet: [c("Ready", (o) => ready(o))],
    DaemonSet: [
        c("Desired", (o) => num(o.status?.desiredNumberScheduled)),
        c("Current", (o) => num(o.status?.currentNumberScheduled)),
        c("Ready", (o) => num(o.status?.numberReady)),
        c("Up-to-date", (o) => num(o.status?.updatedNumberScheduled)),
        c("Available", (o) => num(o.status?.numberAvailable)),
        c("Node selector", nodeSelector),
    ],
    Job: [
        c("Status", jobStatus),
        c(
            "Completions",
            (o) => `${num(o.status?.succeeded)}/${o.spec?.completions ?? 1}`
        ),
        c("Duration", duration),
    ],
    CronJob: [
        c("Schedule", (o) => orNone(o.spec?.schedule)),
        c("Timezone", (o) => orNone(o.spec?.timeZone)),
        c("Suspend", (o) => (o.spec?.suspend ? "True" : "False")),
        c("Active", (o) => num(o.status?.active)),
        c("Last schedule", (o, t) =>
            orNone(age(o.status?.lastScheduleTime, t))
        ),
    ],
    Service: [
        c("Type", (o) => o.spec?.type || "ClusterIP"),
        c("Cluster-IP", (o) => orNone(o.spec?.clusterIP)),
        c("External-IP", externalIP),
        c("Port(s)", servicePorts),
    ],
    Ingress: [
        c("Class", ingressClass),
        c("Hosts", ingressHosts),
        c("Address", (o) => join(addresses(o))),
        c("Ports", (o) => (o.spec?.tls?.length ? "80, 443" : "80")),
    ],
    EndpointSlice: [
        c("AddressType", (o) => orNone(o.addressType)),
        c("Ports", (o) => {
            const ports = (o.ports ?? [])
                .map((p) => p.port)
                .filter((p) => p !== undefined);
            return ports.length ? ports.join(",") : "<unset>";
        }),
        c("Endpoints", sliceEndpoints),
    ],
    ConfigMap: [
        c("Data", (o) =>
            String(
                Object.keys(o.data ?? {}).length +
                    Object.keys(o.binaryData ?? {}).length
            )
        ),
    ],
    PersistentVolumeClaim: [
        c("Status", (o) => orNone(o.status?.phase)),
        c("Volume", (o) => orNone(o.spec?.volumeName)),
        c("Capacity", (o) => orNone(o.status?.capacity?.storage)),
        c("Access modes", (o) => accessModes(o.status?.accessModes)),
        c("StorageClass", (o) => orNone(o.spec?.storageClassName)),
    ],
    PersistentVolume: [
        c("Capacity", (o) => orNone(o.spec?.capacity?.storage)),
        c("Access modes", (o) => accessModes(o.spec?.accessModes)),
        c("Reclaim policy", (o) =>
            orNone(o.spec?.persistentVolumeReclaimPolicy)
        ),
        c("Status", (o) => orNone(o.status?.phase)),
        c("Claim", (o) => {
            const r = o.spec?.claimRef;
            return r?.name
                ? `${r.namespace ? `${r.namespace}/` : ""}${r.name}`
                : none;
        }),
        c("StorageClass", (o) => orNone(o.spec?.storageClassName)),
    ],
    Node: [
        c("Status", nodeStatus),
        c("Roles", nodeRoles),
        c("Version", (o) => orNone(o.status?.nodeInfo?.kubeletVersion)),
    ],
    Namespace: [c("Status", (o) => orNone(o.status?.phase))],
    Event: [
        c("Last seen", (o, t) => orNone(age(lastSeen(o), t))),
        c("Type", (o) => o.type || "Normal"),
        c("Reason", (o) => orNone(o.reason)),
        c("Object", eventObject),
        c("Message", (o) => (o.message || o.note || "").trim() || none),
    ],
};

/** The extra columns of a kind, none for a kind this UI doesn't know. */
export function columnsFor(kind: string): Column[] {
    return byKind[kind] ?? [];
}

/** Events newest first; every other kind keeps the server's order. */
export function sortObjects<T extends { object: AnyObject }>(
    kind: string,
    rows: T[]
): T[] {
    if (kind !== "Event") return rows;
    const t = (r: T) => Date.parse(lastSeen(r.object)) || 0;
    return [...rows].sort((a, b) => t(b) - t(a));
}

/** Whether a row is unhealthy, as the pod table flags pods. */
export function problem(kind: string, o: AnyObject): Problem {
    switch (kind) {
        case "Deployment":
        case "StatefulSet":
        case "ReplicaSet": {
            return shortfall(
                o.spec?.replicas ?? 1,
                o.status?.readyReplicas ?? 0
            );
        }
        case "DaemonSet":
            return shortfall(
                o.status?.desiredNumberScheduled ?? 0,
                o.status?.numberReady ?? 0
            );
        case "Node":
            return condition(o, "Ready") === "True" ? null : "error";
        case "Job":
            return jobStatus(o) === "Failed" ? "error" : null;
        case "PersistentVolumeClaim":
            return o.status?.phase === "Bound" ? null : "warning";
        default:
            return null;
    }
}

function shortfall(desired: number, ready: number): Problem {
    if (ready >= desired) return null;
    return ready === 0 ? "error" : "warning";
}
