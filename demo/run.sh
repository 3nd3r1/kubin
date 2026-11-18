#!/usr/bin/env bash
set -Eeuo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cluster="kubin-demo"
container="kubin-demo"
port="${KUBIN_PORT:-8080}"
server="http://localhost:$port"
work_dir="$(mktemp -d "${TMPDIR:-/tmp}/kubin-demo.XXXXXX")"
cluster_created=0
container_started=0
finished=0

cleanup() {
  local status=$?
  if (( finished == 0 )); then
    if (( container_started == 1 )); then
      docker rm -f "$container" >/dev/null 2>&1 || true
    fi
    if (( cluster_created == 1 )); then
      kind delete cluster --name "$cluster" >/dev/null 2>&1 || true
    fi
  fi
  rm -rf "$work_dir"
  exit "$status"
}
trap cleanup EXIT

for tool in docker kind kubectl go curl; do
  if ! command -v "$tool" >/dev/null 2>&1; then
    echo "Missing required command: $tool" >&2
    exit 1
  fi
done

if kind get clusters | grep -Fxq "$cluster"; then
  echo "A kind cluster named '$cluster' already exists. Delete it or choose another name." >&2
  exit 1
fi
if docker container inspect "$container" >/dev/null 2>&1; then
  echo "A Docker container named '$container' already exists. Remove it or choose another name." >&2
  exit 1
fi

if curl --silent --max-time 2 "$server" >/dev/null 2>&1; then
  echo "Port $port is in use. Set KUBIN_PORT to another port." >&2
  exit 1
fi

echo "Building the Kubin server image..."
docker build -t kubin:demo "$repo_root"

echo "Starting the kind cluster..."
# --wait lets the node lose its not-ready taint first, so the pod does not
# log a FailedScheduling warning.
kind create cluster --name "$cluster" --wait 2m
cluster_created=1
kind export kubeconfig --name "$cluster" --kubeconfig "$work_dir/kubeconfig"
export KUBECONFIG="$work_dir/kubeconfig"

echo "Starting Kubin on $server..."
# docker run can create the container and then fail to start it, so clean
# it up either way.
container_started=1
docker run -d --name "$container" -p "$port:8080" \
  -e SERVER_PUBLIC_URL="$server" \
  -v kubin-demo-data:/data kubin:demo >/dev/null
curl --fail --silent --show-error --retry 45 --retry-all-errors \
  --retry-delay 1 "$server/readyz" >/dev/null

echo "Applying the crashing pod..."
# The pod can't be created until the new namespace has its default
# service account, so retry for a few seconds.
for attempt in {1..30}; do
  if kubectl apply -f "$repo_root/demo/crashing-pod.yaml" 2>"$work_dir/apply.err"; then
    break
  fi
  if (( attempt == 30 )); then
    cat "$work_dir/apply.err" >&2
    exit 1
  fi
  sleep 1
done
for _ in {1..120}; do
  restarts="$(kubectl -n kubin-demo get pod payment-api -o jsonpath='{.status.containerStatuses[0].restartCount}' 2>/dev/null || true)"
  if [[ "$restarts" =~ ^[1-9][0-9]*$ ]]; then
    break
  fi
  sleep 1
done
if [[ ! "${restarts:-}" =~ ^[1-9][0-9]*$ ]]; then
  echo "The payment-api pod did not restart. Check: kubectl -n kubin-demo describe pod payment-api" >&2
  exit 1
fi

echo "Building the kubin CLI..."
(cd "$repo_root/cli" && go build -o "$work_dir/kubin" .)
echo "Running kubin create..."
"$work_dir/kubin" create --server "$server"

finished=1
cat <<EOF

The demo is running. Open the link printed above.
To stop it, run:
  docker rm -f $container
  kind delete cluster --name $cluster
EOF
