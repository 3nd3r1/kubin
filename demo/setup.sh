#!/usr/bin/env bash
# Builds everything the demo needs, without taking a snapshot, and first
# deletes whatever a previous run left. That is a kind cluster running the
# broken checkout app, the kubin server in Docker, the CLI in demo/bin/kubin
# and a fresh Linear ticket (demo/linear.sh). demo/run.sh calls this and then
# takes the snapshot.
set -Eeuo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cluster="kubin-demo"
container="kubin-demo"
# The links show a made-up domain. It resolves to this machine through an
# /etc/hosts entry, so the browser still reaches the container on localhost.
host="${KUBIN_DEMO_HOST:-kubin.sh}"
port="${KUBIN_PORT:-80}"
local_server="http://localhost:$port"
if [[ "$port" == 80 ]]; then
  server="http://$host"
else
  server="http://$host:$port"
fi
apply_err="$(mktemp "${TMPDIR:-/tmp}/kubin-demo.XXXXXX")"
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
  rm -f "$apply_err"
  exit "$status"
}
trap cleanup EXIT

for tool in docker kind kubectl go curl; do
  if ! command -v "$tool" >/dev/null 2>&1; then
    echo "Missing required command: $tool" >&2
    exit 1
  fi
done

if ! getent hosts "$host" 2>/dev/null | grep -Eq '^(127\.0\.0\.1|::1)\s'; then
  echo "$host doesn't point at this machine. Run once:" >&2
  echo "  echo '127.0.0.1 $host' | sudo tee -a /etc/hosts" >&2
  exit 1
fi

# Start from scratch: remove a previous demo's cluster, container and data.
if kind get clusters 2>/dev/null | grep -Fxq "$cluster"; then
  echo "Deleting the old kind cluster '$cluster'..."
  kind delete cluster --name "$cluster"
fi
if docker container inspect "$container" >/dev/null 2>&1; then
  echo "Removing the old Docker container '$container'..."
  docker rm -f "$container" >/dev/null
fi
docker volume rm kubin-demo-data >/dev/null 2>&1 || true

if curl --silent --max-time 2 "$local_server" >/dev/null 2>&1; then
  echo "Port $port is in use. Set KUBIN_PORT to another port." >&2
  exit 1
fi

echo "Building the Kubin server image..."
docker build -t kubin:demo "$repo_root"

echo "Starting the kind cluster..."
# --wait lets the node lose its not-ready taint first. kind writes the
# kind-$cluster context into $KUBECONFIG (or ~/.kube/config) and switches to it.
kind create cluster --name "$cluster" --wait 2m
cluster_created=1
echo "Switched the current kubectl context to kind-$cluster."

echo "Starting Kubin on $server..."
# docker run can create the container and then fail to start it, so clean
# it up either way.
container_started=1
docker run -d --name "$container" -p "$port:8080" \
  -e SERVER_PUBLIC_URL="$server" \
  -v kubin-demo-data:/data kubin:demo >/dev/null
curl --fail --silent --show-error --retry 45 --retry-all-errors \
  --retry-delay 1 "$local_server/readyz" >/dev/null

echo "Applying the checkout app..."
# Resources in the new namespace can't be created until it has its default
# service account, so retry for a few seconds.
for attempt in {1..30}; do
  if kubectl apply -f "$repo_root/demo/checkout.yaml" 2>"$apply_err"; then
    break
  fi
  if (( attempt == 30 )); then
    cat "$apply_err" >&2
    exit 1
  fi
  sleep 1
done

# wait_for <what> <hint command> <check function>: polls for about 2 minutes.
wait_for() {
  local what="$1" hint="$2"
  shift 2
  for _ in {1..120}; do
    if "$@"; then
      return 0
    fi
    sleep 1
  done
  echo "Timed out waiting for: $what. Check: $hint" >&2
  exit 1
}

api_restarted() {
  local n
  n="$(kubectl -n checkout get pods -l app=api -o jsonpath='{.items[0].status.containerStatuses[0].restartCount}' 2>/dev/null || true)"
  [[ "$n" =~ ^[1-9][0-9]*$ ]]
}
postgres_unschedulable() {
  [[ -n "$(kubectl -n checkout get events --field-selector involvedObject.name=postgres-0,reason=FailedScheduling -o name 2>/dev/null)" ]]
}
worker_failed() {
  [[ "$(kubectl -n checkout get job worker -o jsonpath='{.status.conditions[?(@.type=="Failed")].status}' 2>/dev/null || true)" == "True" ]]
}
web_ready() {
  [[ "$(kubectl -n checkout get deploy web -o jsonpath='{.status.readyReplicas}' 2>/dev/null || true)" == "2" ]]
}

echo "Waiting for the problems to show up..."
wait_for "the api pod to restart" "kubectl -n checkout get pods -l app=api" api_restarted
wait_for "a FailedScheduling event on postgres-0" "kubectl -n checkout describe pod postgres-0" postgres_unschedulable
wait_for "job worker to fail" "kubectl -n checkout describe job worker" worker_failed
wait_for "deployment web to be 2/2 ready" "kubectl -n checkout get deploy web" web_ready

echo "Resetting the Linear workspace..."
"$repo_root/demo/linear.sh" || echo "Skipped the Linear reset (see above)." >&2

echo "Building the kubin CLI..."
(cd "$repo_root/cli" && go build -o ../demo/bin/kubin .)

finished=1
cat <<EOM

The demo is ready. Nothing has been snapshotted yet.
kubectl now points at the kind-$cluster context.

Next:
  export PATH="$repo_root/demo/bin:\$PATH" KUBIN_SERVER=$server
  kubin create
  k9s -n checkout

To stop it, run:
  docker rm -f $container
  kind delete cluster --name $cluster
EOM
