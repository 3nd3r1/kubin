# Kubin

<img src="docs/assets/logo_nobg.png" alt="Kubin logo" width="120">

Kubin turns a crashing pod into a link that shows its status, events, and logs.

<!-- Narrated walkthrough goes here: drag the mp4 into the GitHub editor so it plays inline. Script: demo/video-script.md -->

![kubin create prints a link; the page shows the crashing pod and the error in its logs](demo/kubin-demo.gif)

```sh
git clone https://github.com/3nd3r1/kubin.git
cd kubin
./demo/run.sh
```

The demo needs Docker, kind, kubectl, Go, and curl. It builds and starts Kubin, creates a local cluster, applies the failing pod, and uploads a snapshot. Open the URL printed by the script. Kubin listens on port 8080; set `KUBIN_PORT` to use another. The cluster and server stay running until you remove them with the commands printed at the end.

## What works

The demo captures namespaces, pods, pod events, and container logs. Its snapshot link opens a read-only page with the pod status, events, and logs.

## What does not

Kubin does not yet collect the other workload, networking, config, and node resources. The CLI cannot choose which resources to capture. There is no authentication, and the repository does not publish a release image yet. See [the roadmap](backlog/docs/doc-1%20-%20Direction.md) for planned work and [self-hosting](docs/self-hosting.md) for running the server yourself.

## Development

```sh
nix-shell
backlog board
```

More in [docs/development.md](docs/development.md). The system design is in [docs/architecture.md](docs/architecture.md).

## License

MIT
