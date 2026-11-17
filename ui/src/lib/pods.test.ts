import { describe, expect, it } from "vitest";

import type { Pod } from "./k8s";
import { age, byNamespace, problems, summarize } from "./pods";

const captured = "2026-11-17T12:00:00Z";

function pod(over: Pod["status"], spec: Pod["spec"] = {}): Pod {
    return {
        metadata: {
            name: "p",
            namespace: "ns",
            creationTimestamp: "2026-11-17T11:00:00Z",
        },
        spec: { containers: [{ name: "app" }], ...spec },
        status: over,
    };
}

const readyCond = (s: string) => [{ type: "Ready", status: s }];

describe("summarize", () => {
    it("reports a healthy pod", () => {
        const s = summarize(
            pod({
                phase: "Running",
                conditions: readyCond("True"),
                containerStatuses: [
                    {
                        name: "app",
                        ready: true,
                        restartCount: 0,
                        state: { running: {} },
                    },
                ],
            }),
            captured
        );
        expect(s).toMatchObject({
            status: "Running",
            ready: "1/1",
            restarts: 0,
            age: "60m",
            problem: null,
        });
    });

    it("flags CrashLoopBackOff as an error", () => {
        const s = summarize(
            pod({
                phase: "Running",
                conditions: readyCond("False"),
                containerStatuses: [
                    {
                        name: "app",
                        ready: false,
                        restartCount: 6,
                        state: { waiting: { reason: "CrashLoopBackOff" } },
                        lastState: {
                            terminated: { reason: "Error", exitCode: 1 },
                        },
                    },
                ],
            }),
            captured
        );
        expect(s).toMatchObject({
            status: "CrashLoopBackOff",
            ready: "0/1",
            restarts: 6,
            problem: "error",
        });
    });

    it("flags image pull failures as errors", () => {
        const s = summarize(
            pod({
                phase: "Pending",
                containerStatuses: [
                    {
                        name: "app",
                        state: { waiting: { reason: "ImagePullBackOff" } },
                    },
                ],
            }),
            captured
        );
        expect(s.status).toBe("ImagePullBackOff");
        expect(s.problem).toBe("error");
    });

    it("flags an unscheduled pod as a warning", () => {
        const s = summarize(
            pod({
                phase: "Pending",
                conditions: [{ type: "PodScheduled", status: "False" }],
            }),
            captured
        );
        expect(s).toMatchObject({
            status: "Pending",
            ready: "0/1",
            problem: "warning",
        });
    });

    it("flags a running pod that restarted as a warning", () => {
        const s = summarize(
            pod({
                phase: "Running",
                conditions: readyCond("True"),
                containerStatuses: [
                    {
                        name: "app",
                        ready: true,
                        restartCount: 2,
                        state: { running: {} },
                    },
                ],
            }),
            captured
        );
        expect(s.problem).toBe("warning");
    });

    it("treats a completed job pod as fine", () => {
        const s = summarize(
            pod({
                phase: "Succeeded",
                containerStatuses: [
                    {
                        name: "app",
                        state: {
                            terminated: { reason: "Completed", exitCode: 0 },
                        },
                    },
                ],
            }),
            captured
        );
        expect(s).toMatchObject({ status: "Completed", problem: null });
    });

    it("reports a failing init container", () => {
        const s = summarize(
            pod(
                {
                    phase: "Pending",
                    initContainerStatuses: [
                        {
                            name: "migrate",
                            state: { terminated: { exitCode: 2 } },
                        },
                    ],
                },
                {
                    containers: [{ name: "app" }],
                    initContainers: [{ name: "migrate" }],
                }
            ),
            captured
        );
        expect(s).toMatchObject({
            status: "Init:ExitCode:2",
            problem: "error",
        });
    });

    it("ignores running sidecar init containers", () => {
        const s = summarize(
            pod(
                {
                    phase: "Running",
                    conditions: readyCond("True"),
                    initContainerStatuses: [
                        { name: "proxy", state: { running: {} } },
                    ],
                    containerStatuses: [
                        { name: "app", ready: true, state: { running: {} } },
                    ],
                },
                {
                    containers: [{ name: "app" }],
                    initContainers: [
                        { name: "proxy", restartPolicy: "Always" },
                    ],
                }
            ),
            captured
        );
        expect(s.status).toBe("Running");
    });

    it("shows Terminating for a pod being deleted", () => {
        const p = pod({ phase: "Running" });
        p.metadata!.deletionTimestamp = captured;
        expect(summarize(p, captured).status).toBe("Terminating");
    });
});

describe("age", () => {
    it.each([
        ["2026-11-17T11:59:15Z", "45s"],
        ["2026-11-17T11:30:00Z", "30m"],
        ["2026-11-17T07:00:00Z", "5h"],
        ["2026-11-10T12:00:00Z", "7d"],
        [undefined, ""],
    ])("from %s", (from, want) => {
        expect(age(from, captured)).toBe(want);
    });
});

describe("ordering", () => {
    const s = (
        namespace: string,
        name: string,
        problem: "error" | "warning" | null
    ) => ({ namespace, name, problem }) as never;

    it("puts errors before warnings", () => {
        const got = problems([
            s("a", "warn", "warning"),
            s("b", "ok", null),
            s("z", "err", "error"),
        ]);
        expect(got.map((p: { name: string }) => p.name)).toEqual([
            "err",
            "warn",
        ]);
    });

    it("groups by namespace with problem pods first", () => {
        const got = byNamespace([
            s("b", "x", null),
            s("a", "healthy", null),
            s("a", "broken", "error"),
        ]);
        expect(
            got.map(([ns, ps]) => [ns, ps.map((p: { name: string }) => p.name)])
        ).toEqual([
            ["a", ["broken", "healthy"]],
            ["b", ["x"]],
        ]);
    });
});
