import { describe, expect, it } from "vitest";

import { columnsFor, problem, sortObjects } from "./columns";
import type { AnyObject } from "./k8s";

const captured = "2026-11-17T12:00:00Z";

function row(kind: string, o: AnyObject): Record<string, string> {
    return Object.fromEntries(
        columnsFor(kind).map((c) => [c.header, c.cell(o, captured)])
    );
}

const kinds = [
    "Deployment",
    "ReplicaSet",
    "StatefulSet",
    "DaemonSet",
    "Job",
    "CronJob",
    "Service",
    "Ingress",
    "EndpointSlice",
    "ConfigMap",
    "PersistentVolumeClaim",
    "PersistentVolume",
    "Node",
    "Namespace",
    "Event",
];

describe("columns", () => {
    it("never shows undefined or NaN for an empty object", () => {
        for (const kind of kinds) {
            expect(columnsFor(kind).length).toBeGreaterThan(0);
            for (const v of Object.values(row(kind, {}))) {
                expect(v).not.toMatch(/undefined|NaN/);
                expect(v).not.toBe("");
            }
        }
    });

    it("shows 0 and <none> without status", () => {
        expect(row("Deployment", { spec: { replicas: 3 } })).toEqual({
            Ready: "0/3",
            "Up-to-date": "0",
            Available: "0",
        });
        expect(row("Node", {})).toMatchObject({
            Status: "NotReady",
            Roles: "<none>",
            Version: "<none>",
        });
        expect(row("PersistentVolumeClaim", {}).Status).toBe("<none>");
    });

    it("covers workloads", () => {
        expect(
            row("Deployment", {
                spec: { replicas: 3 },
                status: {
                    readyReplicas: 2,
                    updatedReplicas: 3,
                    availableReplicas: 2,
                },
            })
        ).toEqual({ Ready: "2/3", "Up-to-date": "3", Available: "2" });
        expect(
            row("ReplicaSet", {
                spec: { replicas: 2 },
                status: { replicas: 2, readyReplicas: 1 },
            })
        ).toEqual({ Desired: "2", Current: "2", Ready: "1" });
        expect(
            row("StatefulSet", {
                spec: { replicas: 2 },
                status: { readyReplicas: 2 },
            })
        ).toEqual({ Ready: "2/2" });
        expect(
            row("DaemonSet", {
                spec: {
                    template: { spec: { nodeSelector: { b: "2", a: "1" } } },
                },
                status: {
                    desiredNumberScheduled: 3,
                    currentNumberScheduled: 3,
                    numberReady: 2,
                    updatedNumberScheduled: 3,
                    numberAvailable: 2,
                },
            })
        ).toEqual({
            Desired: "3",
            Current: "3",
            Ready: "2",
            "Up-to-date": "3",
            Available: "2",
            "Node selector": "a=1,b=2",
        });
    });

    it("covers jobs and cronjobs against capture time", () => {
        const start = "2026-11-17T11:00:00Z";
        expect(
            row("Job", {
                spec: { completions: 1 },
                status: {
                    succeeded: 1,
                    startTime: start,
                    completionTime: "2026-11-17T11:05:00Z",
                    conditions: [{ type: "Complete", status: "True" }],
                },
            })
        ).toEqual({ Status: "Complete", Completions: "1/1", Duration: "5m" });
        // Still running: duration runs to capture time, not now.
        expect(row("Job", { status: { startTime: start } })).toEqual({
            Status: "Running",
            Completions: "0/1",
            Duration: "60m",
        });
        expect(row("Job", { spec: { suspend: true } }).Status).toBe(
            "Suspended"
        );
        expect(
            row("CronJob", {
                spec: { schedule: "*/5 * * * *", suspend: false },
                status: { active: 1, lastScheduleTime: "2026-11-17T11:55:00Z" },
            })
        ).toEqual({
            Schedule: "*/5 * * * *",
            Timezone: "<none>",
            Suspend: "False",
            Active: "1",
            "Last schedule": "5m",
        });
    });

    it("covers network kinds", () => {
        expect(
            row("Service", {
                spec: {
                    type: "LoadBalancer",
                    clusterIP: "10.0.0.1",
                    ports: [{ port: 80, nodePort: 30080, protocol: "TCP" }],
                },
                status: { loadBalancer: { ingress: [{ ip: "1.2.3.4" }] } },
            })
        ).toEqual({
            Type: "LoadBalancer",
            "Cluster-IP": "10.0.0.1",
            "External-IP": "1.2.3.4",
            "Port(s)": "80:30080/TCP",
        });
        expect(
            row("Service", { spec: { type: "LoadBalancer" } })
        ).toMatchObject({ "External-IP": "<pending>" });
        expect(
            row("Ingress", {
                spec: {
                    ingressClassName: "nginx",
                    rules: [{ host: "a.example.com" }, {}],
                    tls: [{}],
                },
                status: { loadBalancer: { ingress: [{ hostname: "lb" }] } },
            })
        ).toEqual({
            Class: "nginx",
            Hosts: "a.example.com,*",
            Address: "lb",
            Ports: "80, 443",
        });
        expect(
            row("EndpointSlice", {
                addressType: "IPv4",
                ports: [{ port: 80 }, { port: 443 }],
                endpoints: [
                    { addresses: ["1.1.1.1"] },
                    { addresses: ["1.1.1.2"] },
                    { addresses: ["1.1.1.3"] },
                    { addresses: ["1.1.1.4"] },
                ],
            })
        ).toEqual({
            AddressType: "IPv4",
            Ports: "80,443",
            Endpoints: "1.1.1.1,1.1.1.2,1.1.1.3 + 1 more...",
        });
    });

    it("counts configmap data and binaryData together", () => {
        expect(
            row("ConfigMap", {
                data: { a: "1", b: "2" },
                binaryData: { c: "" },
            })
        ).toEqual({ Data: "3" });
    });

    it("covers storage kinds with short access modes", () => {
        expect(
            row("PersistentVolumeClaim", {
                spec: { volumeName: "pv1", storageClassName: "fast" },
                status: {
                    phase: "Bound",
                    capacity: { storage: "1Gi" },
                    accessModes: ["ReadWriteOnce", "ReadOnlyMany"],
                },
            })
        ).toEqual({
            Status: "Bound",
            Volume: "pv1",
            Capacity: "1Gi",
            "Access modes": "RWO,ROX",
            StorageClass: "fast",
        });
        expect(
            row("PersistentVolume", {
                spec: {
                    capacity: { storage: "1Gi" },
                    accessModes: ["ReadWriteMany", "ReadWriteOncePod"],
                    persistentVolumeReclaimPolicy: "Delete",
                    claimRef: { namespace: "ns", name: "c" },
                },
                status: { phase: "Bound" },
            })
        ).toEqual({
            Capacity: "1Gi",
            "Access modes": "RWX,RWOP",
            "Reclaim policy": "Delete",
            Status: "Bound",
            Claim: "ns/c",
            StorageClass: "<none>",
        });
    });

    it("covers nodes and namespaces", () => {
        expect(
            row("Node", {
                metadata: {
                    labels: {
                        "node-role.kubernetes.io/worker": "",
                        "node-role.kubernetes.io/control-plane": "",
                        other: "x",
                    },
                },
                spec: { unschedulable: true },
                status: {
                    conditions: [{ type: "Ready", status: "True" }],
                    nodeInfo: { kubeletVersion: "v1.30.0" },
                },
            })
        ).toEqual({
            Status: "Ready,SchedulingDisabled",
            Roles: "control-plane,worker",
            Version: "v1.30.0",
        });
        expect(row("Namespace", { status: { phase: "Active" } })).toEqual({
            Status: "Active",
        });
    });

    it("covers events for both API versions, against capture time", () => {
        expect(
            row("Event", {
                type: "Warning",
                reason: "BackOff",
                message: " back-off \n",
                involvedObject: { kind: "Pod", name: "web" },
                lastTimestamp: "2026-11-17T11:58:00Z",
            })
        ).toEqual({
            "Last seen": "2m",
            Type: "Warning",
            Reason: "BackOff",
            Object: "pod/web",
            Message: "back-off",
        });
        expect(
            row("Event", {
                series: { lastObservedTime: "2026-11-17T11:00:00Z" },
            })["Last seen"]
        ).toBe("60m");
    });

    it("sorts events newest first and leaves other kinds alone", () => {
        const rows = [
            { object: { lastTimestamp: "2026-11-17T10:00:00Z" } },
            { object: { eventTime: "2026-11-17T11:00:00Z" } },
        ];
        expect(sortObjects("Event", rows)[0]).toBe(rows[1]);
        expect(sortObjects("Node", rows)).toBe(rows);
    });
});

