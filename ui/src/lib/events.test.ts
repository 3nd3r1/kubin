import { describe, expect, it } from "vitest";

import { podEvents } from "./events";
import type { Event, Pod } from "./k8s";

const pod: Pod = { metadata: { name: "api-1", namespace: "shop", uid: "u1" } };

function ev(over: Partial<Event>): Event {
    return {
        involvedObject: { kind: "Pod", name: "api-1", uid: "u1" },
        ...over,
    };
}

describe("podEvents", () => {
    it("keeps only this pod's events, newest first", () => {
        const got = podEvents(pod, [
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
                involvedObject: { kind: "Pod", name: "api-1", uid: "old" },
            }),
        ]);
        expect(got.map((e) => [e.reason, e.type, e.count])).toEqual([
            ["BackOff", "Warning", 12],
            ["Pulled", "Normal", 1],
        ]);
    });

    it("reads times and counts from every event version", () => {
        const got = podEvents(pod, [
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
        const noUid: Pod = { metadata: { name: "api-1", namespace: "shop" } };
        expect(podEvents(noUid, [ev({ reason: "x" })])).toHaveLength(1);
        expect(
            podEvents(pod, [
                ev({ involvedObject: { kind: "Pod", name: "api-1" } }),
            ])
        ).toHaveLength(1);
    });
});
