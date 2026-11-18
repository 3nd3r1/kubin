import { describe, expect, it } from "vitest";

import type { LogFile, Manifest } from "./api";
import type { ContainerStatus, Pod } from "./k8s";
import { defaultLog, podLogs } from "./logs";

function file(container: string, previous = false, pod = "api-1"): LogFile {
    return { namespace: "shop", pod, container, previous, truncated: false };
}

function manifest(files: LogFile[]): Manifest {
    return {
        format_version: 3,
        created_at: "2026-11-17T12:00:00Z",
        cli_version: "dev",
        cluster: { context: "kind", name: "kind", server_version: "v1.33.1" },
        logs: { tail_lines: 1000, max_bytes: 1 << 20, files },
    };
}

function pod(statuses: ContainerStatus[]): Pod {
    return {
        metadata: { name: "api-1", namespace: "shop" },
        spec: {
            initContainers: [{ name: "migrate" }],
            containers: [{ name: "app" }, { name: "sidecar" }],
        },
        status: {
            initContainerStatuses: statuses.filter((s) => s.name === "migrate"),
            containerStatuses: statuses.filter((s) => s.name !== "migrate"),
        },
    };
}

const running = { running: {} };
const crashLoop = { waiting: { reason: "CrashLoopBackOff" } };
const failed = { terminated: { exitCode: 1 } };
const done = { terminated: { exitCode: 0 } };

describe("podLogs", () => {
    it("groups this pod's files by container in spec order", () => {
        const m = manifest([
            file("sidecar"),
            file("app", true),
            file("app"),
            file("migrate"),
            file("app", false, "other-pod"),
        ]);
        expect(podLogs(m, pod([]))).toEqual([
            { container: "migrate", current: file("migrate") },
            {
                container: "app",
                current: file("app"),
                previous: file("app", true),
            },
            { container: "sidecar", current: file("sidecar") },
        ]);
    });

    it("is empty for a snapshot without logs", () => {
        const m = manifest([]);
        delete m.logs;
        expect(podLogs(m, pod([]))).toEqual([]);
    });
});

describe("defaultLog", () => {
    const statuses = (
        app: ContainerStatus["state"],
        extra: Partial<ContainerStatus> = {}
    ) => [
        { name: "migrate", state: done },
        { name: "app", state: app, ...extra },
        { name: "sidecar", state: running },
    ];

    it("opens the last run of a crash-looping container", () => {
        const p = pod(
            statuses(crashLoop, { restartCount: 5, lastState: failed })
        );
        const logs = podLogs(
            manifest([file("migrate"), file("app"), file("sidecar")]),
            p
        );
        expect(defaultLog(p, logs)).toEqual({
            file: file("app"),
            crashing: true,
        });
    });

    it("falls back to the previous log when only that was captured", () => {
        const p = pod(statuses(crashLoop, { restartCount: 5 }));
        const logs = podLogs(manifest([file("app", true), file("sidecar")]), p);
        expect(defaultLog(p, logs)).toEqual({
            file: file("app", true),
            crashing: true,
        });
    });

    it("treats a container that exited non-zero as crashing", () => {
        const p = pod(statuses(failed));
        const logs = podLogs(manifest([file("app"), file("sidecar")]), p);
        expect(defaultLog(p, logs)?.crashing).toBe(true);
    });

    it("opens the previous log of a container that failed and is running again", () => {
        const p = pod(
            statuses(running, { restartCount: 1, lastState: failed })
        );
        const logs = podLogs(
            manifest([
                file("migrate"),
                file("app"),
                file("app", true),
                file("sidecar"),
            ]),
            p
        );
        expect(defaultLog(p, logs)).toEqual({
            file: file("app", true),
            crashing: false,
        });
    });

    it("opens the first container's log for a healthy pod", () => {
        const p = pod(statuses(running));
        const logs = podLogs(manifest([file("app"), file("migrate")]), p);
        expect(defaultLog(p, logs)).toEqual({
            file: file("migrate"),
            crashing: false,
        });
    });

    it("has nothing to open without logs", () => {
        expect(defaultLog(pod(statuses(crashLoop)), [])).toBeUndefined();
    });
});
