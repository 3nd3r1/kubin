import { useEffect, useMemo, useRef, useState } from "react";

import { type LogFile, type Manifest, getLog } from "@/lib/api";
import type { Pod } from "@/lib/k8s";
import { defaultLog, errorLine, podLogs } from "@/lib/logs";
import { useFetch } from "@/lib/use-fetch";
import { cn } from "@/lib/utils";

import { Empty, LoadError, Loading } from "@/components/common/states";

/**
 * The pod's captured logs, one container at a time. It opens on the log
 * most likely to explain a problem, and when the pod is crash-looping it
 * scrolls itself into view so the crash is the first thing on screen.
 */
export function PodLogs({
    snapshotId,
    manifest,
    pod,
}: {
    snapshotId: string;
    manifest: Manifest;
    pod: Pod;
}) {
    const logs = useMemo(() => podLogs(manifest, pod), [manifest, pod]);
    const initial = useMemo(() => defaultLog(pod, logs), [pod, logs]);
    const [selected, setSelected] = useState<LogFile | undefined>(
        initial?.file
    );
    const section = useRef<HTMLElement>(null);

    useEffect(() => {
        if (initial?.crashing) {
            section.current?.scrollIntoView({ block: "start" });
        }
    }, [initial]);

    let body;
    if (!manifest.logs) {
        body = (
            <Empty>
                This snapshot has no logs: it was taken by a kubin CLI that
                didn't collect them.
            </Empty>
        );
    } else if (logs.length === 0 || !selected) {
        body = (
            <Empty>
                No logs were captured for this pod. Kubin collects logs only
                from containers that have started.
            </Empty>
        );
    } else {
        const current = logs.find((c) => c.container === selected.container);
        body = (
            <div className="space-y-2">
                <div className="flex flex-wrap items-center gap-2">
                    <div
                        role="tablist"
                        aria-label="Container"
                        className="flex flex-wrap gap-1"
                    >
                        {logs.map((c) => (
                            <Choice
                                key={c.container}
                                active={c.container === selected.container}
                                onClick={() =>
                                    setSelected((c.current ?? c.previous)!)
                                }
                                mono
                            >
                                {c.container}
                            </Choice>
                        ))}
                    </div>
                    {current?.current && current.previous && (
                        <div
                            role="tablist"
                            aria-label="Run"
                            className="ml-auto flex gap-1"
                        >
                            <Choice
                                active={!selected.previous}
                                onClick={() => setSelected(current.current!)}
                            >
                                Current
                            </Choice>
                            <Choice
                                active={selected.previous}
                                onClick={() => setSelected(current.previous!)}
                            >
                                Previous run
                            </Choice>
                        </div>
                    )}
                </div>
                <LogView
                    snapshotId={snapshotId}
                    file={selected}
                    tailLines={manifest.logs.tail_lines}
                />
            </div>
        );
    }

    return (
        <section
            ref={section}
            aria-labelledby="logs"
            className="scroll-mt-4 space-y-3"
        >
            <h3 id="logs" className="text-lg font-semibold">
                Logs
            </h3>
            {body}
        </section>
    );
}

function Choice({
    active,
    onClick,
    mono,
    children,
}: {
    active: boolean;
    onClick: () => void;
    mono?: boolean;
    children: React.ReactNode;
}) {
    return (
        <button
            type="button"
            role="tab"
            aria-selected={active}
            onClick={onClick}
            className={cn(
                "rounded-md border px-2.5 py-1 text-sm transition-colors",
                mono && "font-mono",
                active
                    ? "bg-primary text-primary-foreground border-primary"
                    : "hover:bg-muted"
            )}
        >
            {children}
        </button>
    );
}

function LogView({
    snapshotId,
    file,
    tailLines,
}: {
    snapshotId: string;
    file: LogFile;
    tailLines: number;
}) {
    const state = useFetch(
        (signal) => getLog(snapshotId, file, signal),
        [snapshotId, file.namespace, file.pod, file.container, file.previous]
    );
    const pre = useRef<HTMLPreElement>(null);

    // The end of a log is where a crash shows up.
    useEffect(() => {
        if (state.status === "ok" && pre.current) {
            pre.current.scrollTop = pre.current.scrollHeight;
        }
    }, [state]);

    switch (state.status) {
        case "loading":
            return <Loading what="log" />;
        case "missing":
            return <Empty>This log isn't stored on the server.</Empty>;
        case "error":
            return <LoadError message={state.message} />;
    }
    if (state.data === "") {
        return <Empty>This container wrote nothing to its log.</Empty>;
    }

    const lines = state.data.replace(/\n$/, "").split("\n");
    return (
        <div className="space-y-1">
            {file.truncated && (
                <p className="text-muted-foreground text-xs">
                    Showing the last {Math.min(lines.length, tailLines)} lines.
                    Older lines weren't captured.
                </p>
            )}
            <pre
                ref={pre}
                aria-label={`${file.container} ${file.previous ? "previous " : ""}log`}
                tabIndex={0}
                className="bg-muted max-h-[32rem] overflow-auto rounded-lg p-4 font-mono text-xs leading-relaxed"
            >
                {lines.map((l, i) => (
                    <div
                        key={i}
                        className={cn(
                            "whitespace-pre-wrap break-all",
                            errorLine.test(l) &&
                                "-mx-4 bg-red-500/10 px-4 text-red-700 dark:text-red-300"
                        )}
                    >
                        {l || " "}
                    </div>
                ))}
            </pre>
        </div>
    );
}
