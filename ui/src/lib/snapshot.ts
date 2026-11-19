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
