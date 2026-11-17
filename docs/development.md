# Development

All tools come from `nix-shell` (or `direnv allow` once). That includes the pinned `backlog` CLI in `tools/bin/`.

## What works today

```bash
cd cli && make test                   # unit tests
cd server && make test                # unit tests
cd server && make run                 # server on :8080, data in server/data/
cd cli && go run . create --server http://localhost:8080   # snapshot the current kube context, print the link
cd cli && go run . create -o cluster.tar.gz                # or keep the tarball instead
cd ui && npm install && npm run dev   # placeholder page on :3000
```

## Local cluster

```bash
kind create cluster --name kubin
skaffold dev                 # builds the server image, deploys charts/kubin to kubin-dev
```

The chart is broken right now: the image path renders wrong and the probes call a subcommand the server doesn't have. TASK-7 fixes both. Until then use `go run`.

## Board

```bash
backlog board                       # kanban in the terminal
backlog browser                     # same in a browser
backlog task list -s "To Do" --plain
backlog milestone list
```

The direction is in `backlog doc view doc-1`. Decisions are in `backlog decision list`.
