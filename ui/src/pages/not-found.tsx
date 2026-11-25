import { Link } from "react-router";

import { Button } from "@/components/ui/button";

export function NotFoundPage({
    title = "Page not found",
    detail = "There is nothing at this address.",
}: {
    title?: string;
    detail?: string;
}) {
    return (
        <div className="mx-auto max-w-md space-y-4 py-24 text-center">
            <h1 className="text-2xl font-semibold">{title}</h1>
            <p className="text-muted-foreground">{detail}</p>
            <Button asChild variant="outline">
                <Link to="/">Go home</Link>
            </Button>
        </div>
    );
}

export function ExpiredPage() {
    return (
        <NotFoundPage
            title="Snapshot expired"
            detail="This snapshot has expired and is no longer available."
        />
    );
}
