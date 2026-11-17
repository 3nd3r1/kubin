import { type DependencyList, useEffect, useState } from "react";

import { NotFoundError } from "./api";

export type FetchState<T> =
    | { status: "loading" }
    | { status: "ok"; data: T }
    | { status: "missing" }
    | { status: "error"; message: string };

const loading = { status: "loading" } as const;

/**
 * Runs load whenever deps change and aborts the previous call. A result
 * belongs to the deps it was loaded for, so changing deps shows loading
 * again without resetting state inside the effect.
 */
export function useFetch<T>(
    load: (signal: AbortSignal) => Promise<T>,
    deps: DependencyList
): FetchState<T> {
    const key = JSON.stringify(deps);
    const [result, setResult] = useState<{
        key: string;
        state: FetchState<T>;
    } | null>(null);

    useEffect(() => {
        const ctrl = new AbortController();
        load(ctrl.signal)
            .then((data) => setResult({ key, state: { status: "ok", data } }))
            .catch((err: unknown) => {
                if (ctrl.signal.aborted) return;
                setResult({
                    key,
                    state:
                        err instanceof NotFoundError
                            ? { status: "missing" }
                            : {
                                  status: "error",
                                  message:
                                      err instanceof Error
                                          ? err.message
                                          : String(err),
                              },
                });
            });
        return () => ctrl.abort();
        // load is a new closure every render; key carries its inputs.
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [key]);

    return result?.key === key ? result.state : loading;
}
