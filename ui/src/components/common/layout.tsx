import { Link, Outlet } from "react-router";

import { ThemeToggle } from "./theme-toggle";

export function Layout() {
    return (
        <div className="flex min-h-screen flex-col">
            <header className="border-b">
                <div className="mx-auto flex h-14 max-w-6xl items-center justify-between px-4">
                    <Link
                        to="/"
                        className="flex items-center gap-2 font-semibold"
                    >
                        <img src="/logo.png" alt="" className="size-7" />
                        kubin
                    </Link>
                    <ThemeToggle />
                </div>
            </header>
            <main className="mx-auto w-full max-w-6xl flex-1 px-4 py-8">
                <Outlet />
            </main>
        </div>
    );
}
