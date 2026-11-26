import { Link, Outlet, useMatch } from "react-router";

import { cn } from "@/lib/utils";

import { ThemeToggle } from "./theme-toggle";

export function Layout() {
    // Snapshot pages need the width for the sidebar and a table.
    const wide = useMatch("/s/*") !== null;
    return (
        <div className="flex min-h-screen flex-col">
            <header className="border-b">
                <div className="mx-auto flex h-14 max-w-6xl items-center justify-between px-4">
                    <Link
                        to="/"
                        className="flex items-center gap-2 font-semibold"
                    >
                        <img src="/icon.png" alt="" className="size-7" />
                        kubin
                    </Link>
                    <ThemeToggle />
                </div>
            </header>
            <main
                className={cn(
                    "mx-auto w-full flex-1 px-4 py-8",
                    !wide && "max-w-6xl"
                )}
            >
                <Outlet />
            </main>
            <footer className="text-muted-foreground border-t text-xs">
                <div className="mx-auto flex max-w-6xl flex-wrap items-center justify-between gap-2 px-4 py-3">
                    <span>
                        Run by 3nd3r1. Anyone with a link can view a snapshot
                        until it expires; delete early with{" "}
                        <code className="font-mono">
                            kubin delete &lt;url&gt;
                        </code>
                        .
                    </span>
                    <span>
                        Abuse reports:{" "}
                        <a
                            href="mailto:abuse@ender.fi"
                            className="hover:text-foreground underline"
                        >
                            abuse@ender.fi
                        </a>
                    </span>
                </div>
            </footer>
        </div>
    );
}
