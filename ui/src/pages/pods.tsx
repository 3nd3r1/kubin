import { useMemo } from "react";

import { Link, useNavigate } from "react-router";

import { listResources } from "@/lib/api";
import type { Pod } from "@/lib/k8s";
import { type PodSummary, byNamespace, problems, summarize } from "@/lib/pods";
import { podPath, useSnapshot } from "@/lib/snapshot";
import { useFetch } from "@/lib/use-fetch";
import { cn } from "@/lib/utils";

import { LoadError, Loading } from "@/components/common/states";
import { PodStatus } from "@/components/common/status";

/** Every pod in the snapshot, with the ones in trouble on top. */
export function PodsPage() {
    const snapshot = useSnapshot();
    const state = useFetch(
        (signal) => listResources<Pod>(snapshot.id, { kind: "Pod" }, signal),
        [snapshot.id]
    );
    const pods = useMemo(
        () =>
            state.status === "ok"
                ? state.data.map((r) =>
                      summarize(r.object, snapshot.manifest.created_at)
                  )
                : [],
        [state, snapshot.manifest.created_at]
    );

    if (state.status === "loading") return <Loading what="pods" />;
    if (state.status === "error") return <LoadError message={state.message} />;
    if (state.status === "missing") return <LoadError message="not found" />;
    if (pods.length === 0)
        return (
            <p className="text-muted-foreground">This snapshot has no pods.</p>
        );

    const trouble = problems(pods);
    return (
        <div className="space-y-10">
            <section aria-labelledby="problems" className="space-y-3">
                <h2 id="problems" className="text-lg font-semibold">
                    {trouble.length === 0
                        ? "No problems"
                        : `${trouble.length} ${trouble.length === 1 ? "pod needs" : "pods need"} attention`}
                </h2>
                {trouble.length === 0 ? (
                    <p className="text-muted-foreground text-sm">
                        All {pods.length} pods are running and ready, with no
                        restarts.
                    </p>
                ) : (
                    <PodTable pods={trouble} showNamespace />
                )}
            </section>

            <section aria-labelledby="all-pods" className="space-y-6">
                <h2 id="all-pods" className="text-lg font-semibold">
                    All pods{" "}
                    <span className="text-muted-foreground font-normal">
                        {pods.length}
                    </span>
                </h2>
                {byNamespace(pods).map(([ns, ps]) => (
                    <div key={ns} className="space-y-2">
                        <h3 className="text-muted-foreground font-mono text-sm">
                            {ns}
                        </h3>
                        <PodTable pods={ps} />
                    </div>
                ))}
            </section>
        </div>
    );
}

function PodTable({
    pods,
    showNamespace = false,
}: {
    pods: PodSummary[];
    showNamespace?: boolean;
}) {
    const { id } = useSnapshot();
    const navigate = useNavigate();
    return (
        <div className="overflow-x-auto rounded-lg border">
            <table className="w-full text-sm">
                <thead className="bg-muted/50 text-muted-foreground text-left">
                    <tr>
                        {showNamespace && (
                            <th className="px-3 py-2 font-medium">Namespace</th>
                        )}
                        <th className="px-3 py-2 font-medium">Name</th>
                        <th className="px-3 py-2 font-medium">Status</th>
                        <th className="px-3 py-2 font-medium">Ready</th>
                        <th className="px-3 py-2 text-right font-medium">
                            Restarts
                        </th>
                        <th className="px-3 py-2 text-right font-medium">
                            Age
                        </th>
                    </tr>
                </thead>
                <tbody>
                    {pods.map((p) => (
                        <tr
                            key={`${p.namespace}/${p.name}`}
                            onClick={() => navigate(podPath(id, p))}
                            className={cn(
                                "hover:bg-muted/50 cursor-pointer border-t",
                                p.problem === "error" && "bg-red-500/5",
                                p.problem === "warning" && "bg-amber-500/5"
                            )}
                        >
                            {showNamespace && (
                                <td className="text-muted-foreground px-3 py-2 font-mono">
                                    {p.namespace}
                                </td>
                            )}
                            <td className="px-3 py-2 font-mono">
                                <Link
                                    to={podPath(id, p)}
                                    className="hover:underline"
                                    onClick={(e) => e.stopPropagation()}
                                >
                                    {p.name}
                                </Link>
                            </td>
                            <td className="px-3 py-2">
                                <PodStatus
                                    status={p.status}
                                    problem={p.problem}
                                />
                            </td>
                            <td className="px-3 py-2 tabular-nums">
                                {p.ready}
                            </td>
                            <td
                                className={cn(
                                    "px-3 py-2 text-right tabular-nums",
                                    p.restarts > 0 && "font-semibold"
                                )}
                            >
                                {p.restarts}
                            </td>
                            <td className="text-muted-foreground px-3 py-2 text-right tabular-nums">
                                {p.age}
                            </td>
                        </tr>
                    ))}
                </tbody>
            </table>
        </div>
    );
}
