import { Document, isMap, isPair, isScalar, isSeq, visit } from "yaml";

export const redactedComment = " value redacted by kubin";

const containerLists = new Set([
    "containers",
    "initContainers",
    "ephemeralContainers",
]);

/**
 * The object as YAML. When the snapshot ran the env-values rule, env entries
 * with neither `value` nor `valueFrom` get a comment saying so. The comment
 * is the only addition: no value is made up.
 */
export function toYaml(object: unknown, envRedacted: boolean): string {
    const doc = new Document(object);
    if (envRedacted) {
        visit(doc, {
            Pair(_, pair, path) {
                if (!isScalar(pair.key) || pair.key.value !== "env") return;
                if (!isSeq(pair.value)) return;
                // Only the env of a container: the seq's grandparent pair is
                // containers, initContainers or ephemeralContainers.
                const owner = path[path.length - 3];
                if (
                    !isPair(owner) ||
                    !isScalar(owner.key) ||
                    !containerLists.has(String(owner.key.value))
                )
                    return;
                for (const item of pair.value.items) {
                    if (!isMap(item)) continue;
                    if (item.has("value") || item.has("valueFrom")) continue;
                    const name = item.items.find(
                        (p) => isScalar(p.key) && p.key.value === "name"
                    );
                    if (name && isScalar(name.value))
                        name.value.comment = redactedComment;
                }
            },
        });
    }
    return doc.toString({ lineWidth: 0 });
}
