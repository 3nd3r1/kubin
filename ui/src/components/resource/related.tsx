import { Link } from "react-router";

import { type ResourceItem, type Snapshot, listResources } from "@/lib/api";
import type { ObjectMeta } from "@/lib/k8s";
import { segmentFor } from "@/lib/kinds";
import {
    type OwnerRef,
    type PodSpec,
    configMapNames,
    findOwner,
    ownedBy,
    ownedKind,
    ownerKind,
    podSpecOf,
    podsFor,
    podsOnNode,
    servicesFor,
} from "@/lib/relations";
import { resourcePath } from "@/lib/snapshot";
import { useFetch } from "@/lib/use-fetch";

import { LoadError, Loading } from "@/components/common/states";

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

function itemPath(id: string, i: ResourceItem) {
    return resourcePath(id, segmentFor(i.kind, i.group), i);
}

function Links({ id, items }: { id: string; items: ResourceItem[] }) {
    return (
        <ul className="space-y-1 text-sm">
            {items.map((i) => (
                <li key={`${i.namespace}/${i.name}`} className="font-mono">
                    <Link
                        to={itemPath(id, i)}
                        className="break-all hover:underline"
                    >
                        {i.namespace ? `${i.namespace}/` : ""}
                        {i.name}
                    </Link>
                </li>
            ))}
        </ul>
    );
}

/** One owner reference: a link when the owner is in the snapshot, else text. */
function OwnerLink({
    snapshot,
    namespace,
    owner,
}: {
    snapshot: Snapshot;
    namespace: string;
    owner: OwnerRef;
}) {
    const target = ownerKind(owner);
    const label = `${owner.kind ?? "?"}/${owner.name ?? "?"}`;
    const collected =
        target &&
        snapshot.counts.some(
            (c) => c.kind === target.kind && c.group === target.group
        );
    const state = useFetch(
        (signal) =>
            collected
                ? listResources<Obj>(
                      snapshot.id,
                      {
                          kind: target.kind,
                          namespace:
                              target.namespaced === false ? "" : namespace,
                      },
                      signal
                  )
                : Promise.resolve([]),
        [snapshot.id, collected, target?.kind, namespace]
    );
    const found =
        state.status === "ok" ? findOwner<Obj>(owner, state.data) : undefined;
    return found ? (
        <Link to={itemPath(snapshot.id, found)} className="hover:underline">
            {label}
        </Link>
    ) : (
        <span>{label}</span>
    );
}

/** The object's owners as links. */
export function OwnerLinks({
    snapshot,
    object,
}: {
    snapshot: Snapshot;
    object: { metadata?: ObjectMeta };
}) {
    const refs = object.metadata?.ownerReferences ?? [];
    if (refs.length === 0) return <>–</>;
    return (
        <span className="flex flex-col">
            {refs.map((r) => (
                <OwnerLink
                    key={r.uid ?? `${r.kind}/${r.name}`}
                    snapshot={snapshot}
                    namespace={object.metadata?.namespace ?? ""}
                    owner={r}
                />
            ))}
        </span>
    );
}

/** The node's name, linked when the node is in the snapshot. */
export function NodeLink({
    snapshot,
    name,
}: {
    snapshot: Snapshot;
    name: string;
}) {
    const has = snapshot.counts.some(
        (c) => c.kind === "Node" && c.group === "core"
    );
    const state = useFetch(
        (signal) =>
            has
                ? listResources<Obj>(snapshot.id, { kind: "Node" }, signal)
                : Promise.resolve([]),
        [snapshot.id, has]
    );
    const found =
        state.status === "ok"
            ? state.data.find((n) => n.name === name)
            : undefined;
    return found ? (
        <Link to={itemPath(snapshot.id, found)} className="hover:underline">
            {name}
        </Link>
    ) : (
        <>{name}</>
    );
}

