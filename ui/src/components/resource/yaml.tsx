import { useMemo, useState } from "react";

import { Check, Copy } from "lucide-react";

import type { Manifest } from "@/lib/api";
import { toYaml } from "@/lib/yaml";

import { Button } from "@/components/ui/button";

/** The object as YAML, with a Copy button that copies the text shown. */
export function YamlView({
    object,
    manifest,
}: {
    object: unknown;
    manifest: Manifest;
}) {
    const envRedacted = manifest.redactions?.includes("env-values") ?? false;
    const yaml = useMemo(
        () => toYaml(object, envRedacted),
        [object, envRedacted]
    );
    const [copied, setCopied] = useState(false);

    async function copy() {
        try {
            await navigator.clipboard.writeText(yaml);
            setCopied(true);
            setTimeout(() => setCopied(false), 2000);
        } catch {
            setCopied(false);
        }
    }

    return (
        <section aria-labelledby="yaml" className="space-y-3">
            <div className="flex items-center justify-between gap-2">
                <h3 id="yaml" className="text-lg font-semibold">
                    YAML
                </h3>
                <Button variant="outline" size="sm" onClick={copy}>
                    {copied ? <Check /> : <Copy />}
                    {copied ? "Copied" : "Copy"}
                </Button>
            </div>
            <pre className="bg-muted max-h-[32rem] overflow-auto rounded-lg p-4 font-mono text-xs leading-relaxed">
                {yaml}
            </pre>
        </section>
    );
}
