import { Moon, Sun } from "lucide-react";

import { useTheme } from "@/lib/theme";

import { Button } from "@/components/ui/button";

export function ThemeToggle() {
    const { theme, setTheme } = useTheme();
    const dark =
        theme === "dark" ||
        (theme === "system" &&
            matchMedia("(prefers-color-scheme: dark)").matches);
    return (
        <Button
            variant="ghost"
            size="icon"
            aria-label={dark ? "Use light theme" : "Use dark theme"}
            title={`Theme: ${theme}`}
            onClick={() => setTheme(dark ? "light" : "dark")}
        >
            {dark ? <Sun /> : <Moon />}
        </Button>
    );
}
