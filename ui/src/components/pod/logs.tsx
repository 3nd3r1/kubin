import { useEffect, useLayoutEffect, useMemo, useRef, useState } from "react";

import { useSearchParams } from "react-router";

import { type LogFile, type Manifest, getLog } from "@/lib/api";
import type { Pod } from "@/lib/k8s";
import {
    defaultLog,
    errorLine,
    formatLogParams,
    parseLogParams,
    podLogs,
    visibleRange,
} from "@/lib/logs";
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
    const [params, setParams] = useSearchParams();
    const wanted = useMemo(() => parseLogParams(params), [params]);
    const fromUrl = useMemo(() => {
        const c = logs.find((l) => l.container === wanted.container);
        return c && (wanted.previous ? c.previous : c.current);
    }, [logs, wanted]);
    const selected = fromUrl ?? initial?.file;
    const line = fromUrl ? wanted.line : undefined;
    const section = useRef<HTMLElement>(null);

    const select = (f: LogFile, n?: number) =>
        setParams(
            formatLogParams(params, {
                container: f.container,
                previous: f.previous,
                line: n,
            }),
            { replace: true, preventScrollReset: true }
        );

    useEffect(() => {
        if (initial?.crashing && !fromUrl) {
            section.current?.scrollIntoView({ block: "start" });
        }
        // Only on arrival: choosing a log later must not scroll the page.
        // eslint-disable-next-line react-hooks/exhaustive-deps
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
                                    select((c.current ?? c.previous)!)
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
                                onClick={() => select(current.current!)}
                            >
                                Current
                            </Choice>
                            <Choice
                                active={selected.previous}
                                onClick={() => select(current.previous!)}
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
                    line={line}
                    onLine={(n) => select(selected, n)}
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

const LINE_HEIGHT = 20;
const OVERSCAN = 20;
const MAX_HEIGHT = 512;

function LogView({
    snapshotId,
    file,
    tailLines,
    line,
    onLine,
}: {
    snapshotId: string;
    file: LogFile;
    tailLines: number;
    line?: number;
    onLine: (n: number) => void;
}) {
    const state = useFetch(
        (signal) => getLog(snapshotId, file, signal),
        [snapshotId, file.namespace, file.pod, file.container, file.previous]
    );
    const box = useRef<HTMLDivElement>(null);
    const [scrollTop, setScrollTop] = useState(0);
    const [viewport, setViewport] = useState(MAX_HEIGHT);

    const lines = useMemo(
        () =>
            state.status === "ok" && state.data !== ""
                ? state.data.replace(/\n$/, "").split("\n")
                : [],
        [state]
    );
    const height = Math.min(MAX_HEIGHT, lines.length * LINE_HEIGHT + 16);

    // Without a line, open at the end, where a crash shows up. With one,
    // centre it. Changing line alone moves the view but doesn't refetch.
    useLayoutEffect(() => {
        const el = box.current;
        if (!el || lines.length === 0) return;
        const target =
            line === undefined
                ? el.scrollHeight
                : (Math.min(line, lines.length) - 1) * LINE_HEIGHT -
                  (el.clientHeight - LINE_HEIGHT) / 2;
        el.scrollTop = Math.max(0, target);
        setScrollTop(el.scrollTop);
        setViewport(el.clientHeight);
    }, [lines, line]);

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

    const { start, end } = visibleRange(
        scrollTop,
        viewport,
        LINE_HEIGHT,
        lines.length,
        OVERSCAN
    );
    const digits = String(lines.length).length;
    return (
        <div className="space-y-1">
            {file.truncated && (
                <p className="text-muted-foreground text-xs">
                    Showing the last {Math.min(lines.length, tailLines)} lines.
                    Older lines weren't captured.
                </p>
            )}
            <div
                ref={box}
                role="log"
                aria-label={`${file.container} ${file.previous ? "previous " : ""}log`}
                tabIndex={0}
                style={{ height }}
                onScroll={(e) => {
                    setScrollTop(e.currentTarget.scrollTop);
                    setViewport(e.currentTarget.clientHeight);
                }}
                className="bg-muted overflow-auto rounded-lg font-mono text-xs"
            >
                <div
                    className="relative w-max min-w-full"
                    style={{
                        height: lines.length * LINE_HEIGHT + 16,
                        paddingTop: 8,
                    }}
                >
                    <div
                        style={{
                            transform: `translateY(${start * LINE_HEIGHT}px)`,
                        }}
                    >
                        {lines.slice(start, end).map((l, k) => {
                            const n = start + k + 1;
                            return (
                                <div
                                    key={n}
                                    style={{ height: LINE_HEIGHT }}
                                    className={cn(
                                        "flex items-center whitespace-pre pr-4 leading-5",
                                        errorLine.test(l) &&
                                            "bg-red-500/10 text-red-700 dark:text-red-300",
                                        n === line &&
                                            "bg-yellow-400/30 text-foreground"
                                    )}
                                >
                                    <button
                                        type="button"
                                        onClick={() => onLine(n)}
                                        aria-label={`Link to line ${n}`}
                                        style={{ minWidth: `${digits + 2}ch` }}
                                        className="text-muted-foreground hover:text-foreground sticky left-0 shrink-0 select-none bg-inherit px-2 text-right"
                                    >
                                        {n}
                                    </button>
                                    <span className="pl-2">{l || " "}</span>
                                </div>
                            );
                        })}
                    </div>
                </div>
            </div>
        </div>
    );
}
