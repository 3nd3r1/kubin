import { useMemo } from "react";

import { Link, useParams, useSearchParams } from "react-router";

import { listResources } from "@/lib/api";
import { columnsFor, problem, sortObjects } from "@/lib/columns";
import type { AnyObject, Pod } from "@/lib/k8s";
import { resolveSegment } from "@/lib/kinds";
import { age, summarize } from "@/lib/pods";
import { nsParam, resourcePath, useSnapshot } from "@/lib/snapshot";
import { useFetch } from "@/lib/use-fetch";
import { cn } from "@/lib/utils";

import { Empty, LoadError, Loading } from "@/components/common/states";

import { ExpiredPage, NotFoundPage } from "./not-found";
import { PodTable } from "./pods";

/** The list page of one kind: /s/{id}/{resource}?ns= */
export function KindPage() {
    const { resource = "" } = useParams();
    const snapshot = useSnapshot();
    const found = resolveSegment(resource, snapshot.counts);
    if (!found) return <NotFoundPage />;
    // Remount per kind so no state leaks between tables.
    return <KindTable key={resource} segment={resource} {...found} />;
}

type ObjectLike = AnyObject;

function KindTable({
    segment,
    kind,
    group,
    info,
}: NonNullable<ReturnType<typeof resolveSegment>> & { segment: string }) {
    const snapshot = useSnapshot();
    const [params] = useSearchParams();
    const ns = nsParam(params);
    // Only namespaced kinds are filtered. A kind this UI doesn't know is
    // filtered too: the API ignores the filter for cluster-scoped objects.
    const filter = info && !info.namespaced ? "" : ns;
    const state = useFetch(
        (signal) =>
            listResources<ObjectLike>(
                snapshot.id,
                { kind, namespace: filter },
                signal
            ),
        [snapshot.id, kind, filter]
    );
    const items = useMemo(
        () =>
            state.status === "ok"
                ? state.data.filter((r) => r.group === group)
                : [],
        [state, group]
    );

    if (state.status === "loading") return <Loading what={kind} />;
    if (state.status === "expired") return <ExpiredPage />;
    if (state.status === "error") return <LoadError message={state.message} />;
    if (state.status === "missing") return <LoadError message="not found" />;
    if (items.length === 0)
        return (
            <Empty>
                No {kind} resources
                {filter ? ` in namespace ${filter}` : " in this snapshot"}.
            </Empty>
        );

    const created = snapshot.manifest.created_at;
    return (
        <div className="space-y-3">
            <h2 className="text-lg font-semibold">
                {kind}{" "}
                <span className="text-muted-foreground font-normal">
                    {items.length}
                </span>
            </h2>
            {kind === "Pod" && group === "core" ? (
                <PodTable
                    pods={items.map((r) => summarize(r.object as Pod, created))}
                    showNamespace
                />
            ) : (
                <GenericTable
                    kind={kind}
                    rows={sortObjects(kind, items)}
                    created={created}
                    snapshotId={snapshot.id}
                    segment={segment}
                    namespaced={
                        info ? info.namespaced : items.some((r) => r.namespace)
                    }
                />
            )}
        </div>
    );
}

function GenericTable({
    snapshotId,
    segment,
    kind,
    rows,
    created,
    namespaced,
}: {
    kind: string;
    snapshotId: string;
    segment: string;
    rows: { namespace: string; name: string; object: ObjectLike }[];
    created: string;
    namespaced: boolean;
}) {
    const columns = columnsFor(kind);
    return (
        <div className="overflow-x-auto rounded-lg border">
            <table className="w-full text-sm">
                <thead className="bg-muted/50 text-muted-foreground text-left">
                    <tr>
                        <th className="px-3 py-2 font-medium">Name</th>
                        {namespaced && (
                            <th className="px-3 py-2 font-medium">Namespace</th>
                        )}
                        {columns.map((col) => (
                            <th
                                key={col.header}
                                className="px-3 py-2 font-medium"
                            >
                                {col.header}
                            </th>
                        ))}
                        <th className="px-3 py-2 text-right font-medium">
                            Age
                        </th>
                    </tr>
                </thead>
                <tbody>
                    {rows.map((r, i) => {
                        const p = problem(kind, r.object);
                        return (
                            <tr
                                key={`${r.namespace}/${r.name}/${i}`}
                                className={cn(
                                    "border-t",
                                    p === "error" && "bg-red-500/5",
                                    p === "warning" && "bg-amber-500/5"
                                )}
                            >
                                <td className="px-3 py-2 font-mono">
                                    <Link
                                        to={resourcePath(
                                            snapshotId,
                                            segment,
                                            r
                                        )}
                                        className="hover:underline"
                                    >
                                        {r.name}
                                    </Link>
                                </td>
                                {namespaced && (
                                    <td className="text-muted-foreground px-3 py-2 font-mono">
                                        {r.namespace}
                                    </td>
                                )}
                                {columns.map((col) => (
                                    <td
                                        key={col.header}
                                        className="px-3 py-2 tabular-nums"
                                    >
                                        {col.cell(r.object, created)}
                                    </td>
                                ))}
                                <td className="text-muted-foreground px-3 py-2 text-right tabular-nums">
                                    {age(
                                        r.object.metadata?.creationTimestamp,
                                        created
                                    )}
                                </td>
                            </tr>
                        );
                    })}
                </tbody>
            </table>
        </div>
    );
}
