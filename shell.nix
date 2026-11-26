{
  pkgs ? import <nixpkgs> { },
}:
let
  coolifyVersion = "1.11.0"; # github.com/coollabsio/coolify-cli, not in nixpkgs
in
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

  # The pinned backlog CLI (tools/bin/backlog) and Go CLIs built once per version into .bin/ (git-ignored).
  shellHook = ''
    coolifyBin="$PWD/.bin/coolify-${coolifyVersion}"
    export PATH="$PWD/tools/bin:$PWD/demo/bin:$coolifyBin:$PATH"
    if [ ! -x "$coolifyBin/coolify" ]; then
      echo "Installing coolify-cli ${coolifyVersion}..."
      GOBIN="$coolifyBin" GOFLAGS=-modcacherw go install github.com/coollabsio/coolify-cli/coolify@v${coolifyVersion}
    fi
  '';
}
