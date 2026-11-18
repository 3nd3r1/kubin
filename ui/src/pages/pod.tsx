import { useMemo } from "react";

import { Link, useParams } from "react-router";
import { stringify } from "yaml";

import { type Snapshot, getResource } from "@/lib/api";
import type { ContainerState, ContainerStatus, Pod } from "@/lib/k8s";
import { summarize } from "@/lib/pods";
import { useSnapshot } from "@/lib/snapshot";
import { useFetch } from "@/lib/use-fetch";

import { Badge } from "@/components/ui/badge";

import { LoadError, Loading } from "@/components/common/states";
import { PodStatus } from "@/components/common/status";
import { PodEvents } from "@/components/pod/events";
import { PodLogs } from "@/components/pod/logs";

import { NotFoundPage } from "./not-found";

export function PodPage() {
    const snapshot = useSnapshot();
    const { ns = "", name = "" } = useParams();
    const state = useFetch(
        (signal) => getResource<Pod>(snapshot.id, "Pod", ns, name, signal),
        [snapshot.id, ns, name]
    );

    const back = (
        <Link
            to={`/s/${snapshot.id}`}
            className="text-muted-foreground text-sm hover:underline"
        >
            ← All pods
        </Link>
    );

    switch (state.status) {
        case "loading":
            return <Loading what="pod" />;
        case "missing":
            return (
                <NotFoundPage
                    title="Pod not found"
                    detail={`There is no pod ${ns}/${name} in this snapshot.`}
                />
            );
        case "error":
            return <LoadError message={state.message} />;
    }

    return (
        <div className="space-y-6">
            {back}
            <PodDetail pod={state.data} snapshot={snapshot} />
        </div>
    );
}

function PodDetail({ pod, snapshot }: { pod: Pod; snapshot: Snapshot }) {
    const capturedAt = snapshot.manifest.created_at;
    const s = summarize(pod, capturedAt);
    const podKey = `${s.namespace}/${s.name}`;
    const yaml = useMemo(() => stringify(pod, { lineWidth: 0 }), [pod]);
    const statuses = new Map(
        [
            ...(pod.status?.initContainerStatuses ?? []),
            ...(pod.status?.containerStatuses ?? []),
        ].map((c) => [c.name, c])
    );
    const containers = [
        ...(pod.spec?.initContainers ?? []).map((c) => ({ ...c, init: true })),
        ...(pod.spec?.containers ?? []).map((c) => ({ ...c, init: false })),
    ];
    const owner = pod.metadata?.ownerReferences?.[0];

    return (
        <>
            <div className="space-y-2">
                <h2 className="font-mono text-xl font-semibold break-all">
                    {s.name}
                </h2>
                <PodStatus
                    status={s.status}
                    problem={s.problem}
                    className="text-base"
                />
                {pod.status?.message && (
                    <p className="text-muted-foreground text-sm">
                        {pod.status.message}
                    </p>
                )}
            </div>

            <dl className="grid grid-cols-2 gap-x-6 gap-y-3 text-sm sm:grid-cols-3 lg:grid-cols-6">
                <Field label="Namespace" value={s.namespace} mono />
                <Field label="Ready" value={s.ready} />
                <Field label="Restarts" value={String(s.restarts)} />
                <Field label="Age" value={s.age || "–"} />
                <Field label="Node" value={s.node || "not scheduled"} mono />
                <Field
                    label="Owner"
                    value={owner ? `${owner.kind}/${owner.name}` : "–"}
                    mono
                />
            </dl>

            <section aria-labelledby="containers" className="space-y-3">
                <h3 id="containers" className="text-lg font-semibold">
                    Containers
                </h3>
                <div className="grid gap-3 md:grid-cols-2">
                    {containers.map((c) => (
                        <ContainerCard
                            key={c.name}
                            name={c.name}
                            image={c.image}
                            init={c.init}
                            status={statuses.get(c.name)}
                        />
                    ))}
                </div>
            </section>

            <PodLogs
                key={podKey}
                snapshotId={snapshot.id}
                manifest={snapshot.manifest}
                pod={pod}
            />

            <PodEvents
                snapshotId={snapshot.id}
                pod={pod}
                capturedAt={capturedAt}
            />

            <section aria-labelledby="yaml" className="space-y-3">
                <h3 id="yaml" className="text-lg font-semibold">
                    YAML
                </h3>
                <pre className="bg-muted max-h-[32rem] overflow-auto rounded-lg p-4 font-mono text-xs leading-relaxed">
                    {yaml}
                </pre>
            </section>
        </>
    );
}

function Field({
    label,
    value,
    mono,
}: {
    label: string;
    value: string;
    mono?: boolean;
}) {
    return (
        <div className="min-w-0">
            <dt className="text-muted-foreground">{label}</dt>
            <dd
                className={mono ? "truncate font-mono" : undefined}
                title={value}
            >
                {value}
            </dd>
        </div>
    );
}

function describeState(s: ContainerState | undefined): {
    label: string;
    detail?: string;
    bad: boolean;
} {
    if (s?.running)
        return {
            label: "Running",
            detail: since(s.running.startedAt),
            bad: false,
        };
    if (s?.waiting)
        return {
            label: `Waiting: ${s.waiting.reason ?? "unknown"}`,
            detail: s.waiting.message,
            bad:
                s.waiting.reason !== "ContainerCreating" &&
                s.waiting.reason !== "PodInitializing",
        };
    if (s?.terminated) {
        const t = s.terminated;
        return {
            label: `Terminated: ${t.reason ?? "unknown"}, exit code ${t.exitCode ?? "?"}`,
            detail:
                t.message ??
                (t.finishedAt
                    ? `at ${new Date(t.finishedAt).toLocaleString()}`
                    : undefined),
            bad: (t.exitCode ?? 0) !== 0,
        };
    }
    return { label: "No status", bad: false };
}

function since(t?: string) {
    return t ? `since ${new Date(t).toLocaleString()}` : undefined;
}

function ContainerCard({
    name,
    image,
    init,
    status,
}: {
    name: string;
    image?: string;
    init: boolean;
    status?: ContainerStatus;
}) {
    const state = describeState(status?.state);
    const last = status?.lastState?.terminated;
    return (
        <div className="space-y-2 rounded-lg border p-4 text-sm">
            <div className="flex items-center justify-between gap-2">
                <span className="truncate font-mono font-medium">{name}</span>
                <div className="flex shrink-0 gap-1">
                    {init && <Badge variant="outline">init</Badge>}
                    {status && (
                        <Badge variant={status.ready ? "secondary" : "outline"}>
                            {status.ready ? "ready" : "not ready"}
                        </Badge>
                    )}
                </div>
            </div>
            <p
                className="text-muted-foreground truncate font-mono text-xs"
                title={image}
            >
                {image}
            </p>
            <p
                className={
                    state.bad
                        ? "font-medium text-red-600 dark:text-red-400"
                        : undefined
                }
            >
                {state.label}
            </p>
            {state.detail && (
                <p className="text-muted-foreground text-xs break-words">
                    {state.detail}
                </p>
            )}
            <p className="text-muted-foreground">
                Restarts:{" "}
                <span className="text-foreground tabular-nums">
                    {status?.restartCount ?? 0}
                </span>
            </p>
            {last && (
                <p className="text-muted-foreground">
                    Last termination:{" "}
                    <span className="text-foreground">
                        {last.reason ?? "unknown"}, exit code{" "}
                        {last.exitCode ?? "?"}
                    </span>
                    {last.finishedAt && (
                        <> at {new Date(last.finishedAt).toLocaleString()}</>
                    )}
                </p>
            )}
        </div>
    );
}
