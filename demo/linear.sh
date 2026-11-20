#!/usr/bin/env bash
# Resets the demo Linear workspace: deletes every issue in the team, then
# creates the one ticket the video starts from. Needs LINEAR_API_KEY, which
# is read from .env if it isn't already set. Deleted issues go to Linear's trash.
set -Eeuo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
team="${LINEAR_TEAM:-TES}"

if [[ -z "${LINEAR_API_KEY:-}" && -f "$repo_root/.env" ]]; then
  set -a
  # shellcheck disable=SC1091
  . "$repo_root/.env"
  set +a
fi
if [[ -z "${LINEAR_API_KEY:-}" ]]; then
  echo "LINEAR_API_KEY is not set (env or .env)." >&2
  exit 1
fi

linear="$repo_root/tools/bin/linear"

echo "Deleting every issue in Linear team $team..."
ids="$("$linear" issue list --team "$team" --all-states -A --limit all --no-pager -j | jq -r '.[].identifier')"
if [[ -n "$ids" ]]; then
  # shellcheck disable=SC2086
  "$linear" issue delete -y --bulk $ids
fi

echo "Creating the demo ticket..."
"$linear" issue create --team "$team" --no-input -p high -s Todo \
  -t "Checkout is failing in staging: api crash-looping, orders not going through" \
  -d "Since this morning's deploy, customers can't complete checkout in staging. The \`api\` pod keeps restarting, the \`postgres-0\` pod is stuck Pending, and the nightly \`worker\` job failed.

Need someone to pick this up and work out what broke. Cluster snapshot is below."
