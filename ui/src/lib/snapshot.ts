import { useOutletContext } from "react-router";

import type { Snapshot } from "./api";

/** The snapshot loaded by SnapshotLayout, for pages under /s/:id. */
export function useSnapshot(): Snapshot {
    return useOutletContext<Snapshot>();
}

export function podPath(id: string, p: { namespace: string; name: string }) {
    return `/s/${id}/pods/${encodeURIComponent(p.namespace)}/${encodeURIComponent(p.name)}`;
}

/** The namespace filter in the URL: `?ns=`. Empty means all namespaces. */
export function nsParam(search: string | URLSearchParams): string {
    return new URLSearchParams(search).get("ns") ?? "";
}

function withNs(path: string, ns: string) {
    return ns ? `${path}?ns=${encodeURIComponent(ns)}` : path;
}

/** The list page of one kind, keeping the namespace filter. */
export function kindPath(id: string, segment: string, ns = "") {
    return withNs(`/s/${id}/${segment}`, ns);
}

/** A resource's detail page. Cluster-scoped kinds use `_cluster`. */
export function resourcePath(
    id: string,
    segment: string,
    r: { namespace: string; name: string }
) {
    const parts = [r.namespace || "_cluster", r.name].map(encodeURIComponent);
    return `/s/${id}/${segment}/${parts.join("/")}`;
}

const MINUTE_MS = 60_000;
const HOUR_MS = 60 * MINUTE_MS;
const DAY_MS = 24 * HOUR_MS;

/**
 * Relative expiry text for the snapshot header, e.g. "expires in 3 days".
 * Returns null when expiresAt is absent or invalid, and "expired" when in the past.
 */
export function formatExpiry(expiresAt?: string): string | null {
    if (!expiresAt) return null;
    const ms = Date.parse(expiresAt);
    if (Number.isNaN(ms)) return null;
    const diff = ms - Date.now();
    if (diff <= 0) return "expired";
    const mins = Math.round(diff / MINUTE_MS);
    if (mins < 1) return "expires in less than a minute";
    if (mins < 60)
        return `expires in ${mins} ${mins === 1 ? "minute" : "minutes"}`;
    const hours = Math.round(diff / HOUR_MS);
    if (hours < 24)
        return `expires in ${hours} ${hours === 1 ? "hour" : "hours"}`;
    const days = Math.round(diff / DAY_MS);
    return `expires in ${days} ${days === 1 ? "day" : "days"}`;
}
