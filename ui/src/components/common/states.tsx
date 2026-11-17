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