function Related({
    title,
    id,
    depKey,
    load,
    pick,
}: {
    title: string;
    id: string;
    depKey: string;
    load: (signal: AbortSignal) => Promise<ResourceItem<Obj>[]>;
    pick: (items: ResourceItem<Obj>[]) => ResourceItem<Obj>[];
}) {
    const state = useFetch(load, [depKey, title]);
    let body;
    if (state.status === "loading") body = <Loading what={title} />;
    else if (state.status !== "ok")
        body = (
            <LoadError
                message={
                    state.status === "error"
                        ? state.message
                        : "snapshot not found"
                }
            />
        );
    else {
        const items = pick(state.data);
        if (items.length === 0) return null;
        body = <Links id={id} items={items} />;
    }
    return (
        <section aria-label={title} className="space-y-2">
            <h3 className="text-lg font-semibold">{title}</h3>
            {body}
        </section>
    );
}

/** ConfigMaps the pod spec references: links, or text when not in the snapshot. */
function ConfigMaps({
    id,
    namespace,
    names,
    depKey,
}: {
    id: string;
    namespace: string;
    names: string[];
    depKey: string;
}) {
    const state = useFetch(
        (signal) =>
            listResources<Obj>(id, { kind: "ConfigMap", namespace }, signal),
        [depKey, "ConfigMaps"]
    );
    let body;
    if (state.status === "loading") body = <Loading what="ConfigMaps" />;
    else if (state.status !== "ok")
        body = (
            <LoadError
                message={
                    state.status === "error"
                        ? state.message
                        : "snapshot not found"
                }
            />
        );
    else {
        const data = state.data;
        body = (
            <ul className="space-y-1 text-sm">
                {names.map((n) => {
                    const found = data.find((c) => c.name === n);
                    return (
                        <li key={n} className="font-mono">
                            {found ? (
                                <Link
                                    to={itemPath(id, found)}
                                    className="break-all hover:underline"
                                >
                                    {n}
                                </Link>
                            ) : (
                                <span className="break-all">
                                    {n}{" "}
                                    <span className="font-sans text-muted-foreground">
                                        (not in snapshot)
                                    </span>
                                </span>
                            )}
                        </li>
                    );
                })}
            </ul>
        );
    }
    return (
        <section aria-label="ConfigMaps" className="space-y-2">
            <h3 className="text-lg font-semibold">ConfigMaps</h3>
            {body}
        </section>
    );
}

/** What the object owns, plus services, pods or nodes around it. */
export function RelatedResources({
    snapshot,
    object,
}: {
    snapshot: Snapshot;
    object: Obj;
}) {
    const id = snapshot.id;
    const kind = object.kind ?? "";
    const meta = object.metadata ?? {};
    const ns = meta.namespace ?? "";
    const has = (k: string) => snapshot.counts.some((c) => c.kind === k);
    const list = (k: string, namespace: string) => (signal: AbortSignal) =>
        listResources<Obj>(id, { kind: k, namespace }, signal);
    const key = `${id}/${kind}/${ns}/${meta.name}/${meta.uid}`;
    const sections = [];

    const child = ownedKind(kind);
    if (child && has(child))
        sections.push(
            <Related
                key="owned"
                title={`${child}s`}
                id={id}
                depKey={key}
                load={list(child, ns)}
                pick={(items) => ownedBy(meta.uid, items)}
            />
        );
    if (kind === "Pod" && has("Service"))
        sections.push(
            <Related
                key="services"
                title="Services"
                id={id}
                depKey={key}
                load={list("Service", ns)}
                pick={(items) => servicesFor(object, items)}
            />
        );
    if (kind === "Service" && has("Pod"))
        sections.push(
            <Related
                key="pods"
                title="Pods"
                id={id}
                depKey={key}
                load={list("Pod", ns)}
                pick={(items) => podsFor(object, items)}
            />
        );
    if (kind === "Node" && has("Pod"))
        sections.push(
            <Related
                key="nodepods"
                title="Pods"
                id={id}
                depKey={key}
                load={list("Pod", "")}
                pick={(items) => podsOnNode(meta.name ?? "", items)}
            />
        );
    const cms = configMapNames(podSpecOf(object));
    if (cms.length > 0 && has("ConfigMap"))
        sections.push(
            <ConfigMaps
                key="configmaps"
                id={id}
                namespace={ns}
                names={cms}
                depKey={key}
            />
        );
    return <>{sections}</>;
}
