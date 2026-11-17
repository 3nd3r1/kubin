import { StrictMode } from "react";

import "@fontsource-variable/geist";
import "@fontsource-variable/geist-mono";
import { createRoot } from "react-dom/client";
import { RouterProvider, createBrowserRouter } from "react-router";

import { Layout } from "@/components/common/layout";
import { ThemeProvider } from "@/components/common/theme-provider";

import "./index.css";

import { HomePage } from "@/pages/home";
import { NotFoundPage } from "@/pages/not-found";
import { SnapshotPage } from "@/pages/snapshot";

const router = createBrowserRouter([
    {
        element: <Layout />,
        children: [
            { path: "/", element: <HomePage /> },
            { path: "/s/:id", element: <SnapshotPage /> },
            { path: "*", element: <NotFoundPage /> },
        ],
    },
]);

createRoot(document.getElementById("root")!).render(
    <StrictMode>
        <ThemeProvider>
            <RouterProvider router={router} />
        </ThemeProvider>
    </StrictMode>
);
