#!/bin/sh
# Builds and publishes a release for the migrate plugin, independent of this
# repo's own .drone.yml release of the standalone kingpin binary. Triggered off
# tags matching migrate/vX.Y.Z — a prefixed series, so it never collides with
# the standalone binary's bare v* tags and never becomes GitHub's "latest".
#
# Lives under plugin/ alongside everything else the plugin owns: it builds only
# the plugin module and shares nothing with the root Makefile/.drone.yml build.
#
# Usage:
#   plugin/scripts/release-migrate.sh <tag> [--publish]
#
# <tag> must be exactly migrate/vX.Y.Z. Artifacts are always built into
# plugin/dist/ so they can be inspected. Publishing is opt-in: without --publish
# the script stops after building, so a bare/mistyped invocation can never
# create a GitHub release. --publish creates the release as a draft — core's
# install/resolve code already ignores drafts — so a human still reviews and
# publishes it from GitHub.
#
# --publish also requires HEAD to actually be the given tag, with no local
# changes on top of it — so the released binaries always match the tagged
# commit, whether run from CI or a workstation.
#
# The artifact names here are core's plugin-release convention, not ours to
# choose: `harness install plugin` looks for
# harness-plugin-migrate_<ver>_<os>_<arch>.{tar.gz,zip} plus checksums.txt in a
# release tagged migrate/vX.Y.Z, and gates the binary inside on being named
# harness-migrate. Change any of these and installs stop resolving.
set -eu

# harness/harness-migrate is where releases belong. Overridable so a fork can
# be used to rehearse the publish path without touching the real repo.
REPO="${RELEASE_REPO:-harness/harness-migrate}"
PLATFORMS="linux_amd64 linux_arm64 darwin_amd64 darwin_arm64 windows_amd64 windows_arm64"
DIST_DIR="dist"

info()  { printf '  \033[34m•\033[0m %s\n' "$*"; }
error() { printf '  \033[31m✗\033[0m %s\n' "$*" >&2; exit 1; }

TAG="${1:-}"
PUBLISH=""
case "${2:-}" in
    --publish) PUBLISH=1 ;;
    "") ;;
    *) error "unrecognized argument ${2:-} (expected --publish)" ;;
esac

[ -n "$TAG" ] || error "usage: $0 <tag> [--publish]  (e.g. $0 migrate/v0.1.0 --publish)"

echo "$TAG" | grep -Eq '^migrate/v[0-9]+\.[0-9]+\.[0-9]+$' || error "tag $TAG does not match migrate/vX.Y.Z"

# Run from plugin/ regardless of where this was invoked, so DIST_DIR and the
# build paths below don't depend on the caller's cwd. The git checks work from
# anywhere inside the repo.
cd "$(dirname "$0")/.."

if [ -n "$PUBLISH" ]; then
    tag_commit="$(git rev-list -n 1 "refs/tags/$TAG" 2>/dev/null)" || error "tag $TAG not found in this repo"
    head_commit="$(git rev-parse HEAD)"
    [ "$tag_commit" = "$head_commit" ] || error "HEAD ($head_commit) is not tag $TAG ($tag_commit) — checkout the tag before publishing"
    [ -z "$(git status --porcelain)" ] || error "working tree has local changes — publishing must build from a clean checkout of $TAG"
fi

VERSION="${TAG#migrate/}"  # v0.1.0
VER="${VERSION#v}"         # 0.1.0
BUILD_TIME="$(date -u +%Y%m%d%H%MZ)"
# hbase lives in core; the version it reports is what the host records at
# install time and compares against on upgrade. No "v" prefix — see Taskfile.yml.
LDFLAGS="-s -w -X github.com/harness/cli/v3/pkg/hbase.Version=${VER} -X github.com/harness/cli/v3/pkg/hbase.BuildTime=${BUILD_TIME}"

rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"
DIST_ABS="$(cd "$DIST_DIR" && pwd)"

info "building harness-migrate ${VERSION}"
for platform in $PLATFORMS; do
    os="${platform%_*}"
    arch="${platform#*_}"
    binary="harness-migrate"
    [ "$os" = "windows" ] && binary="harness-migrate.exe"

    stage="$(mktemp -d)"
    info "  ${platform}"
    GOOS="$os" GOARCH="$arch" CGO_ENABLED=0 \
        go build -trimpath -ldflags "$LDFLAGS" -o "${stage}/${binary}" ./cmd/harness-migrate

    archive_base="harness-plugin-migrate_${VER}_${platform}"
    if [ "$os" = "windows" ]; then
        (cd "$stage" && zip -q "${archive_base}.zip" "$binary")
        mv "${stage}/${archive_base}.zip" "$DIST_ABS/"
    else
        tar -czf "$DIST_ABS/${archive_base}.tar.gz" -C "$stage" "$binary"
    fi
    rm -rf "$stage"
done

info "writing checksums.txt"
(
    cd "$DIST_ABS"
    shasum -a 256 *.tar.gz *.zip > checksums.txt
)

info "artifacts in plugin/${DIST_DIR}:"
ls -la "$DIST_ABS"

if [ -z "$PUBLISH" ]; then
    info "built only — pass --publish to create the GitHub release"
    exit 0
fi

info "creating draft release ${TAG} on ${REPO}"
gh release create "$TAG" \
    --repo "$REPO" \
    --title "harness-migrate plugin ${VERSION}" \
    --notes "harness-migrate plugin ${VERSION}" \
    --latest=false \
    --draft \
    "$DIST_ABS"/*
