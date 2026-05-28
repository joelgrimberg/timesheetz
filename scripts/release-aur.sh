#!/usr/bin/env bash
# Bump the timesheetz-bin AUR PKGBUILD to the given version, regenerate
# .SRCINFO, and optionally smoke-test (--build) and publish (--publish).
#
# Usage:
#   scripts/release-aur.sh <version> [--build] [--publish]
#
# Examples:
#   scripts/release-aur.sh 1.39.0
#   scripts/release-aur.sh v1.39.0 --build
#   scripts/release-aur.sh 1.39.0 --build --publish
#
# Set AUR_REMOTE to override the default AUR git URL (useful for testing
# against a private mirror).
set -euo pipefail

ver="${1:?usage: release-aur.sh <version> [--build] [--publish]}"
ver="${ver#v}" # strip leading 'v' if present

shift || true
do_build=0
do_publish=0
for flag in "$@"; do
    case "$flag" in
        --build)   do_build=1 ;;
        --publish) do_publish=1 ;;
        *) echo "Unknown flag: $flag" >&2; exit 2 ;;
    esac
done

repo_root="$(git rev-parse --show-toplevel)"
pkgdir="${repo_root}/packaging/aur/timesheetz-bin"
[[ -d "$pkgdir" ]] || { echo "Not a timesheetz checkout: $pkgdir missing"; exit 1; }

checksums_url="https://github.com/joelgrimberg/timesheetz/releases/download/v${ver}/timesheetz_${ver}_checksums.txt"
echo "Fetching ${checksums_url}"
checksums="$(curl -fsSL "$checksums_url")"

sha_x86=$(awk '/timesheetz_Linux_x86_64\.tar\.gz$/ {print $1}' <<<"$checksums")
sha_arm=$(awk '/timesheetz_Linux_arm64\.tar\.gz$/  {print $1}' <<<"$checksums")
if [[ -z "$sha_x86" || -z "$sha_arm" ]]; then
    echo "Could not find Linux archives in checksums.txt for v${ver}" >&2
    echo "Checksums fetched:" >&2
    echo "$checksums" >&2
    exit 1
fi

sed -i \
    -e "s/^pkgver=.*/pkgver=${ver}/" \
    -e "s/^pkgrel=.*/pkgrel=1/" \
    -e "s/^sha256sums_x86_64=.*/sha256sums_x86_64=('${sha_x86}')/" \
    -e "s/^sha256sums_aarch64=.*/sha256sums_aarch64=('${sha_arm}')/" \
    "${pkgdir}/PKGBUILD"

(cd "$pkgdir" && makepkg --printsrcinfo > .SRCINFO)
echo "Rewrote PKGBUILD and .SRCINFO to v${ver}"

if (( do_build )); then
    echo "Running makepkg -f (smoke test)…"
    (cd "$pkgdir" && makepkg -f --noconfirm)
fi

if (( do_publish )); then
    aur_remote="${AUR_REMOTE:-ssh://aur@aur.archlinux.org/timesheetz-bin.git}"
    work="$(mktemp -d)"
    trap 'rm -rf "$work"' EXIT
    echo "Cloning ${aur_remote} into ${work}…"
    git clone "$aur_remote" "$work"
    cp "${pkgdir}/PKGBUILD" "${pkgdir}/.SRCINFO" "$work/"
    (
        cd "$work"
        git add PKGBUILD .SRCINFO
        if git diff --cached --quiet; then
            echo "AUR already at v${ver}; nothing to push."
        else
            git commit -m "Update to ${ver}"
            git push
            echo "Pushed v${ver} to AUR."
        fi
    )
fi

echo "Done. Commit packaging/aur/timesheetz-bin/{PKGBUILD,.SRCINFO} to the main repo too."
