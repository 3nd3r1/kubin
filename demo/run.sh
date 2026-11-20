#!/usr/bin/env bash
set -Eeuo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
port="${KUBIN_PORT:-80}"
host="${KUBIN_DEMO_HOST:-kubin.sh}"
if [[ "$port" == 80 ]]; then
  server="http://$host"
else
  server="http://$host:$port"
fi

"$repo_root/demo/setup.sh"

echo "Running kubin create..."
status=0
"$repo_root/demo/bin/kubin" create --server "$server" || status=$?

if (( status == 0 )); then
  echo
  echo "Open the link printed above."
else
  echo "kubin create failed. The cluster and server are still running." >&2
fi
cat <<EOM
To stop the demo, run:
  docker rm -f kubin-demo
  kind delete cluster --name kubin-demo
EOM
exit "$status"
