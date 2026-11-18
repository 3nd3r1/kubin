import { TriangleAlert } from "lucide-react";

import { listResources } from "@/lib/api";
import { podEvents } from "@/lib/events";
import type { Event, Pod } from "@/lib/k8s";
import { age } from "@/lib/pods";
import { useFetch } from "@/lib/use-fetch";
import { cn } from "@/lib/utils";

import { Empty, LoadError, Loading } from "@/components/common/states";

/** The pod's events, newest first, as kubectl describe pod lists them. */
export function PodEvents({
    snapshotId,
    pod,
    capturedAt,
}: {
    snapshotId: string;
    pod: Pod;
    capturedAt: string;
}) {
    const ns = pod.metadata?.namespace ?? "";
    const state = useFetch(
        (signal) =>
            listResources<Event>(
                snapshotId,
                { kind: "Event", namespace: ns },
                signal
            ),
        [snapshotId, ns]
    );

    let body;
    switch (state.status) {
        case "loading":
            body = <Loading what="events" />;
            break;
        case "missing":
        case "error":
            body = (
                <LoadError
                    message={
                        state.status === "error"
                            ? state.message
                            : "snapshot not found"
                    }
                />
            );
            break;
        case "ok": {
            const events = podEvents(
                pod,
                state.data.map((r) => r.object)
            );
            body =
                events.length === 0 ? (
                    <Empty>
                        No events were recorded for this pod. Kubernetes keeps
                        events for an hour by default, so older ones are gone.
                    </Empty>
                ) : (
                    <div className="overflow-x-auto rounded-lg border">
                        <table className="w-full text-sm">
                            <thead className="bg-muted/50 text-muted-foreground text-left">
                                <tr>
                                    <th className="px-3 py-2 font-medium">
                                        Type
                                    </th>
                                    <th className="px-3 py-2 font-medium">
                                        Reason
                                    </th>
                                    <th className="px-3 py-2 font-medium">
                                        Message
                                    </th>
                                    <th className="px-3 py-2 text-right font-medium">
                                        Count
                                    </th>
                                    <th className="px-3 py-2 text-right font-medium whitespace-nowrap">
                                        Last seen
                                    </th>
                                </tr>
                            </thead>
                            <tbody>
                                {events.map((e) => {
                                    const warning = e.type === "Warning";
                                    const ago = age(e.last, capturedAt);
                                    return (
                                        <tr
                                            key={e.key}
                                            className={cn(
                                                "border-t align-top",
                                                warning && "bg-amber-500/5"
                                            )}
                                        >
                                            <td
                                                className={cn(
                                                    "px-3 py-2 whitespace-nowrap",
                                                    warning &&
                                                        "font-medium text-amber-600 dark:text-amber-400"
                                                )}
                                            >
                                                {warning && (
                                                    <TriangleAlert
                                                        className="mr-1 inline size-4 align-text-bottom"
                                                        aria-hidden
                                                    />
                                                )}
                                                {e.type}
                                            </td>
                                            <td className="px-3 py-2 font-medium">
                                                {e.reason}
                                            </td>
                                            <td className="min-w-64 px-3 py-2 break-words">
                                                {e.message}
                                            </td>
                                            <td className="px-3 py-2 text-right tabular-nums">
                                                {e.count > 1
                                                    ? `×${e.count}`
                                                    : 1}
                                            </td>
                                            <td
                                                className="text-muted-foreground px-3 py-2 text-right whitespace-nowrap tabular-nums"
                                                title={
                                                    e.last
                                                        ? new Date(
                                                              e.last
                                                          ).toLocaleString()
                                                        : undefined
                                                }
                                            >
                                                {ago ? `${ago} ago` : "–"}
                                            </td>
                                        </tr>
                                    );
                                })}
                            </tbody>
                        </table>
                    </div>
                );
        }
    }

    return (
        <section aria-labelledby="events" className="space-y-3">
            <h3 id="events" className="text-lg font-semibold">
                Events
            </h3>
            {body}
        </section>
    );
}