describe("problem", () => {
    it("flags workloads with fewer ready than desired", () => {
        const d = (replicas: number, readyReplicas?: number) => ({
            spec: { replicas },
            status: { readyReplicas },
        });
        expect(problem("Deployment", d(3, 3))).toBeNull();
        expect(problem("Deployment", d(3, 1))).toBe("warning");
        expect(problem("StatefulSet", d(2, 0))).toBe("error");
        expect(problem("Deployment", d(0))).toBeNull();
        // Old revision of a Deployment.
        expect(problem("ReplicaSet", d(0, 0))).toBeNull();
        expect(problem("ReplicaSet", d(2, 1))).toBe("warning");
        expect(
            problem("DaemonSet", {
                status: { desiredNumberScheduled: 3, numberReady: 2 },
            })
        ).toBe("warning");
        expect(problem("DaemonSet", {})).toBeNull();
    });

    it("flags nodes, jobs and claims", () => {
        const ready = (s: string) => ({
            status: { conditions: [{ type: "Ready", status: s }] },
        });
        expect(problem("Node", ready("True"))).toBeNull();
        expect(problem("Node", ready("False"))).toBe("error");
        expect(
            problem("Job", {
                status: { conditions: [{ type: "Failed", status: "True" }] },
            })
        ).toBe("error");
        expect(problem("Job", {})).toBeNull();
        expect(
            problem("PersistentVolumeClaim", { status: { phase: "Bound" } })
        ).toBeNull();
        expect(
            problem("PersistentVolumeClaim", { status: { phase: "Pending" } })
        ).toBe("warning");
        expect(problem("ConfigMap", {})).toBeNull();
    });
});
