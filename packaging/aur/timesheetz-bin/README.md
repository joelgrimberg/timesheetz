# timesheetz-bin AUR package

Source of truth for the [`timesheetz-bin`](https://aur.archlinux.org/packages/timesheetz-bin)
AUR package. The package wraps the prebuilt `timesheetz_Linux_*.tar.gz`
archives published by GoReleaser on each GitHub release, so installing it
takes ~1 second and does not require a Go toolchain.

The actual AUR repo lives at:

    ssh://aur@aur.archlinux.org/timesheetz-bin.git

## How updates happen now

On every tag push, the GitHub Actions release workflow runs GoReleaser,
which (via the `aurs:` block in `.goreleaser.yml`) generates a fresh
PKGBUILD and pushes it to the AUR remote using the dedicated `AUR_KEY`
deploy key stored as a GitHub Actions secret. There is no manual step
in the happy path.

The in-repo files in this directory (PKGBUILD, .SRCINFO) are kept as a
reviewable reference and as the source for the manual fallback below;
the PKGBUILD GoReleaser actually pushes is generated from the `aurs:`
template, not copied from here. If you change the install layout (e.g.
adding shell completions), update both the `aurs.package` block in
`.goreleaser.yml` and the local PKGBUILD so they stay in sync.

## Manual fallback (when CI is broken or for testing)

`scripts/release-aur.sh <version> [--build] [--publish]` does the same
work locally:

For example:

```
./scripts/release-aur.sh 1.39.0 --build --publish
```

What that does:

1. Downloads `checksums.txt` from the GitHub release and plucks the
   `Linux_x86_64` and `Linux_arm64` sha256 sums.
2. Rewrites `pkgver` and both `sha256sums_*` entries in this PKGBUILD,
   resets `pkgrel` to 1.
3. Regenerates `.SRCINFO` via `makepkg --printsrcinfo`.
4. `--build`: runs `makepkg -f` so you catch a broken PKGBUILD before
   AUR users do.
5. `--publish`: clones the AUR remote, copies PKGBUILD + .SRCINFO over,
   commits with `Update to <ver>`, pushes.

After it finishes:

```
git add packaging/aur/timesheetz-bin/{PKGBUILD,.SRCINFO}
git commit -m "Bump timesheetz-bin AUR package to <ver>"
git push
```

So the in-repo copy stays in sync with what AUR users actually pull.

## One-time AUR bootstrap

Already done for this repo. For reference:

1. Register at https://aur.archlinux.org/register.
2. Under **My Account → SSH Public Key**, paste both your personal
   public key (for manual fallback access) and the CI deploy key's
   public half. Multiple keys are supported, one per line.
3. Add the CI deploy key's private half as a GitHub Actions secret
   named `AUR_KEY` (used by the `aurs:` block in `.goreleaser.yml`).
4. The first push reserves the package name. After that, every tagged
   release auto-publishes via CI.

## Conventions worth remembering

- **`provides=('timesheetz')` + `conflicts=('timesheetz')`** mean a future
  source-build `timesheetz` package can coexist on the AUR but never be
  installed alongside `timesheetz-bin` on the same system. Keep this
  symmetric in any sibling PKGBUILD.
- The GoReleaser binary inside the tarball is named `timesheet`. We
  install it as `/usr/bin/timesheetz`, matching the Homebrew formula.
- `options=('!strip')` suppresses makepkg's redundant strip pass — the
  binary is already built with `-s -w` by GoReleaser.
