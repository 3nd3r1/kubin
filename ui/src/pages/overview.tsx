import { useEffect, useMemo, useState } from "react";

import { Link, useSearchParams } from "react-router";

import { type ResourceItem, listResources } from "@/lib/api";
import { problem } from "@/lib/columns";
import { recentWarnings } from "@/lib/events";
import type { AnyObject, Event, ObjectMeta, Pod } from "@/lib/k8s";
import { groupKinds, kinds, segmentFor } from "@/lib/kinds";
import { type Problem, problems, summarize } from "@/lib/pods";
import { selectorText, selectsNoPods } from "@/lib/relations";
import { kindPath, nsParam, resourcePath, useSnapshot } from "@/lib/snapshot";
import { useFetch } from "@/lib/use-fetch";
import { cn } from "@/lib/utils";

import { LoadError, Loading } from "@/components/common/states";

import { ExpiredPage } from "./not-found";
import { PodTable } from "./pods";

/** Kinds with a rule in problem(), other than Pod. */
const checked = [
    "Node",
    "Deployment",
    "StatefulSet",
    "DaemonSet",
    "ReplicaSet",
    "Job",
    "PersistentVolumeClaim",
    "PersistentVolume",
];

type Flagged = {
    kind: string;
    item: ResourceItem<AnyObject>;
    level: NonNullable<Problem>;
    note?: string;
};

/** The part of a Service the overview reads. */
type ServiceObject = {
    metadata?: ObjectMeta;
    spec?: { selector?: Record<string, string> };
};

type Loaded = { flagged: Flagged[]; errors: string[]; pending: number };

