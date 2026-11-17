import { useEffect, useState } from "react";

import { useParams } from "react-router";

import { NotFoundError, type Snapshot, getSnapshot } from "@/lib/api";

import { Badge } from "@/components/ui/badge";

import { NotFoundPage } from "./not-found";

type State =
    | { status: "loading" }
    | { status: "ok"; snapshot: Snapshot }
    | { status: "missing" }
    | { status: "error"; message: string };

// A first view of a snapshot. TASK-6 turns this into the real page.
export function SnapshotPage() {
    const { id = "" } = useParams();
    const [state, setState] = useState<State>({ status: "loading" });

    useEffect(() => {
        const ctrl = new AbortController();
        getSnapshot(id, ctrl.signal)
            .then((snapshot) => setState({ status: "ok", snapshot }))
            .catch((err: unknown) => {
                if (ctrl.signal.aborted) return;
                if (err instanceof NotFoundError)
                    setState({ status: "missing" });
                else setState({ status: "error", message: String(err) });
            });
        return () => ctrl.abort();
    }, [id]);

    switch (state.status) {
        case "loading":
            return <p className="text-muted-foreground">Loading snapshot…</p>;
        case "missing":
            return (
                <NotFoundPage
                    title="Snapshot not found"
                    detail="The link is wrong, or the snapshot has expired."
                />
            );
        case "error":
            return <p className="text-destructive">{state.message}</p>;
    }

    const { manifest, counts } = state.snapshot;
    return (
        <div className="space-y-4">
            <div>
                <h1 className="text-2xl font-semibold">
                    {manifest.cluster.name}
                </h1>
                <p className="text-muted-foreground text-sm">
                    Kubernetes {manifest.cluster.server_version}, captured{" "}
                    {new Date(manifest.created_at).toLocaleString()}
                </p>
            </div>
            <div className="flex flex-wrap gap-2">
                {counts.map((c) => (
                    <Badge
                        key={`${c.group}/${c.version}/${c.kind}`}
                        variant="secondary"
                    >
                        {c.kind} {c.count}
                    </Badge>
                ))}
            </div>
        </div>
    );
}
