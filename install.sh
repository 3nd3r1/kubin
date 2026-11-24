#!/bin/sh
# Install the kubin CLI from a GitHub release.
# KUBIN_VERSION       version to install, like v0.1.0 (default: latest)
# KUBIN_INSTALL_DIR   target directory (default: /usr/local/bin)
# KUBIN_DOWNLOAD_URL  base URL holding the archives and checksums.txt
set -eu

repo="3nd3r1/kubin"
dir="${KUBIN_INSTALL_DIR:-/usr/local/bin}"

fail() { echo "install.sh: $*" >&2; exit 1; }

case "$(uname -s)" in
  Linux) os=linux ;;
  Darwin) os=darwin ;;
  *) fail "unsupported OS $(uname -s), download a release by hand" ;;
esac
case "$(uname -m)" in
  x86_64 | amd64) arch=amd64 ;;
  aarch64 | arm64) arch=arm64 ;;
  *) fail "unsupported architecture $(uname -m)" ;;
esac

if [ -n "${KUBIN_DOWNLOAD_URL:-}" ]; then
  base="$KUBIN_DOWNLOAD_URL"
  version="${KUBIN_VERSION:-}"
else
  version="${KUBIN_VERSION:-}"
  if [ -z "$version" ]; then
    version=$(curl -fsSL "https://api.github.com/repos/$repo/releases/latest" |
      sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -n 1)
    [ -n "$version" ] || fail "could not find the latest release"
  fi
  base="https://github.com/$repo/releases/download/$version"
fi

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

curl -fsSL "$base/checksums.txt" -o "$tmp/checksums.txt" || fail "could not download checksums.txt from $base"

# Archive names carry the version without the leading v.
if [ -n "$version" ]; then
  archive="kubin_${version#v}_${os}_${arch}.tar.gz"
else
  archive=$(sed -n "s/^[0-9a-f]*  \(kubin_.*_${os}_${arch}\.tar\.gz\)$/\1/p" "$tmp/checksums.txt" | head -n 1)
  [ -n "$archive" ] || fail "no archive for ${os}_${arch} in checksums.txt"
fi

curl -fsSL "$base/$archive" -o "$tmp/$archive" || fail "could not download $archive"

want=$(sed -n "s/^\([0-9a-f]*\)  $archive$/\1/p" "$tmp/checksums.txt")
[ -n "$want" ] || fail "$archive is not in checksums.txt"
if command -v sha256sum >/dev/null 2>&1; then
  got=$(sha256sum "$tmp/$archive" | cut -d ' ' -f 1)
else
  got=$(shasum -a 256 "$tmp/$archive" | cut -d ' ' -f 1)
fi
[ "$want" = "$got" ] || fail "checksum mismatch for $archive"

tar -xzf "$tmp/$archive" -C "$tmp" kubin
if [ -w "$dir" ] || { [ ! -e "$dir" ] && mkdir -p "$dir" 2>/dev/null; }; then
  install -m 0755 "$tmp/kubin" "$dir/kubin"
else
  echo "$dir is not writable, using sudo" >&2
  sudo install -m 0755 "$tmp/kubin" "$dir/kubin"
fi
echo "installed kubin to $dir/kubin"