/** Fetches each kind separately and keeps what has arrived so far. */
function useFlagged(id: string, ns: string, present: string[]): Loaded {
    const key = JSON.stringify([id, ns, present]);
    const [state, setState] = useState<{ key: string } & Loaded>({
        key,
        flagged: [],
        errors: [],
        pending: present.length,
    });
    useEffect(() => {
        const ctrl = new AbortController();
        // State left over from other inputs starts over.
        const base = (s: { key: string } & Loaded) =>
            s.key === key
                ? s
                : { key, flagged: [], errors: [], pending: present.length };
        for (const kind of present) {
            const namespaced = kinds.find((k) => k.kind === kind)?.namespaced;
            listResources<AnyObject>(
                id,
                { kind, namespace: namespaced ? ns : "" },
                ctrl.signal
            )
                .then((items) => {
                    const flagged: Flagged[] = [];
                    for (const item of items) {
                        const level = problem(kind, item.object);
                        if (level) flagged.push({ kind, item, level });
                    }
                    setState((s) => {
                        const b = base(s);
                        return {
                            ...b,
                            flagged: [...b.flagged, ...flagged],
                            pending: b.pending - 1,
                        };
                    });
                })
                .catch((err: unknown) => {
                    if (ctrl.signal.aborted) return;
                    const message =
                        err instanceof Error ? err.message : String(err);
                    setState((s) => {
                        const b = base(s);
                        return {
                            ...b,
                            errors: [...b.errors, `${kind}: ${message}`],
                            pending: b.pending - 1,
                        };
                    });
                });
        }
        return () => ctrl.abort();
        // key carries every input.
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [key]);
    return state.key === key
        ? state
        : { flagged: [], errors: [], pending: present.length };
}

const rank = { error: 0, warning: 1 } as const;

/** The landing page of a snapshot: what is wrong first, then the counts. */
export function OverviewPage() {
    const snapshot = useSnapshot();
    const [params] = useSearchParams();
    const ns = nsParam(params);
    const { id, manifest, counts } = snapshot;
    const created = manifest.created_at;

    const podsState = useFetch(
        (signal) =>
            listResources<Pod>(id, { kind: "Pod", namespace: ns }, signal),
        [id, ns]
    );
    const eventsState = useFetch(
        (signal) =>
            listResources<Event>(id, { kind: "Event", namespace: ns }, signal),
        [id, ns]
    );
    const hasPods = counts.some((c) => c.kind === "Pod" && c.group === "core");
    const hasServices = counts.some(
        (c) => c.kind === "Service" && c.group === "core"
    );
    const servicesState = useFetch(
        (signal) =>
            hasServices && hasPods
                ? listResources<ServiceObject>(
                      id,
                      { kind: "Service", namespace: ns },
                      signal
                  )
                : Promise.resolve([]),
        [id, ns, hasServices, hasPods]
    );
    const present = checked.filter((k) => counts.some((c) => c.kind === k));
    const other = useFlagged(id, ns, present);

    const pods = useMemo(
        () =>
            podsState.status === "ok"
                ? problems(
                      podsState.data
                          .filter((r) => r.group === "core")
                          .map((r) => summarize(r.object, created))
                  )
                : [],
        [podsState, created]
    );
    const orphans = useMemo(() => {
        if (!hasPods || servicesState.status !== "ok") return [];
        if (podsState.status !== "ok") return [];
        const podItems = podsState.data.filter((r) => r.group === "core");
        return servicesState.data
            .filter((r) => r.group === "core")
            .filter((r) => selectsNoPods(r.object, podItems))
            .map((item): Flagged => ({
                kind: "Service",
                item,
                level: "warning",
                note: `selects no pods (${selectorText(item.object)})`,
            }));
    }, [hasPods, servicesState, podsState]);
    const flagged = useMemo(
        () =>
            [...other.flagged, ...orphans].sort(
                (a, b) =>
                    rank[a.level] - rank[b.level] ||
                    a.kind.localeCompare(b.kind) ||
                    a.item.namespace.localeCompare(b.item.namespace) ||
                    a.item.name.localeCompare(b.item.name)
            ),
        [other.flagged, orphans]
    );
    const warnings = useMemo(
        () =>
            eventsState.status === "ok"
                ? recentWarnings(
                      eventsState.data.filter(
                          (r) =>
                              r.group === "core" || r.group === "events.k8s.io"
                      ),
                      created
                  )
                : [],
        [eventsState, created]
    );

    if (
        podsState.status === "expired" ||
        eventsState.status === "expired" ||
        servicesState.status === "expired"
    ) {
        return <ExpiredPage />;
    }

    const nodes = counts.find((c) => c.kind === "Node" && c.group === "core");
    const loading =
        podsState.status === "loading" ||
        eventsState.status === "loading" ||
        servicesState.status === "loading" ||
        other.pending > 0;
    const failed = [
        podsState.status === "error" && `Pod: ${podsState.message}`,
        eventsState.status === "error" && `Event: ${eventsState.message}`,
        servicesState.status === "error" && `Service: ${servicesState.message}`,
        ...other.errors,
    ].filter((e): e is string => !!e);
    const clean =
        !loading &&
        failed.length === 0 &&
        pods.length === 0 &&
        flagged.length === 0 &&
        warnings.length === 0;

    return (
        <div className="space-y-10">
            <section aria-labelledby="about" className="space-y-3">
                <h2 id="about" className="sr-only">
                    Snapshot
                </h2>
                <dl className="grid grid-cols-2 gap-4 text-sm sm:grid-cols-4">
                    <Fact term="Cluster">
                        {manifest.cluster.name ||
                            manifest.cluster.context ||
                            "unknown"}
                    </Fact>
                    <Fact term="Kubernetes">
                        {manifest.cluster.server_version}
                    </Fact>
                    <Fact term="Captured">
                        <time dateTime={created} title={created}>
                            {new Date(created).toLocaleString()}
                        </time>
                    </Fact>
                    <Fact term="Nodes">{nodes?.count ?? 0}</Fact>
                </dl>
                <div className="flex flex-wrap gap-2">
                    {groupKinds(counts).flatMap((s) =>
                        s.items.map((i) => (
                            <Link
                                key={i.segment}
                                to={kindPath(
                                    id,
                                    i.segment,
                                    i.info?.namespaced === false ? "" : ns
                                )}
                                className="hover:bg-muted rounded border px-2 py-1 text-sm"
                            >
                                {i.kind}{" "}
                                <span className="text-muted-foreground tabular-nums">
                                    {i.count}
                                </span>
                            </Link>
                        ))
                    )}
                </div>
            </section>

            {failed.map((m) => (
                <LoadError key={m} message={m} />
            ))}

            {clean && (
                <p className="text-sm">
                    Nothing looks wrong
                    {ns ? ` in namespace ${ns}` : ""}: no failing pods or
                    workloads, and no warnings in the hour before capture.
                </p>
            )}

            {pods.length > 0 && (
                <section aria-labelledby="pod-problems" className="space-y-3">
                    <h2 id="pod-problems" className="text-lg font-semibold">
                        {pods.length}{" "}
                        {pods.length === 1 ? "pod needs" : "pods need"}{" "}
                        attention
                    </h2>
                    <PodTable pods={pods} showNamespace />
                </section>
            )}
            {podsState.status === "loading" && <Loading what="pods" />}

            {flagged.length > 0 && (
                <section aria-labelledby="other-problems" className="space-y-3">
                    <h2 id="other-problems" className="text-lg font-semibold">
                        Other problems{" "}
                        <span className="text-muted-foreground font-normal">
                            {flagged.length}
                        </span>
                    </h2>
                    <ul className="divide-y rounded-lg border text-sm">
                        {flagged.map((f) => (
                            <li
                                key={`${f.kind}/${f.item.namespace}/${f.item.name}`}
                                className={cn(
                                    "flex flex-wrap items-baseline gap-x-3 px-3 py-2",
                                    f.level === "error"
                                        ? "bg-red-500/5"
                                        : "bg-amber-500/5"
                                )}
                            >
                                <span
                                    className={cn(
                                        "w-16 font-medium",
                                        f.level === "error"
                                            ? "text-red-600 dark:text-red-400"
                                            : "text-amber-600 dark:text-amber-400"
                                    )}
                                >
                                    {f.level === "error" ? "Error" : "Warning"}
                                </span>
                                <span className="text-muted-foreground">
                                    {f.kind}
                                </span>
                                <Link
                                    to={resourcePath(
                                        id,
                                        segmentFor(f.kind, f.item.group),
                                        f.item
                                    )}
                                    className="font-mono hover:underline"
                                >
                                    {f.item.namespace
                                        ? `${f.item.namespace}/`
                                        : ""}
                                    {f.item.name}
                                </Link>
                                {f.note && (
                                    <span className="text-muted-foreground">
                                        {f.note}
                                    </span>
                                )}
                            </li>
                        ))}
                    </ul>
                </section>
            )}
            {other.pending > 0 && <Loading what="workloads" />}

            {warnings.length > 0 && (
                <section aria-labelledby="warnings" className="space-y-3">
                    <h2 id="warnings" className="text-lg font-semibold">
                        Warnings in the hour before capture{" "}
                        <span className="text-muted-foreground font-normal">
                            {warnings.length}
                        </span>
                    </h2>
                    <ul className="divide-y rounded-lg border text-sm">
                        {warnings.map((w, i) => {
                            const o = w.object;
                            const target = o.involvedObject;
                            return (
                                <li
                                    key={`${w.namespace}/${w.name}/${i}`}
                                    className="space-y-0.5 px-3 py-2"
                                >
                                    <div className="flex flex-wrap gap-x-3">
                                        <Link
                                            to={resourcePath(
                                                id,
                                                segmentFor(w.kind, w.group),
                                                w
                                            )}
                                            className="font-medium hover:underline"
                                        >
                                            {o.reason || "Warning"}
                                        </Link>
                                        {target?.kind && (
                                            <span className="text-muted-foreground font-mono">
                                                {target.kind.toLowerCase()}/
                                                {target.name}
                                            </span>
                                        )}
                                    </div>
                                    <p className="text-muted-foreground">
                                        {(o.message || "").trim()}
                                    </p>
                                </li>
                            );
                        })}
                    </ul>
                </section>
            )}
            {eventsState.status === "loading" && <Loading what="events" />}
        </div>
    );
}

function Fact({ term, children }: { term: string; children: React.ReactNode }) {
    return (
        <div>
            <dt className="text-muted-foreground text-xs uppercase">{term}</dt>
            <dd className="font-medium">{children}</dd>
        </div>
    );
}
