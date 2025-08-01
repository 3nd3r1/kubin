import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";

import ThemeProvider from "@/components/common/layouts/theme-provider";

import "./globals.css";

const geist = Geist({
    variable: "--font-geist-sans",
    subsets: ["latin"],
});

const geistMono = Geist_Mono({
    variable: "--font-geist-mono",
    subsets: ["latin"],
});

export const metadata: Metadata = {
    title: "Kubin UI - Kubernetes Snapshot Viewer",
    description: "Lens-like interface for viewing Kubernetes cluster snapshots",
};

export default function RootLayout({
    children,
}: Readonly<{
    children: React.ReactNode;
}>) {
    return (
        <html lang="en" suppressHydrationWarning>
            <body
                className={`${geist.variable} ${geistMono.variable} antialiased`}
            >
                <ThemeProvider
                    attribute="class"
                    defaultTheme="dark"
                    enableSystem
                    disableTransitionOnChange
                >
                    {children}
                </ThemeProvider>
            </body>
        </html>
    );
}
