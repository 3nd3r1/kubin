import { useOutletContext } from "react-router";

import type { Snapshot } from "./api";

/** The snapshot loaded by SnapshotLayout, for pages under /s/:id. */
export function useSnapshot(): Snapshot {
    return useOutletContext<Snapshot>();
}

export function podPath(id: string, p: { namespace: string; name: string }) {
    return `/s/${id}/pods/${encodeURIComponent(p.namespace)}/${encodeURIComponent(p.name)}`;
}
