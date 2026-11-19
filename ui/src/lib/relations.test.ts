import { describe, expect, it } from "vitest";

import type { ResourceItem } from "./api";
import {
    findOwner,
    ownedBy,
    ownedKind,
    ownerKind,
    podsFor,
    podsOnNode,
    selectorMatches,
    servicesFor,
} from "./relations";

type O = Parameters<typeof ownedBy>[1][number]["object"];

const item = (
    kind: string,
    namespace: string,
    name: string,
    object: O
): ResourceItem<O> => ({
    group: "core",
    version: "v1",
    kind,
    namespace,
    name,
    object: {
        kind,
        ...object,
        metadata: { name, namespace, ...object.metadata },
    },
});

const rs = (name: string, owner: string) =>
    item("ReplicaSet", "app", name, {
        metadata: {
            ownerReferences: [{ kind: "Deployment", name: "web", uid: owner }],
        },
    });

describe("ownedBy", () => {
    it("matches by UID, not name", () => {
        const items = [rs("web-1", "old"), rs("web-2", "new")];
        expect(ownedBy("new", items).map((i) => i.name)).toEqual(["web-2"]);
    });
    it("matches nothing without a UID", () => {
        expect(ownedBy(undefined, [rs("web-1", "")])).toEqual([]);
        expect(ownedBy("", [rs("web-1", "")])).toEqual([]);
    });
    it("knows the owned kinds", () => {
        expect(ownedKind("Deployment")).toBe("ReplicaSet");
        expect(ownedKind("CronJob")).toBe("Job");
        expect(ownedKind("Service")).toBeUndefined();
    });
});

describe("selectorMatches", () => {
    it("needs every entry", () => {
        expect(selectorMatches({ a: "1" }, { a: "1", b: "2" })).toBe(true);
        expect(selectorMatches({ a: "1", c: "3" }, { a: "1" })).toBe(false);
        expect(selectorMatches({ a: "1" }, undefined)).toBe(false);
    });
    it("matches nothing for no or an empty selector", () => {
        expect(selectorMatches(undefined, { a: "1" })).toBe(false);
        expect(selectorMatches({}, { a: "1" })).toBe(false);
    });
});

describe("services and pods", () => {
    const svc = item("Service", "app", "web", {
        spec: { selector: { app: "web" } },
    });
    const none = item("Service", "app", "ext", { spec: {} });
    const pod = (
        ns: string,
        name: string,
        labels: Record<string, string>,
        node = ""
    ) =>
        item("Pod", ns, name, {
            metadata: { labels },
            spec: { nodeName: node },
        });

    it("finds services for a pod, skipping selectorless ones", () => {
        const p = pod("app", "p", { app: "web" });
        expect(servicesFor(p.object, [svc, none])).toEqual([svc]);
    });
    it("finds pods for a service in its namespace only", () => {
        const pods = [
            pod("app", "a", { app: "web" }),
            pod("other", "b", { app: "web" }),
            pod("app", "c", { app: "db" }),
        ];
        expect(podsFor(svc.object, pods).map((p) => p.name)).toEqual(["a"]);
        expect(podsFor(none.object, pods)).toEqual([]);
    });
    it("finds pods on a node across namespaces", () => {
        const pods = [
            pod("a", "x", {}, "n1"),
            pod("b", "y", {}, "n1"),
            pod("a", "z", {}, "n2"),
        ];
        expect(podsOnNode("n1", pods).map((p) => p.name)).toEqual(["x", "y"]);
        expect(podsOnNode("", pods)).toEqual([]);
    });
});

describe("owners", () => {
    const ref = { kind: "Deployment", name: "web", uid: "u1" };
    it("finds the owner by UID", () => {
        const d = item("Deployment", "app", "web", { metadata: { uid: "u1" } });
        expect(findOwner(ref, [d])).toBe(d);
    });
    it("does not link a recreated owner with the same name", () => {
        const d = item("Deployment", "app", "web", { metadata: { uid: "u2" } });
        expect(findOwner(ref, [d])).toBeUndefined();
    });
    it("finds nothing when the owner is not in the snapshot", () => {
        expect(findOwner(ref, [])).toBeUndefined();
    });
    it("reads the group from apiVersion", () => {
        expect(ownerKind({ ...ref, apiVersion: "apps/v1" })).toMatchObject({
            kind: "Deployment",
            group: "apps",
        });
        expect(ownerKind({ kind: "Widget", apiVersion: "x.io/v1" })).toEqual({
            kind: "Widget",
            group: "x.io",
        });
    });
});
