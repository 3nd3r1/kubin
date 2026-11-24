{
  pkgs ? import <nixpkgs> { },
}:

pkgs.mkShell {
  buildInputs = with pkgs; [
    # Go: CLI and server
    go
    golangci-lint
    gnumake

    # UI and the backlog CLI
    nodejs_22

    # Local cluster and deploys
    kubectl
    kind
    kubernetes-helm
    skaffold
    k9s

    # Releases
    goreleaser
  ];

  # The pinned backlog CLI (tools/bin/backlog).
  shellHook = ''
    export PATH="$PWD/tools/bin:$PWD/demo/bin:$PATH"
  '';
}
