import { StrictMode } from "react";

import "@fontsource-variable/geist";
import "@fontsource-variable/geist-mono";
import { createRoot } from "react-dom/client";
import { RouterProvider, createBrowserRouter } from "react-router";

import { Layout } from "@/components/common/layout";
import { ThemeProvider } from "@/components/common/theme-provider";

import "./index.css";

import { HomePage } from "@/pages/home";
import { KindPage } from "@/pages/kind";
import { NotFoundPage } from "@/pages/not-found";
import { PodPage } from "@/pages/pod";
import { PodsPage } from "@/pages/pods";
import { ResourcePage } from "@/pages/resource";
import { SnapshotLayout } from "@/pages/snapshot";

const router = createBrowserRouter([
    {
        element: <Layout />,
        children: [
            { path: "/", element: <HomePage /> },
            {
                path: "/s/:id",
                element: <SnapshotLayout />,
                children: [
                    { index: true, element: <PodsPage /> },
                    { path: "pods/:ns/:name", element: <PodPage /> },
                    { path: ":resource", element: <KindPage /> },
                    {
                        path: ":resource/:ns/:name",
                        element: <ResourcePage />,
                    },
                    { path: "*", element: <NotFoundPage /> },
                ],
            },
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
