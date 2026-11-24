import type { ResourceItem } from "./api";
import type { ObjectMeta } from "./k8s";
import { kinds } from "./kinds";

export type OwnerRef = NonNullable<ObjectMeta["ownerReferences"]>[number];

type ConfigMapSource = { configMap?: { name?: string } };
type Container = {
    envFrom?: { configMapRef?: { name?: string } }[];
    env?: { valueFrom?: { configMapKeyRef?: { name?: string } } }[];
};

/** The parts of a pod spec that can reference a ConfigMap. */
export type PodSpec = {
    containers?: Container[];
    initContainers?: Container[];
    volumes?: (ConfigMapSource & {
        projected?: { sources?: ConfigMapSource[] };
    })[];
};

type Obj = {
    kind?: string;
    metadata?: ObjectMeta;
    spec?: {
        selector?: Record<string, string>;
        nodeName?: string;
        template?: { spec?: PodSpec };
        jobTemplate?: { spec?: { template?: { spec?: PodSpec } } };
    } & PodSpec;
};

/** The pod spec of a Pod or of a workload's pod template, if the kind has one. */
export function podSpecOf(object: Obj): PodSpec | undefined {
    const spec = object.spec;
    switch (object.kind) {
        case "Pod":
            return spec;
        case "Deployment":
        case "ReplicaSet":
        case "StatefulSet":
        case "DaemonSet":
        case "Job":
            return spec?.template?.spec;
        case "CronJob":
            return spec?.jobTemplate?.spec?.template?.spec;
        default:
            return undefined;
    }
}

/** Names of the ConfigMaps a pod spec references, each once, sorted. */
export function configMapNames(spec: PodSpec | undefined): string[] {
    const names = new Set<string>();
    const add = (n: string | undefined) => {
        if (n) names.add(n);
    };
    for (const c of [
        ...(spec?.containers ?? []),
        ...(spec?.initContainers ?? []),
    ]) {
        for (const e of c.envFrom ?? []) add(e.configMapRef?.name);
        for (const e of c.env ?? []) add(e.valueFrom?.configMapKeyRef?.name);
    }
    for (const v of spec?.volumes ?? []) {
        add(v.configMap?.name);
        for (const s of v.projected?.sources ?? []) add(s.configMap?.name);
    }
    return [...names].sort();
}

/** The kind each owner kind owns. Only these pairs are shown. */
const owned: Record<string, string> = {
    Deployment: "ReplicaSet",
    ReplicaSet: "Pod",
    StatefulSet: "Pod",
    DaemonSet: "Pod",
    Job: "Pod",
    CronJob: "Job",
};

export function ownedKind(kind: string): string | undefined {
    return owned[kind];
}

/** Items whose owner references contain this UID. Names are not compared. */
export function ownedBy<T extends Obj>(
    uid: string | undefined,
    items: ResourceItem<T>[]
): ResourceItem<T>[] {
    if (!uid) return [];
    return items.filter((i) =>
        i.object.metadata?.ownerReferences?.some((r) => r.uid === uid)
    );
}

/**
 * True when every selector entry is in labels. A missing or empty selector
 * selects nothing, as for a Service without one.
 */
export function selectorMatches(
    selector: Record<string, string> | undefined,
    labels: Record<string, string> | undefined
): boolean {
    const entries = Object.entries(selector ?? {});
    if (entries.length === 0) return false;
    return entries.every(([k, v]) => labels?.[k] === v);
}

/** Services whose selector matches the pod's labels. */
export function servicesFor<T extends Obj>(
    pod: Obj,
    services: ResourceItem<T>[]
): ResourceItem<T>[] {
    return services.filter((s) =>
        selectorMatches(s.object.spec?.selector, pod.metadata?.labels)
    );
}

/** Pods the service selects. Pods must come from the service's namespace. */
export function podsFor<T extends Obj>(
    service: Obj,
    pods: ResourceItem<T>[]
): ResourceItem<T>[] {
    const ns = service.metadata?.namespace ?? "";
    return pods.filter(
        (p) =>
            p.namespace === ns &&
            selectorMatches(service.spec?.selector, p.object.metadata?.labels)
    );
}

/** Pods scheduled on the node, from any namespace. */
export function podsOnNode<T extends Obj>(
    nodeName: string,
    pods: ResourceItem<T>[]
): ResourceItem<T>[] {
    if (!nodeName) return [];
    return pods.filter((p) => p.object.spec?.nodeName === nodeName);
}

/**
 * The item an owner reference points at, among the items of its kind. With a
 * UID in the reference the UID must match; the name alone is not enough.
 */
export function findOwner<T>(
    ref: OwnerRef,
    items: ResourceItem<T & Obj>[]
): ResourceItem<T & Obj> | undefined {
    return items.find((i) =>
        ref.uid ? i.object.metadata?.uid === ref.uid : i.name === ref.name
    );
}

/** The kind an owner reference names, with its group read from apiVersion. */
export function ownerKind(
    ref: OwnerRef
): { kind: string; group: string; namespaced?: boolean } | undefined {
    if (!ref.kind) return undefined;
    const slash = ref.apiVersion?.indexOf("/") ?? -1;
    const group = ref.apiVersion
        ? slash < 0
            ? "core"
            : ref.apiVersion.slice(0, slash)
        : undefined;
    const info = kinds.find(
        (i) => i.kind === ref.kind && (!group || i.group === group)
    );
    return info
        ? { kind: info.kind, group: info.group, namespaced: info.namespaced }
        : { kind: ref.kind, group: group ?? "" };
}
