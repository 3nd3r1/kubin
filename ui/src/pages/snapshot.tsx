import {
    Link,
    NavLink,
    Outlet,
    useLocation,
    useNavigate,
    useParams,
} from "react-router";

import { getSnapshot, listResources } from "@/lib/api";
import { groupKinds } from "@/lib/kinds";
import { formatExpiry, kindPath, nsParam } from "@/lib/snapshot";
import { useFetch } from "@/lib/use-fetch";
import { cn } from "@/lib/utils";

import { LoadError, Loading } from "@/components/common/states";

import { ExpiredPage, NotFoundPage } from "./not-found";

/** Loads the snapshot once for every page under /s/:id. */
export function SnapshotLayout() {
    const { id = "" } = useParams();
    const { search } = useLocation();
    const ns = nsParam(search);
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
        case "expired":
            return <ExpiredPage />;
        case "error":
            return <LoadError message={state.message} />;
    }

    const snapshot = state.data;
    const { cluster, created_at } = snapshot.manifest;
    const expiry = formatExpiry(snapshot.expires_at);
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
                    {expiry && snapshot.expires_at && (
                        <>
                            {" · "}
                            <time
                                dateTime={snapshot.expires_at}
                                title={snapshot.expires_at}
                            >
                                {expiry}
                            </time>
                        </>
                    )}
                </p>
            </div>
            <div className="flex flex-col gap-6 md:flex-row">
                <aside className="shrink-0 space-y-4 md:w-56">
                    <NamespaceSelect id={snapshot.id} ns={ns} />
                    <nav aria-label="Kinds" className="space-y-4">
                        {groupKinds(snapshot.counts).map((s) => (
                            <div key={s.title}>
                                <h2 className="text-muted-foreground mb-1 px-2 text-xs font-medium uppercase">
                                    {s.title}
                                </h2>
                                {s.items.map((i) => (
                                    <NavLink
                                        key={i.segment}
                                        to={kindPath(
                                            snapshot.id,
                                            i.segment,
                                            i.info?.namespaced === false
                                                ? ""
                                                : ns
                                        )}
                                        className={({ isActive }) =>
                                            cn(
                                                "hover:bg-muted flex justify-between rounded px-2 py-1 text-sm",
                                                isActive &&
                                                    "bg-muted font-medium"
                                            )
                                        }
                                    >
                                        {i.kind}
                                        <span className="text-muted-foreground tabular-nums">
                                            {i.count}
                                        </span>
                                    </NavLink>
                                ))}
                            </div>
                        ))}
                    </nav>
                </aside>
                <div className="min-w-0 flex-1">
                    <Outlet context={snapshot} />
                </div>
            </div>
        </div>
    );
}

/** Namespace choices come from the snapshot's Namespace resources. */
function NamespaceSelect({ id, ns }: { id: string; ns: string }) {
    const navigate = useNavigate();
    const { pathname } = useLocation();
    const state = useFetch(
        (signal) => listResources(id, { kind: "Namespace" }, signal),
        [id]
    );
    const names = state.status === "ok" ? state.data.map((r) => r.name) : [];
    // Keep a filter from the URL selectable even if it isn't listed.
    if (ns && !names.includes(ns)) names.push(ns);
    if (names.length === 0) return null;
    return (
        <label className="block text-sm">
            <span className="text-muted-foreground mb-1 block px-2 text-xs font-medium uppercase">
                Namespace
            </span>
            <select
                value={ns}
                onChange={(e) =>
                    navigate(
                        e.target.value
                            ? `${pathname}?ns=${encodeURIComponent(e.target.value)}`
                            : pathname
                    )
                }
                className="bg-background w-full rounded border px-2 py-1"
            >
                <option value="">All namespaces</option>
                {names.map((n) => (
                    <option key={n} value={n}>
                        {n}
                    </option>
                ))}
            </select>
        </label>
    );
}
