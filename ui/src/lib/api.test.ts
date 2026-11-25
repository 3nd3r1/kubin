import { afterEach, describe, expect, it, vi } from "vitest";

import {
    ExpiredError,
    NotFoundError,
    getLog,
    getResource,
    getSnapshot,
    listResources,
} from "./api";

afterEach(() => {
    vi.unstubAllGlobals();
});

function mockFetch(status: number, body: unknown = { error: "boom" }) {
    const fn = vi.fn().mockResolvedValue(
        new Response(JSON.stringify(body), {
            status,
            headers: { "Content-Type": "application/json" },
        })
    );
    vi.stubGlobal("fetch", fn);
    return fn;
}

describe("api error mapping", () => {
    it("throws ExpiredError on 410 for getSnapshot", async () => {
        mockFetch(410, { error: "snapshot expired" });
        await expect(getSnapshot("snap-1")).rejects.toBeInstanceOf(
            ExpiredError
        );
    });

    it("throws ExpiredError on 410 for listResources, getResource, and getLog", async () => {
        mockFetch(410, { error: "snapshot expired" });
        await expect(
            listResources("snap-1", { kind: "Pod" })
        ).rejects.toBeInstanceOf(ExpiredError);

        mockFetch(410, { error: "snapshot expired" });
        await expect(
            getResource("snap-1", "Pod", "default", "web")
        ).rejects.toBeInstanceOf(ExpiredError);

        mockFetch(410, { error: "snapshot expired" });
        await expect(
            getLog("snap-1", {
                namespace: "default",
                pod: "web",
                container: "app",
                previous: false,
            })
        ).rejects.toBeInstanceOf(ExpiredError);
    });

    it("throws NotFoundError on 404", async () => {
        mockFetch(404, { error: "not found" });
        await expect(getSnapshot("snap-1")).rejects.toBeInstanceOf(
            NotFoundError
        );
    });

    it("throws a generic Error on other non-ok statuses", async () => {
        mockFetch(500, { error: "internal error" });
        await expect(getSnapshot("snap-1")).rejects.toThrow("internal error");
    });
});
