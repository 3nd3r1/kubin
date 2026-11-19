import type { Event, ObjectMeta } from "./k8s";

export type ResourceEvent = {
    key: string;
    type: string;
    reason: string;
    message: string;
    count: number;
    /** When it last happened, as an ISO time, or "" if the event has none. */
    last: string;
    source: string;
};

/** When an event last happened, from whichever field its API version set. */
export function lastSeen(e: Event): string {
    return (
        e.series?.lastObservedTime ||
        e.lastTimestamp ||
        e.eventTime ||
        e.firstTimestamp ||
        e.metadata?.creationTimestamp ||
        ""
    );
}

/**
 * The events about one object, newest first. Events are matched by kind,
 * namespace and name, and by UID when both sides have one, so an object
 * recreated under the same name doesn't pick up its predecessor's events.
 * Cluster-scoped objects have no namespace, and neither does the
 * involvedObject of their events.
 */
export function resourceEvents(
    object: { kind?: string; metadata?: ObjectMeta },
    events: Event[]
): ResourceEvent[] {
    const { name, uid } = object.metadata ?? {};
    const namespace = object.metadata?.namespace ?? "";
    return (
        events
            .filter((e) => {
                const o = e.involvedObject;
                if (!o || o.kind !== object.kind || o.name !== name)
                    return false;
                if ((o.namespace ?? "") !== namespace) return false;
                return !uid || !o.uid || o.uid === uid;
            })
            .map((e, i) => ({
                key: e.metadata?.uid || e.metadata?.name || String(i),
                type: e.type || "Normal",
                reason: e.reason || "",
                message: (e.message || "").trim(),
                count: e.series?.count || e.count || 1,
                last: lastSeen(e),
                source: e.reportingComponent || e.source?.component || "",
            }))
            // Times differ in precision between event versions, so compare
            // them as dates. Events without a time go last.
            .sort(
                (a, b) => (Date.parse(b.last) || 0) - (Date.parse(a.last) || 0)
            )
    );
}

const hour = 60 * 60 * 1000;

/**
 * The Warning events from the hour before capturedAt, newest first. The hour
 * is measured from the capture, not from now. Events without a time, and
 * events from after the capture, are left out.
 */
export function recentWarnings<T extends { object: Event }>(
    items: T[],
    capturedAt: string
): T[] {
    const end = Date.parse(capturedAt);
    if (Number.isNaN(end)) return [];
    return items
        .map((item) => ({ item, t: Date.parse(lastSeen(item.object)) }))
        .filter(
            ({ item, t }) =>
                item.object.type === "Warning" &&
                !Number.isNaN(t) &&
                t <= end &&
                t >= end - hour
        )
        .sort((a, b) => b.t - a.t)
        .map(({ item }) => item);
}
