export function Loading({ what }: { what: string }) {
    return <p className="text-muted-foreground py-8">Loading {what}…</p>;
}

export function LoadError({ message }: { message: string }) {
    return (
        <p className="text-destructive py-8" role="alert">
            Couldn't load this: {message}
        </p>
    );
}

/** Says why a section has nothing to show, in place of an empty panel. */
export function Empty({ children }: { children: React.ReactNode }) {
    return (
        <p className="text-muted-foreground rounded-lg border border-dashed px-4 py-6 text-sm">
            {children}
        </p>
    );
}
