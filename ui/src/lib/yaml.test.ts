import { describe, expect, it } from "vitest";

import { toYaml } from "./yaml";

const pod = {
    kind: "Pod",
    spec: {
        initContainers: [{ name: "init", env: [{ name: "A" }] }],
        containers: [
            {
                name: "app",
                env: [
                    { name: "PLAIN", value: "" },
                    { name: "GONE" },
                    {
                        name: "REF",
                        valueFrom: { fieldRef: { fieldPath: "x" } },
                    },
                ],
            },
        ],
    },
    metadata: { labels: { env: "prod" }, annotations: { name: "x" } },
};

describe("toYaml", () => {
    it("marks env entries that lost their value", () => {
        const out = toYaml(pod, true);
        expect(out).toContain("name: GONE # value redacted by kubin");
        expect(out).toContain("name: A # value redacted by kubin");
        expect(out.match(/redacted/g)).toHaveLength(2);
    });

    it("leaves entries with a value or valueFrom alone", () => {
        const out = toYaml(pod, true);
        expect(out).toMatch(/name: PLAIN\n\s+value: ""/);
        expect(out).not.toMatch(/name: REF #/);
    });

    it("adds no value and parses back to the same object", async () => {
        const { parse } = await import("yaml");
        expect(parse(toYaml(pod, true))).toEqual(pod);
    });

    it("marks nothing when the rule did not run", () => {
        expect(toYaml(pod, false)).not.toContain("redacted");
    });
});
