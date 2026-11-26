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
            <section className="text-muted-foreground space-y-2 border-t pt-6 text-sm">
                <h2 className="text-foreground font-medium">
                    Privacy &amp; abuse
                </h2>
                <p>
                    This instance is run by 3nd3r1. Anyone with a snapshot link
                    can browse it until it expires. The CLI strips Kubernetes
                    Secrets and env var values by default, but logs, container
                    args, and ConfigMaps are uploaded as-is.
                </p>
                <p>
                    Delete a snapshot early with{" "}
                    <code className="text-foreground font-mono">
                        kubin delete &lt;url&gt;
                    </code>
                    . For abuse or takedown requests, email{" "}
                    <a
                        href="mailto:abuse@ender.fi"
                        className="text-foreground underline"
                    >
                        abuse@ender.fi
                    </a>
                    .
                </p>
            </section>
        </div>
    );
}
