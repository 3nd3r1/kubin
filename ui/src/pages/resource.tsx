import { Link, useParams } from "react-router";

import { getResource } from "@/lib/api";
import type { ObjectMeta } from "@/lib/k8s";
import { resolveSegment } from "@/lib/kinds";
import { age } from "@/lib/pods";
import { kindPath, useSnapshot } from "@/lib/snapshot";
import { useFetch } from "@/lib/use-fetch";

import { Badge } from "@/components/ui/badge";

import { LoadError, Loading } from "@/components/common/states";
import { ResourceEvents } from "@/components/resource/events";
import { OwnerLinks, RelatedResources } from "@/components/resource/related";
import { YamlView } from "@/components/resource/yaml";

import { ExpiredPage, NotFoundPage } from "./not-found";

type Obj = { kind?: string; metadata?: ObjectMeta };

/** The detail page of any kind: /s/{id}/{resource}/{namespace}/{name} */
export function ResourcePage() {
    const { resource = "", ns = "", name = "" } = useParams();
    const snapshot = useSnapshot();
    const found = resolveSegment(resource, snapshot.counts);
    const kind = found?.kind ?? "";
    const namespace = ns === "_cluster" ? "" : ns;
    const state = useFetch(
        (signal) =>
            kind
                ? getResource<Obj>(snapshot.id, kind, namespace, name, signal)
                : Promise.reject(new Error("unknown kind")),
        [snapshot.id, kind, namespace, name]
    );

    if (!found) return <NotFoundPage />;
    switch (state.status) {
        case "loading":
            return <Loading what={kind} />;
        case "missing":
            return (
                <NotFoundPage
                    title={`${kind} not found`}
                    detail={`There is no ${kind} ${namespace ? `${namespace}/` : ""}${name} in this snapshot.`}
                />
            );
        case "expired":
            return <ExpiredPage />;
        case "error":
            return <LoadError message={state.message} />;
    }

    const obj = state.data;
    const meta = obj.metadata ?? {};
    const labels = Object.entries(meta.labels ?? {});
    const annotations = Object.entries(meta.annotations ?? {});
    return (
        <div className="space-y-6">
            <Link
                to={kindPath(snapshot.id, resource, namespace)}
                className="text-muted-foreground text-sm hover:underline"
            >
                ← All {kind} resources
            </Link>
            <div className="space-y-2">
                <h2 className="font-mono text-xl font-semibold break-all">
                    {name}
                </h2>
                <dl className="grid grid-cols-2 gap-x-6 gap-y-3 text-sm sm:grid-cols-3">
                    <Field label="Kind" value={kind} />
                    {namespace && (
                        <Field label="Namespace" value={namespace} mono />
                    )}
                    {(meta.ownerReferences?.length ?? 0) > 0 && (
                        <Field label="Owner" mono>
                            <OwnerLinks snapshot={snapshot} object={obj} />
                        </Field>
                    )}
                    <Field
                        label="Age"
                        value={
                            age(
                                meta.creationTimestamp,
                                snapshot.manifest.created_at
                            ) || "–"
                        }
                    />
                </dl>
            </div>

            <Pairs id="labels" title="Labels" entries={labels} />
            <Pairs id="annotations" title="Annotations" entries={annotations} />

            <RelatedResources snapshot={snapshot} object={{ ...obj, kind }} />

            <ResourceEvents
                snapshotId={snapshot.id}
                object={obj}
                capturedAt={snapshot.manifest.created_at}
            />
            <YamlView object={obj} manifest={snapshot.manifest} />
        </div>
    );
}

function Field({
    label,
    value,
    children,
    mono,
}: {
    label: string;
    value?: string;
    children?: React.ReactNode;
    mono?: boolean;
}) {
    return (
        <div className="min-w-0">
            <dt className="text-muted-foreground">{label}</dt>
            <dd
                className={mono ? "truncate font-mono" : undefined}
                title={value}
            >
                {children ?? value}
            </dd>
        </div>
    );
}

function Pairs({
    id,
    title,
    entries,
}: {
    id: string;
    title: string;
    entries: [string, string][];
}) {
    if (entries.length === 0) return null;
    return (
        <section aria-labelledby={id} className="space-y-2">
            <h3 id={id} className="text-lg font-semibold">
                {title}
            </h3>
            <div className="flex flex-wrap gap-2">
                {entries.map(([k, v]) => (
                    <Badge
                        key={k}
                        variant="secondary"
                        className="max-w-full font-mono break-all whitespace-normal"
                    >
                        {k}={v}
                    </Badge>
                ))}
            </div>
        </section>
    );
}
