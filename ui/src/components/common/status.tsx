import { CircleAlert, CircleCheck, TriangleAlert } from "lucide-react";

import type { Problem } from "@/lib/pods";
import { cn } from "@/lib/utils";

const styles = {
    error: "text-red-600 dark:text-red-400",
    warning: "text-amber-600 dark:text-amber-400",
    ok: "text-emerald-600 dark:text-emerald-400",
};

/** A pod status with an icon and colour for its problem level. */
export function PodStatus({
    status,
    problem,
    className,
}: {
    status: string;
    problem: Problem;
    className?: string;
}) {
    const level = problem ?? "ok";
    const Icon =
        level === "error"
            ? CircleAlert
            : level === "warning"
              ? TriangleAlert
              : CircleCheck;
    return (
        <span
            className={cn(
                "inline-flex items-center gap-1.5 font-medium",
                styles[level],
                className
            )}
        >
            <Icon className="size-4 shrink-0" aria-hidden />
            {status}
        </span>
    );
}
