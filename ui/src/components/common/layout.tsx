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
        </div>
    );
}
