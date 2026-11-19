import type { KindCount } from "./api";

export type Section =
    "Workloads" | "Network" | "Config" | "Storage" | "Cluster" | "Other";

export type KindInfo = {
    /** The group as the API returns it: "core" for the core group. */
    group: string;
    kind: string;
    /** The URL segment, the lowercase plural kubectl uses. */
    segment: string;
    section: Section;
    namespaced: boolean;
};

const k = (
    group: string,
    kind: string,
    segment: string,
    section: Section,
    namespaced = true
): KindInfo => ({ group, kind, segment, section, namespaced });

/** The kinds in Kinds in cli/pkg/collector/resources.go, in sidebar order. */
export const kinds: KindInfo[] = [
    k("core", "Pod", "pods", "Workloads"),
    k("apps", "Deployment", "deployments", "Workloads"),
    k("apps", "DaemonSet", "daemonsets", "Workloads"),
    k("apps", "StatefulSet", "statefulsets", "Workloads"),
    k("apps", "ReplicaSet", "replicasets", "Workloads"),
    k("batch", "Job", "jobs", "Workloads"),
    k("batch", "CronJob", "cronjobs", "Workloads"),
    k("core", "Service", "services", "Network"),
    k("networking.k8s.io", "Ingress", "ingresses", "Network"),
    k("discovery.k8s.io", "EndpointSlice", "endpointslices", "Network"),
    k("core", "ConfigMap", "configmaps", "Config"),
    k("core", "PersistentVolumeClaim", "persistentvolumeclaims", "Storage"),
    k("core", "PersistentVolume", "persistentvolumes", "Storage", false),
    k("core", "Node", "nodes", "Cluster", false),
    k("core", "Namespace", "namespaces", "Cluster", false),
    k("core", "Event", "events", "Cluster"),
];

export const sections: Section[] = [
    "Workloads",
    "Network",
    "Config",
    "Storage",
    "Cluster",
    "Other",
];

export type SidebarItem = {
    /** What the URL carries after /s/{id}/. */
    segment: string;
    kind: string;
    count: number;
    /** Undefined for a kind this UI doesn't know. */
    info?: KindInfo;
};

export type SidebarSection = { title: Section; items: SidebarItem[] };

function find(group: string, kind: string): KindInfo | undefined {
    return kinds.find((i) => i.group === group && i.kind === kind);
}

/** A segment for a kind missing from the table: kubectl's plural guess. */
export function otherSegment(c: Pick<KindCount, "group" | "kind">): string {
    const plural = `${c.kind.toLowerCase()}s`;
    return c.group === "core" || c.group === ""
        ? plural
        : `${plural}.${c.group}`;
}

/**
 * Groups the kinds a snapshot holds into sidebar sections. Known kinds keep
 * the table's order, unknown ones go under Other. Empty sections are left out.
 */
export function groupKinds(counts: KindCount[]): SidebarSection[] {
    const items = new Map<Section, SidebarItem[]>();
    const add = (s: Section, i: SidebarItem) =>
        items.set(s, [...(items.get(s) ?? []), i]);
    for (const info of kinds) {
        const c = counts.find(
            (c) => c.group === info.group && c.kind === info.kind
        );
        if (c)
            add(info.section, {
                segment: info.segment,
                kind: info.kind,
                count: c.count,
                info,
            });
    }
    for (const c of counts) {
        if (find(c.group, c.kind)) continue;
        add("Other", {
            segment: otherSegment(c),
            kind: c.kind,
            count: c.count,
        });
    }
    return sections
        .filter((s) => items.has(s))
        .map((title) => ({ title, items: items.get(title)! }));
}

/** The kind a URL segment names in this snapshot, or undefined. */
export function resolveSegment(
    segment: string,
    counts: KindCount[]
): { kind: string; group: string; info?: KindInfo } | undefined {
    const info = kinds.find((i) => i.segment === segment);
    if (info) return { kind: info.kind, group: info.group, info };
    const c = counts.find(
        (c) => !find(c.group, c.kind) && otherSegment(c) === segment
    );
    return c && { kind: c.kind, group: c.group };
}

/** The URL segment for any kind, known or not. */
export function segmentFor(kind: string, group: string): string {
    return find(group, kind)?.segment ?? otherSegment({ kind, group });
}
