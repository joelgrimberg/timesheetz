# timesheetz-bin AUR package

Source of truth for the [`timesheetz-bin`](https://aur.archlinux.org/packages/timesheetz-bin)
AUR package. The package wraps the prebuilt `timesheetz_Linux_*.tar.gz`
archives published by GoReleaser on each GitHub release, so installing it
takes ~1 second and does not require a Go toolchain.

The actual AUR repo lives at:

    ssh://aur@aur.archlinux.org/timesheetz-bin.git

This directory is just where the canonical PKGBUILD is kept; the file is
copied into the AUR repo and pushed there by `scripts/release-aur.sh`.

## Updating after a new GitHub release

From the main repo root, after the GoReleaser pipeline for the new tag
has finished:

```
./scripts/release-aur.sh <version> --build --publish
```

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

Skip if your AUR account + SSH key are already wired up.

1. Register at https://aur.archlinux.org/register.
2. Under **My Account → SSH Public Key**, paste your public key (e.g.
   `cat ~/.ssh/id_ed25519.pub`).
3. Confirm auth works:
   ```
   ssh aur@aur.archlinux.org help
   ```
   You should get a non-interactive help banner, not a permission denied.

The first push to `ssh://aur@aur.archlinux.org/timesheetz-bin.git` is
what reserves the package name. Easiest first run:

```
./scripts/release-aur.sh <current-version> --build --publish
```

## Conventions worth remembering

- **`provides=('timesheetz')` + `conflicts=('timesheetz')`** mean a future
  source-build `timesheetz` package can coexist on the AUR but never be
  installed alongside `timesheetz-bin` on the same system. Keep this
  symmetric in any sibling PKGBUILD.
- The GoReleaser binary inside the tarball is named `timesheet`. We
  install it as `/usr/bin/timesheetz`, matching the Homebrew formula.
- `options=('!strip')` suppresses makepkg's redundant strip pass — the
  binary is already built with `-s -w` by GoReleaser.
