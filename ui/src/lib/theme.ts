import { createContext, use } from "react";

export type Theme = "dark" | "light" | "system";

export type ThemeState = { theme: Theme; setTheme: (t: Theme) => void };

export const ThemeContext = createContext<ThemeState | null>(null);

export function useTheme(): ThemeState {
    const ctx = use(ThemeContext);
    if (!ctx) throw new Error("useTheme outside ThemeProvider");
    return ctx;
}
