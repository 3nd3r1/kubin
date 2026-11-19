import { describe, expect, it } from "vitest";

import type { KindCount } from "./api";
import { groupKinds, kinds, resolveSegment } from "./kinds";
import { kindPath, nsParam, podPath, resourcePath } from "./snapshot";

const c = (group: string, kind: string, count = 1): KindCount => ({
    group,
    version: "v1",
    kind,
    count,
});

describe("groupKinds", () => {
    it("groups by section in table order and hides absent kinds", () => {
        const got = groupKinds([
            c("core", "Node", 3),
            c("apps", "Deployment", 2),
            c("core", "Pod", 9),
        ]);
        expect(got.map((s) => s.title)).toEqual(["Workloads", "Cluster"]);
        expect(got[0].items.map((i) => [i.segment, i.count])).toEqual([
            ["pods", 9],
            ["deployments", 2],
        ]);
    });

    it("puts unknown kinds under Other", () => {
        const got = groupKinds([c("core", "Pod"), c("foo.io", "Widget", 4)]);
        const other = got.at(-1)!;
        expect(other.title).toBe("Other");
        expect(other.items[0]).toMatchObject({
            kind: "Widget",
            segment: "widgets.foo.io",
            count: 4,
        });
        expect(other.items[0].info).toBeUndefined();
    });

    it("matches the core group by 'core', not an empty string", () => {
        expect(groupKinds([c("", "Pod")])[0].title).toBe("Other");
    });
});

describe("resolveSegment", () => {
    it("has a unique segment per kind", () => {
        expect(new Set(kinds.map((i) => i.segment)).size).toBe(kinds.length);
    });
    it("resolves known and unknown kinds", () => {
        expect(resolveSegment("pods", [])?.kind).toBe("Pod");
        expect(
            resolveSegment("widgets.foo.io", [c("foo.io", "Widget")])
        ).toEqual({ kind: "Widget", group: "foo.io" });
        expect(resolveSegment("nope", [c("foo.io", "Widget")])).toBeUndefined();
    });
});

describe("paths", () => {
    it("keeps the namespace filter on kind links", () => {
        expect(kindPath("a", "pods")).toBe("/s/a/pods");
        expect(kindPath("a", "pods", "kube system")).toBe(
            "/s/a/pods?ns=kube%20system"
        );
    });
    it("reads ns", () => {
        expect(nsParam("?ns=web")).toBe("web");
        expect(nsParam("")).toBe("");
    });
    it("builds detail paths", () => {
        expect(resourcePath("a", "nodes", { namespace: "", name: "n1" })).toBe(
            "/s/a/nodes/_cluster/n1"
        );
        expect(podPath("a", { namespace: "x", name: "y" })).toBe(
            "/s/a/pods/x/y"
        );
    });
});
