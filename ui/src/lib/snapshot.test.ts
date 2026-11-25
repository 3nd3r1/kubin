import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { formatExpiry } from "./snapshot";

const NOW = new Date("2026-11-25T12:00:00Z");

describe("formatExpiry", () => {
    beforeEach(() => {
        vi.useFakeTimers();
        vi.setSystemTime(NOW);
    });

    afterEach(() => {
        vi.useRealTimers();
    });

    it("returns null when expires_at is missing or invalid", () => {
        expect(formatExpiry(undefined)).toBeNull();
        expect(formatExpiry("")).toBeNull();
        expect(formatExpiry("not-a-date")).toBeNull();
    });

    it("returns expired when expires_at is now or in the past", () => {
        expect(formatExpiry("2026-11-25T12:00:00Z")).toBe("expired");
        expect(formatExpiry("2026-11-24T12:00:00Z")).toBe("expired");
    });

    it("formats relative durations for future expiries", () => {
        expect(formatExpiry("2026-11-25T12:00:20Z")).toBe(
            "expires in less than a minute"
        );
        expect(formatExpiry("2026-11-25T12:01:00Z")).toBe(
            "expires in 1 minute"
        );
        expect(formatExpiry("2026-11-25T12:15:00Z")).toBe(
            "expires in 15 minutes"
        );
        expect(formatExpiry("2026-11-25T13:00:00Z")).toBe("expires in 1 hour");
        expect(formatExpiry("2026-11-25T17:00:00Z")).toBe("expires in 5 hours");
        expect(formatExpiry("2026-11-26T12:00:00Z")).toBe("expires in 1 day");
        expect(formatExpiry("2026-11-28T12:00:00Z")).toBe("expires in 3 days");
    });
});
