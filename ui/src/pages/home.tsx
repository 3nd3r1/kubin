export function HomePage() {
    return (
        <div className="mx-auto max-w-2xl space-y-6 py-12">
            <h1 className="text-3xl font-bold">
                Share a Kubernetes cluster as a link
            </h1>
            <p className="text-muted-foreground">
                Snapshot the cluster in your current kube context and send the
                link. Anyone who opens it can browse the resources as they were,
                without access to the cluster.
            </p>
            <pre className="bg-muted overflow-x-auto rounded-lg p-4 font-mono text-sm">
                <code>{`kubin create --server ${window.location.origin}`}</code>
            </pre>
        </div>
    );
}
