import { Link, useNavigate } from "react-router";

import type { PodSummary } from "@/lib/pods";
import { podPath, useSnapshot } from "@/lib/snapshot";
import { cn } from "@/lib/utils";

import { PodStatus } from "@/components/common/status";

export function PodTable({
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
