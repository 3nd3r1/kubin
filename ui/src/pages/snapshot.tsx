import { Link, Outlet, useParams } from "react-router";

import { getSnapshot } from "@/lib/api";
import { useFetch } from "@/lib/use-fetch";

import { LoadError, Loading } from "@/components/common/states";

import { NotFoundPage } from "./not-found";

/** Loads the snapshot once for every page under /s/:id. */
export function SnapshotLayout() {
    const { id = "" } = useParams();
    const state = useFetch((signal) => getSnapshot(id, signal), [id]);

    switch (state.status) {
        case "loading":
            return <Loading what="snapshot" />;
        case "missing":
            return (
                <NotFoundPage
                    title="Snapshot not found"
                    detail="The link is wrong, or the snapshot has expired."
                />
            );
        case "error":
            return <LoadError message={state.message} />;
    }

    const snapshot = state.data;
    const { cluster, created_at } = snapshot.manifest;
    return (
        <div className="space-y-6">
            <div>
                <h1 className="text-2xl font-semibold">
                    <Link to={`/s/${snapshot.id}`} className="hover:underline">
                        {cluster.name || cluster.context || "Cluster"}
                    </Link>
                </h1>
                <p className="text-muted-foreground text-sm">
                    Kubernetes {cluster.server_version}, captured{" "}
                    <time dateTime={created_at} title={created_at}>
                        {new Date(created_at).toLocaleString()}
                    </time>
                </p>
            </div>
            <Outlet context={snapshot} />
        </div>
    );
}
