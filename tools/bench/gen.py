"""Scale up a real kubin snapshot for load tests (docs/capacity.md).

Clones every namespace and its resources into <copies> namespaces and gives
every log file about <log_kb> KiB of generated lines, which compress worse
than real logs.

    tar xzf real.tar.gz -C real/
    python3 gen.py real/ big.tar.gz <copies> <log_kb>
"""
import json, os, random, sys, tarfile, io

src, out, copies, log_kb = sys.argv[1], sys.argv[2], int(sys.argv[3]), int(sys.argv[4])
random.seed(1)
m = json.load(open(os.path.join(src, "manifest.json")))
files = {}  # path -> bytes
counts = {}

def add_res(g, v, k, ns, name, body):
    files[f"resources/{g}/{v}/{k}/{ns}/{name}.json"] = body
    counts[(g, v, k)] = counts.get((g, v, k), 0) + 1

for root, _, fs in os.walk(os.path.join(src, "resources")):
    for f in fs:
        p = os.path.relpath(os.path.join(root, f), src).split("/")
        _, g, v, k, ns, name = p[0], p[1], p[2], p[3], p[4], p[5][:-5]
        raw = open(os.path.join(root, f)).read()
        if ns == "_cluster":
            add_res(g, v, k, ns, name, raw.encode())
            if k == "Namespace":
                for i in range(1, copies):
                    obj = json.loads(raw); obj["metadata"]["name"] = f"{name}-{i}"
                    add_res(g, v, k, ns, f"{name}-{i}", json.dumps(obj, indent=2).encode())
            continue
        for i in range(copies):
            nns = ns if i == 0 else f"{ns}-{i}"
            obj = json.loads(raw); obj["metadata"]["namespace"] = nns
            add_res(g, v, k, nns, name, json.dumps(obj, indent=2).encode())

words = "GET POST PUT /api/v1/orders /healthz /cart user_id order_id status=200 status=500 latency_ms connection refused retrying timeout ok processed queue db pool".split()
def logtext(n):
    lines, size = [], 0
    while size < n:
        l = f"2026-11-24T10:{random.randint(0,59):02d}:{random.randint(0,59):02d}.{random.randint(0,999999):06d}Z {random.choice(['INFO','WARN','ERROR','DEBUG'])} " + " ".join(random.choice(words) + ("=" + str(random.randint(0, 99999)) if random.random() < .4 else "") for _ in range(random.randint(4, 14)))
        lines.append(l); size += len(l) + 1
    return ("\n".join(lines) + "\n").encode()

logs = []
for lf in m["logs"]["files"]:
    for i in range(copies):
        ns = lf["namespace"] if i == 0 else f"{lf['namespace']}-{i}"
        e = dict(lf, namespace=ns)
        logs.append(e)
        suffix = ".previous.log" if lf["previous"] else ".log"
        files[f"logs/{ns}/{lf['pod']}/{lf['container']}{suffix}"] = logtext(log_kb * 1024)

m["resources"] = [{"group": g, "version": v, "kind": k, "count": c} for (g, v, k), c in sorted(counts.items())]
m["logs"]["files"] = logs
files["manifest.json"] = json.dumps(m, indent=2).encode()
with tarfile.open(out, "w:gz") as t:
    for p in ["manifest.json"] + sorted(k for k in files if k != "manifest.json"):
        ti = tarfile.TarInfo(p); ti.size = len(files[p]); ti.mode = 0o644
        t.addfile(ti, io.BytesIO(files[p]))
print(f"{out}: {sum(counts.values())} resources, {len(logs)} logs, unpacked {sum(map(len, files.values()))/1e6:.1f} MB, gz {os.path.getsize(out)/1e6:.1f} MB")
