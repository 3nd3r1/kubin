import { describe, expect, it } from "vitest";

import { recentWarnings, resourceEvents } from "./events";
import type { Event, Pod } from "./k8s";

const pod: Pod = {
    kind: "Pod",
    metadata: { name: "api-1", namespace: "shop", uid: "u1" },
};

function ev(over: Partial<Event>): Event {
    return {
        involvedObject: {
            kind: "Pod",
            namespace: "shop",
            name: "api-1",
            uid: "u1",
        },
        ...over,
    };
}

describe("resourceEvents", () => {
    it("keeps only this pod's events, newest first", () => {
        const got = resourceEvents(pod, [
            ev({ reason: "Pulled", lastTimestamp: "2026-11-17T10:00:00Z" }),
            ev({
                reason: "BackOff",
                type: "Warning",
                count: 12,
                lastTimestamp: "2026-11-17T10:05:00Z",
            }),
            ev({
                reason: "Other pod",
                involvedObject: { kind: "Pod", name: "api-2" },
            }),
            ev({
                reason: "Same name, a Deployment",
                involvedObject: { kind: "Deployment", name: "api-1" },
            }),
            ev({
                reason: "Earlier pod with this name",
                involvedObject: {
                    kind: "Pod",
                    namespace: "shop",
                    name: "api-1",
                    uid: "old",
                },
            }),
        ]);
        expect(got.map((e) => [e.reason, e.type, e.count])).toEqual([
            ["BackOff", "Warning", 12],
            ["Pulled", "Normal", 1],
        ]);
    });

    it("reads times and counts from every event version", () => {
        const got = resourceEvents(pod, [
            ev({ reason: "a", firstTimestamp: "2026-11-17T10:01:00Z" }),
            ev({
                reason: "b",
                eventTime: "2026-11-17T10:02:00.500000Z",
                series: {
                    count: 7,
                    lastObservedTime: "2026-11-17T10:04:00.250000Z",
                },
            }),
            ev({ reason: "c", eventTime: "2026-11-17T10:03:00.000001Z" }),
            ev({ reason: "no time" }),
        ]);
        expect(got.map((e) => e.reason)).toEqual(["b", "c", "a", "no time"]);
        expect(got[0].count).toBe(7);
    });

    it("matches by name when either side has no uid", () => {
        const noUid: Pod = {
            kind: "Pod",
            metadata: { name: "api-1", namespace: "shop" },
        };
        expect(resourceEvents(noUid, [ev({ reason: "x" })])).toHaveLength(1);
        expect(
            resourceEvents(pod, [
                ev({
                    involvedObject: {
                        kind: "Pod",
                        namespace: "shop",
                        name: "api-1",
                    },
                }),
            ])
        ).toHaveLength(1);
    });

    it("keeps events of the same name in another namespace apart", () => {
        expect(
            resourceEvents(pod, [
                ev({
                    involvedObject: {
                        kind: "Pod",
                        namespace: "other",
                        name: "api-1",
                        uid: "u1",
                    },
                }),
            ])
        ).toHaveLength(0);
    });

    it("matches a node, whose events have no namespace", () => {
        const node = {
            kind: "Node",
            metadata: { name: "n1", uid: "nu" },
        };
        const got = resourceEvents(node, [
            ev({
                reason: "NodeNotReady",
                involvedObject: { kind: "Node", name: "n1", uid: "nu" },
                lastTimestamp: "2026-11-17T10:00:00Z",
            }),
            ev({
                reason: "Rebooted",
                involvedObject: { kind: "Node", name: "n1" },
                lastTimestamp: "2026-11-17T11:00:00Z",
            }),
            ev({
                reason: "Other node",
                involvedObject: { kind: "Node", name: "n2" },
            }),
            ev({
                reason: "A pod called n1",
                involvedObject: { kind: "Pod", name: "n1" },
            }),
        ]);
        expect(got.map((e) => e.reason)).toEqual(["Rebooted", "NodeNotReady"]);
    });
});

describe("recentWarnings", () => {
    const captured = "2026-11-17T12:00:00Z";
    const item = (reason: string, over: Partial<Event>) => ({
        object: ev({ type: "Warning", reason, ...over }),
    });

    it("keeps warnings from the hour before capture, newest first", () => {
        const got = recentWarnings(
            [
                item("old", { lastTimestamp: "2026-11-17T10:59:59Z" }),
                item("edge", { lastTimestamp: "2026-11-17T11:00:00Z" }),
                item("newest", { eventTime: "2026-11-17T11:50:00Z" }),
                item("normal", {
                    type: "Normal",
                    lastTimestamp: "2026-11-17T11:30:00Z",
                }),
                item("after", { lastTimestamp: "2026-11-17T12:00:01Z" }),
                item("mid", { lastTimestamp: "2026-11-17T11:20:00Z" }),
            ],
            captured
        );
        expect(got.map((e) => e.object.reason)).toEqual([
            "newest",
            "mid",
            "edge",
        ]);
    });

    it("leaves out events without a time", () => {
        expect(recentWarnings([item("none", {})], captured)).toEqual([]);
    });
});
