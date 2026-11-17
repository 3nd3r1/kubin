import { useEffect, useState } from "react";

import { type Theme, ThemeContext } from "@/lib/theme";

const storageKey = "kubin-theme";

function stored(): Theme {
    try {
        const t = localStorage.getItem(storageKey);
        if (t === "dark" || t === "light" || t === "system") return t;
    } catch {
        // Storage can be blocked; fall back to the default.
    }
    return "dark";
}

// index.html applies the stored theme before first paint; this keeps it in
// sync afterwards.
export function ThemeProvider({ children }: { children: React.ReactNode }) {
    const [theme, setTheme] = useState<Theme>(stored);

    useEffect(() => {
        const media = matchMedia("(prefers-color-scheme: dark)");
        const apply = () =>
            document.documentElement.classList.toggle(
                "dark",
                theme === "dark" || (theme === "system" && media.matches)
            );
        apply();
        try {
            localStorage.setItem(storageKey, theme);
        } catch {
            // Not remembered, still applied.
        }
        if (theme !== "system") return;
        media.addEventListener("change", apply);
        return () => media.removeEventListener("change", apply);
    }, [theme]);

    return <ThemeContext value={{ theme, setTheme }}>{children}</ThemeContext>;
}
